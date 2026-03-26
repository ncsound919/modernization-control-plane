FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .
COPY orchestrator/ orchestrator/

ENV REDIS_URL=redis://localhost:6379 \
    POSTGRES_DSN=postgres://mcp:mcp_dev_password@localhost:5432/mcp \
    KAFKA_BROKERS=localhost:9092 \
    PORT=8080

EXPOSE 8080

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
