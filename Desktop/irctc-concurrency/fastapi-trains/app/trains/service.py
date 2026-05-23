from sqlalchemy.orm import Session
from app.trains.repository import TrainRepository
from app.constants.constants import MSG_TRAINS_FOUND, MSG_NO_TRAINS, MSG_SEATS_FOUND
from app.helpers.response import success_response, error_response
from typing import List


class TrainService:

    def __init__(self, db: Session):
        self.repo = TrainRepository(db)

    def search_trains(self, source, destination, journey_date):
        trains = self.repo.find_trains_by_date(source, destination, journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains) 

    def get_all_trains(self, journey_date):
        trains = self.repo.find_all_by_date(journey_date)
        msg = MSG_TRAINS_FOUND if trains else MSG_NO_TRAINS
        return success_response(msg, trains)  

    def get_seats(self, train_id, journey_date):
        train = self.repo.find_by_id(train_id)
        if not train:
            return error_response("Train not found", None)
        seats = self.repo.find_seats_by_date(train_id, journey_date)
        return success_response(MSG_SEATS_FOUND, seats)  