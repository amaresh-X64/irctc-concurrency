from sqlalchemy.orm import Session
from app.trains.repository import TrainRepository
from app.constants.constants import MSG_TRAINS_FOUND, MSG_NO_TRAINS, MSG_SEATS_FOUND
from app.helpers.response import success_response, error_response
from datetime import date, datetime
import pytz
class TrainService:

    def __init__(self, db: Session):
        self.repo = TrainRepository(db)

    def _validate_date(self, journey_date: str):
        try:
            jd = date.fromisoformat(journey_date[:10])
        except ValueError:
            return None
        return jd

    def get_all_trains(self, journey_date):
        jd = self._validate_date(journey_date)
        if jd is None or jd < date.today():
            return error_response("Invalid or past journey date", None)
        trains = self.repo.find_all_by_date(journey_date)
        trains = self._filter_departed(trains, journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains)

    def search_trains(self, source, destination, journey_date):
        jd = self._validate_date(journey_date)
        if jd is None or jd < date.today():
            return error_response("Invalid or past journey date", None)
        trains = self.repo.find_trains_by_date(source, destination, journey_date)
        trains = self._filter_departed(trains, journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains)

    def get_seats(self, train_id, journey_date):
        train = self.repo.find_by_id(train_id)
        if not train:
            return error_response("Train not found", None)
        seats = self.repo.find_seats_by_date(train_id, journey_date)
        return success_response(MSG_SEATS_FOUND, seats)

    def _filter_departed(self, trains, journey_date):
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