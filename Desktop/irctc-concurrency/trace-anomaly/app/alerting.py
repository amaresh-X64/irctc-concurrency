"""
alerting.py

Background poller that turns the percolator from "pull" (manual /percolator/match
calls) into "push" (automatic alerting).

Every POLL_INTERVAL_SECONDS:
  1. Query irctc-traces for BookSeat spans with @timestamp > last_checked
  2. Run each span through the percolator
  3. For any span where rules fire, write a document to irctc-alerts
  4. Advance last_checked to the newest span's timestamp seen

GET /alerts then just reads from irctc-alerts — no percolation happens
on the read path, keeping it fast.
"""

import asyncio
import logging
from datetime import datetime, timezone

from .es_client import get_es, TRACES_INDEX
from .percolator import match_span_against_rules, ensure_percolator_index

logger = logging.getLogger(__name__)

ALERTS_INDEX = "irctc-alerts"
POLL_INTERVAL_SECONDS = 10

ALERTS_INDEX_MAPPING = {
    "mappings": {
        "properties": {
            "alert_time":    {"type": "date"},
            "trace_id":      {"type": "keyword"},
            "span_id":       {"type": "keyword"},
            "span_name":     {"type": "keyword"},
            "service":       {"type": "keyword"},
            "outcome":       {"type": "keyword"},
            "rule_id":       {"type": "keyword"},
            "rule_name":     {"type": "keyword"},
            "severity":      {"type": "keyword"},
            "description":   {"type": "text"},
            "span_timestamp": {"type": "date"},
        }
    }
}

# Module-level watermark — tracks the newest span @timestamp we've already
# checked, so each poll only looks at spans newer than this.
_last_checked: str | None = None


def _ensure_alerts_index():
    es = get_es()
    if not es.indices.exists(index=ALERTS_INDEX):
        es.indices.create(index=ALERTS_INDEX, body=ALERTS_INDEX_MAPPING)
        logger.info("Created %s index", ALERTS_INDEX)


def _poll_once():
    global _last_checked

    es = get_es()
    ensure_percolator_index()
    _ensure_alerts_index()

    # On first run, only look back 5 minutes so we don't flood on startup
    # with every historical span.
    if _last_checked is None:
        _last_checked = "now-5m"

    query = {
        "bool": {
            "must": [
                {"term": {"Name.keyword": "BookSeat"}},
                {"range": {"@timestamp": {"gt": _last_checked}}}
            ]
        }
    }

    resp = es.search(
        index=TRACES_INDEX,
        body={
            "query": query,
            "sort": [{"@timestamp": {"order": "asc"}}],
            "size": 100,
        }
    )

    hits = resp["hits"]["hits"]
    if not hits:
        return 0

    new_alerts = 0
    newest_ts = _last_checked

    for hit in hits:
        span = hit["_source"]
        ts = span.get("@timestamp")
        if ts:
            newest_ts = ts

        try:
            fired = match_span_against_rules(span)
        except Exception as e:
            logger.error("Percolator match failed for span %s: %s", hit["_id"], e)
            continue

        if not fired:
            continue

        attrs = span.get("Attributes", {})
        booking = attrs.get("booking", {})

        for rule in fired:
            doc = {
                "alert_time": datetime.now(timezone.utc).isoformat(),
                "trace_id": span.get("TraceId"),
                "span_id": span.get("SpanId"),
                "span_name": span.get("Name"),
                "service": span.get("Resource", {}).get("service", {}).get("name"),
                "outcome": booking.get("outcome"),
                "rule_id": rule["rule_id"],
                "rule_name": rule["rule_name"],
                "severity": rule["severity"],
                "description": rule["description"],
                "span_timestamp": ts,
            }
            es.index(index=ALERTS_INDEX, document=doc)
            new_alerts += 1

    if newest_ts and newest_ts != "now-5m":
        _last_checked = newest_ts

    return new_alerts


async def alert_poller_loop():
    """
    Long-running background task. Started on FastAPI startup,
    runs forever until the app shuts down.
    """
    logger.info("Alert poller started — polling every %ds", POLL_INTERVAL_SECONDS)
    while True:
        try:
            count = _poll_once()
            if count > 0:
                logger.info("Alert poller: %d new alert(s) written", count)
        except Exception as e:
            logger.error("Alert poller error: %s", e)

        await asyncio.sleep(POLL_INTERVAL_SECONDS)


def get_recent_alerts(size: int = 50, severity: str | None = None) -> dict:
    es = get_es()
    _ensure_alerts_index()

    query: dict = {"match_all": {}}
    if severity:
        query = {"term": {"severity": severity.upper()}}

    resp = es.search(
        index=ALERTS_INDEX,
        body={
            "query": query,
            "sort": [{"alert_time": {"order": "desc"}}],
            "size": size,
        }
    )

    alerts = [hit["_source"] for hit in resp["hits"]["hits"]]
    return {
        "total": resp["hits"]["total"]["value"],
        "returned": len(alerts),
        "alerts": alerts,
    }