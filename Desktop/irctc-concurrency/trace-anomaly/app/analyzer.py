"""
analyzer.py

Scores a new BookSeat span against the baseline fingerprint.

Anomaly signals:
1. outcome is not in the normal significant_terms set         → OUTCOME_ANOMALY
2. locker was not acquired (contention)                       → LOCK_CONTENTION
3. db reported seat not available                             → DB_CONTENTION
4. Duration exceeds mean + 2 std_dev                         → SLOW_SPAN
5. TraceStatus != 1 (OTel error status)                      → TRACE_ERROR
6. seat_type not seen in normal significant_terms             → UNUSUAL_SEAT_TYPE

Each flag has a severity: WARNING or CRITICAL.
"""

from .baseline import build_baseline


def analyze_trace(span: dict) -> dict:
    baseline = build_baseline()
    flags = []

    attrs = span.get("Attributes", {})
    booking = attrs.get("booking", {})
    locker = attrs.get("locker", {})
    db = attrs.get("db", {})

    outcome = booking.get("outcome", "")
    locker_acquired = locker.get("acquired", booking.get("locker_hit", None))
    db_available = db.get("seat_available", None)
    duration = span.get("Duration", 0)
    trace_status = span.get("TraceStatus", 0)
    seat_type = booking.get("seat_type", "")

    # ── Normal outcome terms from baseline ───────────────────────────────
    normal_outcomes = {
        t["term"] for t in baseline["significant_terms"]["outcomes"]
    }
    normal_seat_types = {
        t["term"] for t in baseline["significant_terms"]["seat_types"]
    }

    # ── Flag 1: outcome anomaly ───────────────────────────────────────────
    if outcome and outcome not in normal_outcomes:
        flags.append({
            "type": "OUTCOME_ANOMALY",
            "severity": "CRITICAL",
            "detail": f"outcome '{outcome}' not in normal baseline {normal_outcomes}",
        })

    # ── Flag 2: lock contention ───────────────────────────────────────────
    if locker_acquired is False:
        flags.append({
            "type": "LOCK_CONTENTION",
            "severity": "WARNING",
            "detail": "in-memory seat locker was not acquired — concurrent booking collision",
        })

    # ── Flag 3: DB contention ─────────────────────────────────────────────
    if db_available is False:
        flags.append({
            "type": "DB_CONTENTION",
            "severity": "CRITICAL",
            "detail": "seat was unavailable at DB level despite locker — possible race condition",
        })

    # ── Flag 4: slow span ─────────────────────────────────────────────────
    slow_threshold = baseline["duration"]["slow_threshold_us"]
    if slow_threshold > 0 and duration > slow_threshold:
        flags.append({
            "type": "SLOW_SPAN",
            "severity": "WARNING",
            "detail": (
                f"Duration {duration}µs exceeds slow threshold "
                f"{round(slow_threshold)}µs "
                f"(mean={round(baseline['duration']['mean_us'])}µs, "
                f"std={round(baseline['duration']['std_dev_us'])}µs)"
            ),
        })

    # ── Flag 5: trace error status ────────────────────────────────────────
    if trace_status != 0 and trace_status != 1:
        flags.append({
            "type": "TRACE_ERROR",
            "severity": "CRITICAL",
            "detail": f"TraceStatus={trace_status} indicates an OTel error status",
        })

    # ── Flag 6: unusual seat type ─────────────────────────────────────────
    if seat_type and normal_seat_types and seat_type not in normal_seat_types:
        flags.append({
            "type": "UNUSUAL_SEAT_TYPE",
            "severity": "WARNING",
            "detail": f"seat_type '{seat_type}' not seen in normal confirmed bookings",
        })

    # ── Overall verdict ───────────────────────────────────────────────────
    critical_count = sum(1 for f in flags if f["severity"] == "CRITICAL")
    warning_count = sum(1 for f in flags if f["severity"] == "WARNING")

    if critical_count > 0:
        verdict = "ANOMALOUS"
    elif warning_count > 0:
        verdict = "SUSPICIOUS"
    else:
        verdict = "NORMAL"

    return {
        "verdict": verdict,
        "flag_count": len(flags),
        "critical": critical_count,
        "warnings": warning_count,
        "flags": flags,
        "span_summary": {
            "name": span.get("Name"),
            "outcome": outcome,
            "locker_acquired": locker_acquired,
            "duration_us": duration,
            "trace_status": trace_status,
            "service": span.get("Resource", {}).get("service", {}).get("name"),
        },
        "baseline_used": {
            "total_confirmed_spans": baseline["total_confirmed_spans"],
            "slow_threshold_us": round(slow_threshold, 2),
            "normal_outcomes": list(normal_outcomes),
        }
    }


def analyze_by_trace_id(trace_id: str) -> dict:
    """
    Fetch all spans for a trace_id from ES and analyze the BookSeat span.
    """
    from .es_client import get_es, TRACES_INDEX

    es = get_es()
    resp = es.search(
        index=TRACES_INDEX,
        body={
            "query": {"term": {"TraceId.keyword": trace_id}},
            "size": 50
        }
    )

    spans = [hit["_source"] for hit in resp["hits"]["hits"]]
    if not spans:
        return {"error": f"No spans found for trace_id={trace_id}"}

    # Find the BookSeat span — that's the one we analyze
    book_span = next(
        (s for s in spans if s.get("Name") == "BookSeat"),
        spans[0]  # fallback to first span
    )

    result = analyze_trace(book_span)
    result["trace_id"] = trace_id
    result["total_spans_in_trace"] = len(spans)
    result["all_span_names"] = [s.get("Name") for s in spans]
    return result