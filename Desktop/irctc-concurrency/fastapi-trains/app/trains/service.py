from sqlalchemy.orm import Session
from app.trains.repository import TrainRepository
from app.constants.constants import MSG_TRAINS_FOUND, MSG_NO_TRAINS, MSG_SEATS_FOUND
from app.helpers.response import success_response, error_response
from app.search import elastic as es_client
from datetime import date, datetime
import logging
import pytz

logger = logging.getLogger(__name__)


class TrainService:

    def __init__(self, db: Session):
        self.repo = TrainRepository(db)

    def _validate_date(self, journey_date: str):
        try:
            jd = date.fromisoformat(journey_date[:10])
        except ValueError:
            return None
        return jd

    # ── Startup indexing ────────────────────────────────────────────────────
    @staticmethod
    def bootstrap_es_index(db: Session) -> None:
        """
        Called once at app startup. Fetches all trains from Postgres and
        bulk-indexes them into Elasticsearch. Safe to call repeatedly —
        it is a no-op if ES is unreachable.
        """
        client = es_client.get_es_client()
        if client is None:
            logger.warning("ES unavailable at startup — skipping index bootstrap")
            return
        try:
            es_client.ensure_index(client)
            repo = TrainRepository(db)
            # Seed with today's date just to get the full train list;
            # available_seats here is a point-in-time snapshot.
            trains = repo.find_all_by_date(str(date.today()))
            es_client.bulk_index_trains(client, trains)
        except Exception as exc:
            logger.error("ES bootstrap failed: %s", exc)

    # ── Public API ──────────────────────────────────────────────────────────
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

        trains = self._search_with_live_seats(source, destination, journey_date)
        trains = self._filter_departed(trains, journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains)

    def get_seats(self, train_id: int, journey_date: str):
        train = self.repo.find_by_id(train_id)
        if not train:
            return error_response("Train not found", None)
        seats = self.repo.find_seats_by_date(train_id, journey_date)
        return success_response(MSG_SEATS_FOUND, seats)

    # ── Internal helpers ────────────────────────────────────────────────────
    def _get_all_trains_with_live_seats(self, journey_date: str) -> list[dict]:
        """
        Try ES for the train list, but always fetch live available_seats
        from Postgres so the count reflects real-time bookings.
        """
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
        """
        ES provides fuzzy/typo-tolerant matching; Postgres provides the
        live available_seats count.  Falls back to Postgres-only on error.
        """
        client = es_client.get_es_client()
        if client is not None:
            try:
                es_trains = es_client.search_trains(client, source, destination)
                if es_trains:
                    return self._patch_available_seats(es_trains, journey_date)
                # ES returned empty — could be a genuine no-result, return it
                return []
            except Exception as exc:
                logger.warning("ES search failed (%s) — falling back to Postgres", exc)

        return self.repo.find_trains_by_date(source, destination, journey_date)

    def _patch_available_seats(
        self, es_trains: list[dict], journey_date: str
    ) -> list[dict]:
        """
        ES documents hold a stale available_seats snapshot.
        Overwrite it with live counts from Postgres for the requested date.
        """
        # One DB call to get live counts for ALL trains on this date
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