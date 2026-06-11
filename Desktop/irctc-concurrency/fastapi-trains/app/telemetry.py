"""
telemetry.py — OTel bootstrap for fastapi-trains.

Call setup_tracing(app) once during startup (before any requests).
Instruments FastAPI routes, SQLAlchemy queries, and httpx calls automatically.
"""
import logging
import os

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.resource import ResourceAttributes

logger = logging.getLogger(__name__)


def setup_tracing(app, engine=None) -> None:
    """
    Wire up OTel tracing for the FastAPI app.

    Parameters
    ----------
    app    : the FastAPI application instance
    engine : optional SQLAlchemy engine — if provided, SQL queries get spans too
    """
    endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
    service_name = os.getenv("OTEL_SERVICE_NAME", "fastapi-trains")

    resource = Resource.create({
        ResourceAttributes.SERVICE_NAME: service_name,
        ResourceAttributes.SERVICE_VERSION: "1.0.0",
        "deployment.environment": "local",
    })

    exporter = OTLPSpanExporter(
        endpoint=f"{endpoint}/v1/traces",
        headers={},
    )

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)

    # Auto-instrument FastAPI — wraps every route handler in a span
    # span name = "GET /api/v1/trains/search" etc.
    FastAPIInstrumentor.instrument_app(
        app,
        tracer_provider=provider,
        excluded_urls="/health",   # skip health-check noise
    )

    # Auto-instrument httpx — captures outbound calls to gin-booking, springboot
    HTTPXClientInstrumentor().instrument(tracer_provider=provider)

    # Auto-instrument SQLAlchemy — captures every DB query as a child span
    if engine is not None:
        SQLAlchemyInstrumentor().instrument(
            engine=engine,
            tracer_provider=provider,
        )

    logger.info("OTel tracing initialized: service=%s endpoint=%s", service_name, endpoint)


def get_tracer(name: str = "fastapi-trains"):
    """Convenience helper for manual spans in service code."""
    return trace.get_tracer(name)
