from sqlalchemy.orm import Session
from app.trains.repository import TrainRepository
from app.constants.constants import MSG_TRAINS_FOUND, MSG_NO_TRAINS, MSG_SEATS_FOUND
from app.helpers.response import success_response, error_response
from app.search import elastic as es_client
from app.telemetry import get_tracer
from datetime import date, datetime
from opentelemetry.trace import Status, StatusCode
import logging
import pytz

logger = logging.getLogger(__name__)
tracer = get_tracer("fastapi-trains/trains")


class TrainService:

    def __init__(self, db: Session):
        self.repo = TrainRepository(db)

    def _validate_date(self, journey_date: str):
        try:
            jd = date.fromisoformat(journey_date[:10])
        except ValueError:
            return None
        return jd

    @staticmethod
    def bootstrap_es_index(db: Session) -> None:
        client = es_client.get_es_client()
        if client is None:
            logger.warning("ES unavailable at startup — skipping index bootstrap")
            return
        try:
            es_client.ensure_index(client)
            repo = TrainRepository(db)
            trains = repo.find_all_by_date(str(date.today()))
            es_client.bulk_index_trains(client, trains)
        except Exception as exc:
            logger.error("ES bootstrap failed: %s", exc)

    def get_all_trains(self, journey_date: str):
        jd = self._validate_date(journey_date)
        if jd is None or jd < date.today():
            return error_response("Invalid or past journey date", None)

        trains = self._get_all_trains_with_live_seats(journey_date)
        trains = self._filter_departed(trains, journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains)

    def search_trains(self, source: str, destination: str, journey_date: str):
        jd = self._validate_date(journey_date)
        if jd is None or jd < date.today():
            return error_response("Invalid or past journey date", None)

        # ── Manual span: captures ES vs Postgres path and result count ───────
        with tracer.start_as_current_span("TrainService.search_trains") as span:
            span.set_attribute("search.source", source)
            span.set_attribute("search.destination", destination)
            span.set_attribute("search.journey_date", journey_date)

            trains = self._search_with_live_seats(source, destination, journey_date)
            trains = self._filter_departed(trains, journey_date)

            span.set_attribute("search.result_count", len(trains))
            span.set_attribute("search.found", len(trains) > 0)

            if len(trains) == 0:
                span.set_status(Status(StatusCode.OK))
                span.set_attribute("search.outcome", "no_results")
            else:
                span.set_status(Status(StatusCode.OK))
                span.set_attribute("search.outcome", "results_found")

            msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
            return success_response(msg, trains)

    def get_seats(self, train_id: int, journey_date: str):
        train = self.repo.find_by_id(train_id)
        if not train:
            return error_response("Train not found", None)
        seats = self.repo.find_seats_by_date(train_id, journey_date)
        return success_response(MSG_SEATS_FOUND, seats)

    def _get_all_trains_with_live_seats(self, journey_date: str) -> list[dict]:
        client = es_client.get_es_client()
        if client is not None:
            try:
                es_trains = es_client.search_all_trains(client)
                if es_trains:
                    return self._patch_available_seats(es_trains, journey_date)
            except Exception as exc:
                logger.warning("ES get_all failed (%s) — falling back to Postgres", exc)

        return self.repo.find_all_by_date(journey_date)

    def _search_with_live_seats(
        self, source: str, destination: str, journey_date: str
    ) -> list[dict]:
        # ── Span: records whether ES or Postgres served this search ──────────
        with tracer.start_as_current_span("TrainService._search_backend") as span:
            client = es_client.get_es_client()
            if client is not None:
                try:
                    es_trains = es_client.search_trains(client, source, destination)
                    span.set_attribute("search.backend", "elasticsearch")
                    span.set_attribute("search.es_hit", True)
                    if es_trains:
                        return self._patch_available_seats(es_trains, journey_date)
                    return []
                except Exception as exc:
                    logger.warning("ES search failed (%s) — falling back to Postgres", exc)
                    span.set_attribute("search.backend", "postgres_fallback")
                    span.set_attribute("search.es_hit", False)
                    span.record_exception(exc)
            else:
                span.set_attribute("search.backend", "postgres_fallback")
                span.set_attribute("search.es_hit", False)

            return self.repo.find_trains_by_date(source, destination, journey_date)

    def _patch_available_seats(
        self, es_trains: list[dict], journey_date: str
    ) -> list[dict]:
        live_trains = self.repo.find_all_by_date(journey_date)
        live_seats_map = {t["id"]: t["available_seats"] for t in live_trains}

        patched = []
        for t in es_trains:
            train = dict(t)
            if train["id"] in live_seats_map:
                train["available_seats"] = live_seats_map[train["id"]]
            patched.append(train)
        return patched

    def _filter_departed(self, trains: list[dict], journey_date: str) -> list[dict]:
        try:
            jd = date.fromisoformat(journey_date[:10])
        except ValueError:
            return trains
        if jd != date.today():
            return trains
        ist = pytz.timezone("Asia/Kolkata")
        now = datetime.now(ist).time()
        result = []
        for t in trains:
            try:
                dep = datetime.strptime(str(t["departure_time"])[:8], "%H:%M:%S").time()
                if dep > now:
                    result.append(t)
            except Exception:
                result.append(t)
        return result
