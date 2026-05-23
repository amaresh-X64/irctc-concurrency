from pydantic import BaseModel

class TrainSearchRequest(BaseModel):
    source: str
    destination: str
    journey_date: str  
    class Config:
        str_strip_whitespace = True
        str_to_upper = True