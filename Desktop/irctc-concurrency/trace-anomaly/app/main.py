"""
Trace Anomaly Memory Service

Endpoints:
  GET  /health                        — liveness check
  GET  /baseline                      — run significant_terms, return normal fingerprint
  POST /analyze                       — score a raw span dict against baseline
  GET  /analyze/{trace_id}            — fetch trace from ES by ID and analyze it
  GET  /analyze/latest                — analyze the most recent BookSeat span
  POST /percolator/rules              — seed default alert rules
  POST /percolator/match              — match a span against all percolator rules
  GET  /eql/race-conditions           — find lock-acquired-then-db-rejected sequences
  GET  /contention/summary            — aggregated view of contention in last 1 hour
  GET  /alerts                        — view automatically-generated alerts
"""

import asyncio
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Any

from .baseline import build_baseline
from .analyzer import analyze_trace, analyze_by_trace_id
from .percolator import register_default_rules, match_span_against_rules
from .eql import find_race_conditions
from .alerting import alert_poller_loop, get_recent_alerts
from .es_client import get_es, TRACES_INDEX

app = FastAPI(
    title="IRCTC Trace Anomaly Memory",
    description="Detects anomalies in IRCTC booking traces using Elasticsearch significant_terms + JLH scoring",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.on_event("startup")
async def startup():
    # Seed percolator rules on startup
    try:
        register_default_rules()
    except Exception as e:
        print(f"Warning: could not seed percolator rules: {e}")

    # Start the background alert poller — turns percolation into
    # automatic, push-based alerting instead of manual /percolator/match calls.
    asyncio.create_task(alert_poller_loop())


@app.get("/health")
def health():
    return {"status": "ok", "service": "trace-anomaly"}


# ── Baseline ──────────────────────────────────────────────────────────────────

@app.get("/baseline")
def get_baseline():
    """
    Runs significant_terms + JLH on all confirmed BookSeat spans.
    Returns the statistical fingerprint of what normal looks like.
    This is the core Elasticsearch showcase — JLH scoring is unique to ES.
    """
    try:
        return build_baseline()
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ── Analyzer ──────────────────────────────────────────────────────────────────

class SpanPayload(BaseModel):
    span: dict[str, Any]


@app.post("/analyze")
def analyze(payload: SpanPayload):
    """Score a raw span document against the baseline."""
    try:
        return analyze_trace(payload.span)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/analyze/latest")
def analyze_latest():
    """Fetch and analyze the most recent BookSeat span from ES."""
    es = get_es()
    resp = es.search(
        index=TRACES_INDEX,
        body={
            "query": {"term": {"Name.keyword": "BookSeat"}},
            "sort": [{"@timestamp": {"order": "desc"}}],
            "size": 1
        }
    )
    hits = resp["hits"]["hits"]
    if not hits:
        raise HTTPException(status_code=404, detail="No BookSeat spans found")
    return analyze_trace(hits[0]["_source"])


@app.get("/analyze/{trace_id}")
def analyze_by_id(trace_id: str):
    """Fetch all spans for a trace ID and analyze the BookSeat span."""
    try:
        result = analyze_by_trace_id(trace_id)
        if "error" in result:
            raise HTTPException(status_code=404, detail=result["error"])
        return result
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ── Percolator ────────────────────────────────────────────────────────────────

@app.post("/percolator/rules")
def seed_rules():
    """Register default alert rules in the percolator index."""
    try:
        return register_default_rules()
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/percolator/match")
def percolator_match(payload: SpanPayload):
    """Match a span against all percolator rules. Returns fired rules."""
    try:
        fired = match_span_against_rules(payload.span)
        return {
            "fired_rules": fired,
            "rule_count": len(fired),
            "alerted": len(fired) > 0
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ── EQL ────────────────────────────────────────────────────────────────────────

@app.get("/eql/race-conditions")
def eql_race_conditions(size: int = 20):
    """
    Runs an EQL sequence query to find every trace where the in-memory
    SeatLocker granted the lock (Attributes.locker.acquired == true) but
    the Postgres transaction-level check rejected it
    (Attributes.db.seat_available == false) within 5 seconds.

    This is the second-layer race condition — EQL's `sequence` syntax
    expresses "A then B, same trace, within N seconds" in a way a normal
    query cannot.
    """
    try:
        return find_race_conditions(size=size)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ── Contention Summary ────────────────────────────────────────────────────────

@app.get("/contention/summary")
def contention_summary():
    """
    Aggregated view of booking contention over the last hour.
    Shows the ratio of confirmed vs contention outcomes — 
    a contention spike is the primary anomaly this system detects.
    """
    es = get_es()
    resp = es.search(
        index=TRACES_INDEX,
        body={
            "size": 0,
            "query": {
                "bool": {
                    "must": [
                        {"term": {"Name.keyword": "BookSeat"}},
                        {"range": {"@timestamp": {"gte": "now-1h"}}}
                    ]
                }
            },
            "aggs": {
                "outcomes": {
                    "terms": {"field": "Attributes.booking.outcome.keyword"}
                },
                "locker_results": {
                    "terms": {"field": "Attributes.booking.locker_hit"}
                },
                "duration_percentiles": {
                    "percentiles": {
                        "field": "Duration",
                        "percents": [50, 90, 95, 99]
                    }
                },
                "over_time": {
                    "date_histogram": {
                        "field": "@timestamp",
                        "fixed_interval": "5m"
                    },
                    "aggs": {
                        "outcomes": {
                            "terms": {"field": "Attributes.booking.outcome.keyword"}
                        }
                    }
                }
            }
        }
    )

    aggs = resp["aggregations"]
    total = resp["hits"]["total"]["value"]

    outcome_buckets = aggs["outcomes"]["buckets"]
    confirmed = next((b["doc_count"] for b in outcome_buckets if b["key"] == "confirmed"), 0)

    # Two distinct contention signals:
    #  - lock_contention   : in-memory locker rejected the request outright
    #  - seat_unavailable  : locker was acquired but DB-level check rejected it
    #                        (the second-layer race condition catch)
    lock_contention = next((b["doc_count"] for b in outcome_buckets if b["key"] == "lock_contention"), 0)
    seat_unavailable = next((b["doc_count"] for b in outcome_buckets if b["key"] == "seat_unavailable"), 0)
    total_contention = lock_contention + seat_unavailable

    contention_rate = round(total_contention / total, 4) if total > 0 else 0
    anomaly = contention_rate > 0.3  # >30% contention rate = anomalous

    return {
        "window": "last_1_hour",
        "total_book_attempts": total,
        "confirmed": confirmed,
        "contention": {
            "total": total_contention,
            "lock_contention": lock_contention,
            "seat_unavailable": seat_unavailable,
        },
        "contention_rate": contention_rate,
        "anomaly_detected": anomaly,
        "anomaly_reason": "contention_rate > 30%" if anomaly else None,
        "duration_percentiles_us": aggs["duration_percentiles"]["values"],
        "outcomes_breakdown": outcome_buckets,
        "over_time": [
            {
                "timestamp": b["key_as_string"],
                "total": b["doc_count"],
                "outcomes": b["outcomes"]["buckets"]
            }
            for b in aggs["over_time"]["buckets"]
        ]
    }


# ── Alerts ───────────────────────────────────────────────────────────────────

@app.get("/alerts")
def alerts(size: int = 50, severity: str | None = None):
    """
    View automatically-generated alerts.

    A background poller runs every 10 seconds, percolates new BookSeat
    spans against the rules in irctc-anomaly-rules, and writes any matches
    here to irctc-alerts. This is the "push" side — no manual
    /percolator/match calls needed.

    Optional ?severity=CRITICAL or ?severity=WARNING to filter.
    """
    try:
        return get_recent_alerts(size=size, severity=severity)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))