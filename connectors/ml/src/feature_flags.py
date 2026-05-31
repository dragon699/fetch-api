from common.feature_flags.launchdarkly import LaunchDarklyClient
from connectors.ml.settings import settings
from connectors.ml.src.telemetry.logging import log


ld_flags = LaunchDarklyClient(
    sdk_key=settings.launchdarkly_sdk_key,
    service_name=settings.name,
    logger=log
)


def is_ai_summary_enabled(route: str | None = None, provider: str | None = None) -> bool:
    return ld_flags.bool_variation(
        flag_key=settings.launchdarkly_flag_enable_ai_summary,
        context_key=f'{settings.name}:ai-summary',
        default=True,
        attributes={
            'route': route,
            'provider': provider
        }
    )

