from common.feature_flags.launchdarkly import LaunchDarklyClient
from fetch_api.settings import settings
from fetch_api.src.telemetry.logging import log


ld_flags = LaunchDarklyClient(
    sdk_key=settings.launchdarkly_sdk_key,
    service_name=settings.name,
    logger=log
)


def is_ai_summary_enabled(route: str | None = None, connector: str | None = None) -> bool:
    return ld_flags.bool_variation(
        flag_key=settings.launchdarkly_flag_enable_ai_summary,
        context_key=f'{settings.name}:ai-summary',
        default=True,
        attributes={
            'route': route,
            'connector': connector
        }
    )

