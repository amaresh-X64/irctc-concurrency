"""
baseline.py

Runs significant_terms aggregations on confirmed BookSeat spans to build
a statistical profile of what "normal" looks like.

How it works:
- The foreground set = all BookSeat spans where outcome = "confirmed"
- The background set = ALL spans in irctc-traces (the full corpus)
- significant_terms + JLH scoring finds terms that appear
  disproportionately more in confirmed bookings vs the full corpus
- These become the "normal fingerprint" the analyzer scores against

JLH formula: (fg_percent - bg_percent)^2 / bg_percent
Penalizes both very rare terms (noise) and very common terms (uninformative).
The sweet spot is terms that are strongly characteristic of confirmed bookings.
"""

from .es_client import get_es, TRACES_INDEX


def build_baseline() -> dict:
    es = get_es()

    resp = es.search(
        index=TRACES_INDEX,
        body={
            "size": 0,

            # Foreground: only confirmed BookSeat spans from gin-booking
            "query": {
                "bool": {
                    "must": [
                        {"term": {"Name.keyword": "BookSeat"}},
                        {"term": {"Attributes.booking.outcome.keyword": "confirmed"}}
                    ]
                }
            },

            "aggs": {

                # ── significant_terms on booking outcome ──────────────────
                # Even though foreground is all "confirmed", this anchors
                # the JLH score — tells us how over-represented "confirmed"
                # is vs the full corpus (which includes contention spans)
                "sig_outcomes": {
                    "significant_terms": {
                        "field": "Attributes.booking.outcome.keyword",
                        "jlh": {}
                    }
                },

                # ── significant_terms on seat type ────────────────────────
                # Which seat types (FIRST_AC, SECOND_AC etc.) are
                # characteristic of successful bookings?
                "sig_seat_types": {
                    "significant_terms": {
                        "field": "Attributes.booking.seat_type.keyword",
                        "jlh": {}
                    }
                },

                # ── significant_terms on span name ────────────────────────
                # Which operation names are normal in a confirmed flow?
                # e.g. BookSeat + SeatLocker.TryLock + BookSeat.DBTransaction
                "sig_span_names": {
                    "significant_terms": {
                        "field": "Name.keyword",
                        "jlh": {}
                    }
                },

                # ── significant_terms on service name ─────────────────────
                "sig_services": {
                    "significant_terms": {
                        "field": "Resource.service.name.keyword",
                        "jlh": {}
                    }
                },

                # ── stats on Duration for confirmed spans ─────────────────
                # Used to flag duration anomalies (e.g. DB slowdown)
                "duration_stats": {
                    "extended_stats": {
                        "field": "Duration"
                    }
                },

                # ── locker hit rate in confirmed bookings ─────────────────
                "locker_hit_rate": {
                    "terms": {
                        "field": "Attributes.booking.locker_hit"
                    }
                },

                # ── db seat available rate ────────────────────────────────
                "db_available_rate": {
                    "terms": {
                        "field": "Attributes.db.seat_available"
                    }
                }
            }
        }
    )

    aggs = resp["aggregations"]
    total_confirmed = resp["hits"]["total"]["value"]

    # ── Duration thresholds ───────────────────────────────────────────────
    # A span is "slow" if it exceeds mean + 2 standard deviations
    dur = aggs["duration_stats"]
    duration_mean = dur.get("avg") or 0
    duration_std = dur.get("std_deviation") or 0
    duration_slow_threshold = duration_mean + (2 * duration_std)

    # ── Extract JLH-scored term lists ─────────────────────────────────────
    def extract_sig_terms(agg_key: str) -> list[dict]:
        return [
            {
                "term": b["key"],
                "score": round(b["score"], 4),
                "doc_count": b["doc_count"],
                "bg_count": b["bg_count"]
            }
            for b in aggs[agg_key].get("buckets", [])
        ]

    baseline = {
        "total_confirmed_spans": total_confirmed,
        "significant_terms": {
            "outcomes":    extract_sig_terms("sig_outcomes"),
            "seat_types":  extract_sig_terms("sig_seat_types"),
            "span_names":  extract_sig_terms("sig_span_names"),
            "services":    extract_sig_terms("sig_services"),
        },
        "duration": {
            "mean_us":           round(duration_mean, 2),
            "std_dev_us":        round(duration_std, 2),
            "slow_threshold_us": round(duration_slow_threshold, 2),
            "min_us":            dur.get("min"),
            "max_us":            dur.get("max"),
        },
        "locker_hit_rate": {
            b["key_as_string"]: b["doc_count"]
            for b in aggs["locker_hit_rate"].get("buckets", [])
        },
        "db_available_rate": {
            b["key_as_string"]: b["doc_count"]
            for b in aggs["db_available_rate"].get("buckets", [])
        }
    }

    return baseline