"""
eql.py

EQL (Event Query Language) is built for detecting SEQUENCES of events
across time — a different capability than significant_terms (statistical)
or percolator (single-document matching).

Here we detect the exact race-condition signature in your two-layer lock:

  1. SeatLocker.TryLock  fires with  Attributes.locker.acquired == true
     (in-memory lock granted)
  2. BookSeat.DBTransaction fires with Attributes.db.seat_available == false
     (but the DB-level check rejects it anyway)

Both events share the same TraceId (one booking attempt) and happen
within a tight time window. EQL's `sequence by ... with maxspan` is
purpose-built for this — a regular query can't express "A then B
within the same trace, in this order, within N seconds".
"""

from .es_client import get_es, TRACES_INDEX

RACE_CONDITION_QUERY = """
sequence by TraceId with maxspan=5s
  [any where Name.keyword == "SeatLocker.TryLock" and Attributes.locker.acquired == true]
  [any where Name.keyword == "BookSeat.DBTransaction" and Attributes.db.seat_available == false]
"""


def find_race_conditions(size: int = 20) -> dict:
    """
    Runs the EQL sequence query and returns every trace where the
    in-memory locker succeeded but the DB-level check rejected the booking.

    This is the second-layer race condition: requests that get past the
    Go SeatLocker mutex but collide at the Postgres transaction level.
    """
    es = get_es()

    resp = es.eql.search(
        index=TRACES_INDEX,
        body={
            "query": RACE_CONDITION_QUERY,
            "size": size,
        }
    )

    sequences = resp.get("hits", {}).get("sequences", [])

    results = []
    for seq in sequences:
        events = seq.get("events", [])
        trace_id = seq.get("join_keys", [None])[0]

        lock_event = events[0]["_source"] if len(events) > 0 else {}
        db_event = events[1]["_source"] if len(events) > 1 else {}

        results.append({
            "trace_id": trace_id,
            "lock_acquired_at": lock_event.get("@timestamp"),
            "db_rejected_at": db_event.get("@timestamp"),
            "seat_id": lock_event.get("Attributes", {})
                                  .get("locker", {})
                                  .get("key"),
            "lock_span_id": lock_event.get("SpanId"),
            "db_span_id": db_event.get("SpanId"),
            "db_duration_us": db_event.get("Duration"),
        })

    return {
        "pattern": "lock_acquired_then_db_rejected",
        "description": (
            "In-memory SeatLocker granted the lock, but the Postgres "
            "transaction-level check found the seat already taken — "
            "the second-layer race condition."
        ),
        "total_matches": len(results),
        "matches": results,
    }