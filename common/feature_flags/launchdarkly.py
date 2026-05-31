import atexit
import logging
from typing import Any

import ldclient
from ldclient import Config, Context
from ldotel.tracing import Hook


class LaunchDarklyClient:
    def __init__(self, sdk_key: str | None, service_name: str, logger: logging.Logger) -> None:
        self.client = None
        self.service_name = service_name
        self.log = logger

        if not sdk_key:
            self.log.warning('LaunchDarkly SDK key is missing, feature flags will use fallback values')
            return

        config = Config(
            sdk_key=sdk_key,
            hooks=[Hook()]
        )
        ldclient.set_config(config=config)
        self.client = ldclient.get()
        atexit.register(self.shutdown)


    def bool_variation(
        self,
        flag_key: str,
        context_key: str,
        default: bool,
        attributes: dict[str, Any] | None = None
    ) -> bool:
        if self.client is None:
            return default

        try:
            builder = Context.builder(context_key).kind('service').name(self.service_name)
            for key, value in (attributes or {}).items():
                if value is None:
                    continue
                builder.set(key, value)

            context = builder.build()
            return self.client.variation(flag_key, context, default)

        except Exception as err:
            self.log.warning('LaunchDarkly flag evaluation failed', extra={
                'flag_key': flag_key,
                'error': str(err)
            })
            return default


    def shutdown(self) -> None:
        if self.client is not None:
            self.client.close()

