import json, redis
from typing import Any
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword



class RedisClient:
    def __init__(self, host: str, port: int, db: int, password: str | None = None) -> None:
        self.client = redis.Redis(
            host=host,
            port=port,
            db=db,
            password=password,
            decode_responses=True
        )


    def ping(self) -> bool:
        return self.client.ping()


    @traced('retrieve cache')
    def get(self, key: str, span=None) -> Any:
        value = self.client.get(key)
        span.set_attributes({
            'redis.cache.key': key
        })

        if value is None:
            span.set_attributes({
                'redis.cache.status': 'miss'
            })
            return None

        try:
            json_value = json.loads(value)
            span.set_attributes(
                reword({
                    'redis.cache.status': 'hit',
                    'redis.cache.value': json_value
                })
            )

            return json_value

        except Exception as err:
            span.set_attributes(
                reword({
                    'redis.cache.status': 'error',
                    'redis.cache.value': value,
                    'redis.cache.error.message': str(err),
                    'redis.cache.error.type': type(err).__name__
                })
            )
            return value


    @traced('update cache')
    def set(self, key: str, value: Any, ttl: int | None = None, span=None) -> None:
        encoded_value = json.dumps(value)
        span.set_attributes(
            reword({
                'redis.cache.key': key,
                'redis.cache.value': encoded_value,
                'redis.cache.ttl': ttl
            })
        )

        if ttl:
            self.client.setex(key, ttl, encoded_value)
        
        else:
            self.client.set(key, encoded_value)
