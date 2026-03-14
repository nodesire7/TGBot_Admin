# TGBot Admin - All-in-One Docker Image
# 包含: Go API + Python Bot + Web UI

#############################################
# Stage 1: Build Go API
#############################################
FROM golang:1.24-alpine AS api-builder

WORKDIR /app

# Copy API source
COPY api/ .

# Build API binary
ENV GOTOOLCHAIN=auto
RUN go mod init github.com/tgbot/admin || true && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tgbot-admin-api .

#############################################
# Stage 2: Build Python Bot
#############################################
FROM python:3.12-slim AS bot-builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc-dev \
    linux-libc-dev \
    python3-dev \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Create virtual environment
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Install Python dependencies
COPY bot/requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

#############################################
# Stage 3: Final Image
#############################################
FROM python:3.12-slim

WORKDIR /app

# Install runtime dependencies and supervisor
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    supervisor \
    curl \
    netcat-openbsd \
    postgresql-client \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Copy API binary from builder
COPY --from=api-builder /app/tgbot-admin-api /app/bin/api

# Copy Python virtual environment from builder
COPY --from=bot-builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy Bot source code
COPY bot/ /app/bot/

# Copy Web UI
COPY web/ /app/web/

# Copy migrations
COPY migrations/ /app/migrations/

# Copy supervisor configuration
COPY docker/supervisord.conf /etc/supervisor/conf.d/tgbot-admin.conf

# Copy entrypoint script
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Create non-root user
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app

# Create directories for supervisor and data
RUN mkdir -p /var/log/supervisor /var/run /app/data && \
    chown -R appuser:appuser /var/log/supervisor /var/run /app/data

USER appuser

# Environment defaults
ENV DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=tgbot \
    DB_PASSWORD=tgbot123 \
    DB_NAME=tgbot \
    REDIS_HOST=redis \
    REDIS_PORT=6379 \
    API_PORT=8000 \
    PYTHONUNBUFFERED=1

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8000/ || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["supervisord", "-c", "/etc/supervisor/conf.d/tgbot-admin.conf"]
