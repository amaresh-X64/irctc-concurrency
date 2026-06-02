from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.trains.controller import router as trains_router, internal_router as trains_internal_router
from app.waitlist.controller import router as waitlist_router
from app.auth.controller import router as auth_router
from app.constants.constants import APP_NAME, APP_VERSION, API_PREFIX
from app.middleware.auth import SessionLocal
from app.trains.service import TrainService
import logging
logging.basicConfig(level=logging.INFO)
app = FastAPI(
    title=APP_NAME,
    version=APP_VERSION,
    swagger_ui_parameters={"persistAuthorization": True},
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(auth_router,             prefix=API_PREFIX)
app.include_router(trains_router,           prefix=API_PREFIX)
app.include_router(trains_internal_router)
app.include_router(waitlist_router,         prefix=API_PREFIX)


@app.on_event("startup")  
def startup_event():
    db = SessionLocal()
    try:
        TrainService.bootstrap_es_index(db)
    finally:
        db.close()


@app.get("/health")
def health():
    return {"status": "ok", "service": APP_NAME}