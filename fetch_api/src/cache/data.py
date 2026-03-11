from common.telemetry.src.tracing.wrappers import traced



class CachedResponse:
    def __init__(self, cached_at: str = '', status_code: int = 200, headers: dict | None = None, json_data: dict | None = None) -> None:
        self.cached_at = cached_at
        self.status_code = status_code
        self.headers = headers or {}
        self._json = json_data or {}

    @traced('serve cache')
    def json(self, span=None) -> dict:
        return {
            'cached': True,
            'cached_at': self.cached_at,
            **self._json
        }
