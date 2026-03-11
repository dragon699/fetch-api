import uvicorn
from connectors.ml.settings import settings
from connectors.ml.src.api import app
from connectors.ml.src.telemetry.logging import logger


if __name__ == "__main__":
    uvicorn.run(
        app,
        host=settings.listen_host,
        port=settings.listen_port,
        log_config=logger.get_uvicorn_config()
    )
