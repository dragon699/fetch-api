# Fetch API AI summaries

Requests with `ai: true` return upstream telemetry and the most recent completed
Redis summary, when one exists. They do not wait for the ML connector. The existing
`ai_summary.answer`, `provider`, `model`, and timing fields remain available, with
`cached: true`, `cached_at`, and `stale` indicating when the answer was generated
and whether its telemetry has changed or its refresh interval has elapsed.

On the first request for a summary stream, the response has no `ai_summary` yet.
A worker generates it and saves it to Redis; a subsequent request can retrieve it.
For Glance's `cache: 10s` widgets, the summary appears on a subsequent dashboard
load after the widget cache expires. Generation does not update an already rendered
browser page directly. No Glance template or timeout changes are required.

Background refresh is triggered by requests, not by a periodic polling job. A worker
resolves the configured ML provider and model independently of the dashboard request.
It regenerates when the telemetry fingerprint changes or the summary is five minutes
old. Relative-duration strings recognized by the existing `omit_volatile_data` helper
are normalized so their formatting alone does not force generation.

Each stream is scoped by route, upstream query parameters, prompt, and instruction
template. Provider/model results are stored separately. Changing telemetry retains
the last completed answer while its replacement runs. After the worker observes a
provider/model switch, future requests select that provider/model's cache; if it has
no result yet, they omit the summary until generation succeeds. A request immediately
preceding discovery of the switch can still return the previous known selection.

Failures preserve the last good answer for the same selection. Redis locks prevent
concurrent generation of a stream across replicas, and a bounded thread pool prevents
unlimited queued work. When all workers are busy, a later request retries scheduling.
If Redis is unavailable, telemetry is still returned without a summary; cache socket
operations have a short timeout. Jobs run inside Fetch API and can be interrupted by
pod termination. They are not a durable task queue; later requests retry after any
expired refresh lock is released. Completed summaries remain in Redis until their TTL.

| Fetch API variable | Default | Purpose |
| --- | --- | --- |
| `AI_SUMMARY_BACKGROUND_TIMEOUT_SECONDS` | `100` | Background ML response timeout; connection timeout is 5 seconds |
| `AI_SUMMARY_REFRESH_INTERVAL_SECONDS` | `300` | Refresh unchanged data after this interval |
| `AI_SUMMARY_BACKGROUND_WORKERS` | `2` | Maximum outstanding jobs per process |
| `AI_SUMMARY_CACHE_TIMEOUT_SECONDS` | `0.5` | Redis socket/connect timeout for summaries |
| `REDIS_CACHE_TTL` | `86400` | Retention for completed summaries |

`AI_SUMMARY_REQUESTS_TIMEOUT=4` can stay as configured. It no longer limits automatic
summary generation; direct `/ml/ask` calls retain their existing synchronous behavior.
The ML connector still enforces its own provider timeout, which must allow enough time
for generation (the configured 30-second OpenClaw timeout already covers 5–10 seconds).

Deploy the updated Fetch API image to enable this behavior. New summary keys use the
`ai-summary:v1:` namespace, so old synchronous summary cache entries are not reused.
Existing telemetry and direct-query caches retain their behavior.

Run focused regression tests from the repository root:

```bash
python3 -m unittest discover -s fetch_api/tests -v
```
