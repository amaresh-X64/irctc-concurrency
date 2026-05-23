from pydantic import BaseModel

class WaitlistRequest(BaseModel):
    user_id: int
    train_id: int
    journey_date: str

