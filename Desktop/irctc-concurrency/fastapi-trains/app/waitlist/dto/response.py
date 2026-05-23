from pydantic import BaseModel

class WaitlistResponse(BaseModel):
    position: int
    total_waiting: int
    confirmation_probability: float
    message: str