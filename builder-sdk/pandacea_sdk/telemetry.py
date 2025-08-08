import logging
import os
from typing import Optional

def init(service_name: Optional[str] = None) -> None:
    """
    Initialize OpenTelemetry for the SDK when PANDACEA_OTEL=1.
    Configures OTLP/HTTP exporter via OTEL_EXPORTER_OTLP_ENDPOINT.
    Sets JSON logging with trace correlation where supported.
    """
    if os.getenv("PANDACEA_OTEL") != "1":
        return

    try:
        from opentelemetry import trace
        from opentelemetry.sdk.resources import Resource
        from opentelemetry.sdk.trace import TracerProvider
        from opentelemetry.sdk.trace.export import BatchSpanProcessor
        from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
        from opentelemetry.propagate import set_global_textmap, get_global_textmap
        from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
        from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
        from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
        from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
        from opentelemetry.sdk.resources import SERVICE_NAME
    except Exception:
        # If otel not installed, silently no-op
        return

    endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
    svc = service_name or os.getenv("PANDACEA_SERVICE_NAME", "builder-sdk")
    res = Resource.create({SERVICE_NAME: svc})

    # Traces
    tracer_provider = TracerProvider(resource=res)
    span_exporter = OTLPSpanExporter(endpoint=endpoint)
    tracer_provider.add_span_processor(BatchSpanProcessor(span_exporter))
    trace.set_tracer_provider(tracer_provider)
    set_global_textmap(TraceContextTextMapPropagator())

    # Logs (best-effort; not required)
    try:
        logger_provider = LoggerProvider(resource=res)
        log_exporter = OTLPLogExporter(endpoint=endpoint)
        logger_provider.add_log_record_processor(BatchLogRecordProcessor(log_exporter))
        handler = LoggingHandler(level=logging.INFO, logger_provider=logger_provider)
        root = logging.getLogger()
        if not any(isinstance(h, LoggingHandler) for h in root.handlers):
            root.addHandler(handler)
        root.setLevel(getattr(logging, os.getenv("LOG_LEVEL", "INFO").upper(), logging.INFO))
    except Exception:
        pass


