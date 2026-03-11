FROM python:3.13-slim AS builder

WORKDIR /app

COPY common/requirements.txt .

RUN pip install --user --no-cache-dir -r requirements.txt



FROM python:3.13-slim

WORKDIR /app

COPY --from=builder /root/.local /root/.local

COPY common /app/common
COPY connectors/grafana/src /app/connectors/grafana/src
COPY connectors/grafana/templates /app/connectors/grafana/templates
COPY connectors/grafana/*.py connectors/grafana/VERSION /app/connectors/grafana/

ENV PATH=/root/.local/bin:$PATH

CMD ["python", "-m", "connectors.grafana.app"]
