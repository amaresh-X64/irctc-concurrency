"""
percolator.py

The percolator flips normal ES querying:
  Normal:     store documents, query against them
  Percolator: store QUERIES (rules), match new documents against them

Here we register alerting rules for contention spikes.
When a new span comes in, we run it through the percolator index —
any matching rule fires an alert.
"""

from .es_client import get_es, PERCOLATOR_INDEX

RULES_INDEX_MAPPING = {
    "mappings": {
        "properties": {
            # The percolator field stores the query definition
            "query": {"type": "percolator"},
            "rule_name": {"type": "keyword"},
            "severity": {"type": "keyword"},
            "description": {"type": "text"},
            # Mirror the fields from irctc-traces that rules will match on
            "Attributes.booking.outcome": {"type": "keyword"},
            "Attributes.locker.acquired": {"type": "boolean"},
            "Attributes.db.seat_available": {"type": "boolean"},
            "TraceStatus": {"type": "long"},
            "Name": {"type": "keyword"},
        }
    }
}


def ensure_percolator_index():
    es = get_es()
    if not es.indices.exists(index=PERCOLATOR_INDEX):
        es.indices.create(index=PERCOLATOR_INDEX, body=RULES_INDEX_MAPPING)


def register_default_rules():
    """
    Seed the percolator index with the default alert rules.
    These fire when a new span matches the condition.
    """
    ensure_percolator_index()
    es = get_es()

    rules = [
        {
            "id": "rule-lock-contention",
            "rule_name": "seat_lock_contention",
            "severity": "WARNING",
            "description": "Seat locker was not acquired — concurrent booking collision detected",
            "query": {
                "term": {"Attributes.locker.acquired": False}
            }
        },
        {
            "id": "rule-db-contention",
            "rule_name": "db_seat_unavailable",
            "severity": "CRITICAL",
            "description": "Seat unavailable at DB level despite passing locker — possible race condition",
            "query": {
                "term": {"Attributes.db.seat_available": False}
            }
        },
        {
            "id": "rule-booking-failed",
            "rule_name": "booking_outcome_anomaly",
            "severity": "CRITICAL",
            "description": "BookSeat span completed with non-confirmed outcome",
            "query": {
                "bool": {
                    "must": [
                        {"term": {"Name": "BookSeat"}},
                        {"term": {"TraceStatus": 2}}  # OTel ERROR status
                    ]
                }
            }
        },
        {
            "id": "rule-contention-spike",
            "rule_name": "lock_contention_outcome",
            "severity": "CRITICAL",
            "description": "Booking resulted in lock_contention outcome — seat taken by another user",
            "query": {
                "term": {"Attributes.booking.outcome": "lock_contention"}
            }
        },
    ]

    registered = []
    for rule in rules:
        rule_id = rule.pop("id")
        es.index(index=PERCOLATOR_INDEX, id=rule_id, document=rule)
        registered.append(rule_id)

    return {"registered": registered}


def match_span_against_rules(span: dict) -> list[dict]:
    """
    Run a span document through the percolator — returns all rules that fire.
    """
    ensure_percolator_index()
    es = get_es()

    # Flatten span for percolator matching
    flat = {
        "Name": span.get("Name", ""),
        "TraceStatus": span.get("TraceStatus", 0),
        "Attributes.booking.outcome": span.get("Attributes", {}).get("booking", {}).get("outcome", ""),
        "Attributes.locker.acquired": span.get("Attributes", {}).get("locker", {}).get("acquired"),
        "Attributes.db.seat_available": span.get("Attributes", {}).get("db", {}).get("seat_available"),
    }

    resp = es.search(
        index=PERCOLATOR_INDEX,
        body={
            "query": {
                "percolate": {
                    "field": "query",
                    "document": flat
                }
            }
        }
    )

    return [
        {
            "rule_id": hit["_id"],
            "rule_name": hit["_source"]["rule_name"],
            "severity": hit["_source"]["severity"],
            "description": hit["_source"]["description"],
        }
        for hit in resp["hits"]["hits"]
    ]