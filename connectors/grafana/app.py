import uvicorn
from connectors.grafana.settings import settings
from connectors.grafana.src.api import app
from connectors.grafana.src.telemetry.logging import logger


if __name__ == "__main__":
    uvicorn.run(
        app,
        host=settings.listen_host,
        port=settings.listen_port,
        log_config=logger.get_uvicorn_config()
    )
