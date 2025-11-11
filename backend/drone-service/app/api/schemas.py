from pydantic import BaseModel


class DroneCommandRequest(BaseModel):
    command: str
