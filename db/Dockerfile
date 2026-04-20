# TimescaleDB + pgvector for local Docker Compose (Phase 24 RAG).
# The stock Timescale image does not ship pgvector; we build it from source
# once at image build time.
FROM timescale/timescaledb:latest-pg16

USER root

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        build-essential \
        postgresql-server-dev-16; \
    git clone --branch v0.8.0 --depth 1 https://github.com/pgvector/pgvector.git /tmp/pgvector; \
    cd /tmp/pgvector; \
    make; \
    make install; \
    cd /; \
    rm -rf /tmp/pgvector; \
    apt-get purge -y git build-essential postgresql-server-dev-16; \
    apt-get autoremove -y; \
    rm -rf /var/lib/apt/lists/*

USER postgres
