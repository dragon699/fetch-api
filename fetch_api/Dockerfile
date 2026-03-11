FROM python:3.13-slim AS builder

WORKDIR /app

COPY common/requirements.txt .
COPY fetch_api/requirements.txt ./fetch_api-requirements.txt

RUN \
  pip install --user --no-cache-dir -r requirements.txt && \
  pip install --user --no-cache-dir -r fetch_api-requirements.txt



FROM python:3.13-slim

WORKDIR /app

COPY --from=builder /root/.local /root/.local

COPY common /app/common
COPY fetch_api/src /app/fetch_api/src
COPY fetch_api/*.py fetch_api/VERSION /app/fetch_api/

ENV PATH=/root/.local/bin:$PATH

CMD ["python", "-m", "fetch_api.app"]
