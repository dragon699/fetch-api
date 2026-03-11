import uvicorn
from fetch_api.settings import settings
from fetch_api.src.api import app
from fetch_api.src.telemetry.logging import logger


if __name__ == "__main__":
    uvicorn.run(
        app,
        host=settings.listen_host,
        port=settings.listen_port,
        log_config=logger.get_uvicorn_config()
    )
