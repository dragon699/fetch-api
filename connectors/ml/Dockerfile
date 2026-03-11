FROM python:3.13-slim AS builder

WORKDIR /app

COPY common/requirements.txt .
COPY connectors/ml/requirements.txt ./ml-requirements.txt

RUN \
  pip install --user --no-cache-dir -r requirements.txt && \
  pip install --user --no-cache-dir -r ml-requirements.txt



FROM python:3.13-slim

WORKDIR /app

COPY --from=builder /root/.local /root/.local

COPY common /app/common
COPY connectors/ml/src /app/connectors/ml/src
COPY connectors/ml/templates /app/connectors/ml/templates
COPY connectors/ml/*.py connectors/ml/VERSION /app/connectors/ml/

ENV PATH=/root/.local/bin:$PATH

CMD ["python", "-m", "connectors.ml.app"]
