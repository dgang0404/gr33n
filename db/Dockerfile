# TimescaleDB + pgvector for local Docker Compose (Phase 24 RAG).
#
# Upstream timescale/timescaledb OSS images are Alpine-based (apk), not Debian.
# Stock Timescale does not ship pgvector; we build it from source once at image build time.
FROM timescale/timescaledb:latest-pg16

USER root

# Match Timescale's own pgvector build pattern (see timescaledb-docker on GitHub).
# postgresql*-dev lives in the community repo on many Alpine images.
RUN set -eux; \
    VER="v$(cut -d. -f1,2 /etc/alpine-release)"; \
    if ! grep -q 'alpine/.*/community' /etc/apk/repositories 2>/dev/null; then \
      echo "https://dl-cdn.alpinelinux.org/alpine/${VER}/community" >> /etc/apk/repositories; \
    fi; \
    apk update; \
    apk add --no-cache --virtual .pgvector-build \
        git \
        build-base \
        clang \
        llvm \
        llvm-dev \
        musl-dev \
        postgresql16-dev; \
    git clone --branch v0.8.0 --depth 1 https://github.com/pgvector/pgvector.git /tmp/pgvector; \
    cd /tmp/pgvector; \
    make OPTFLAGS=""; \
    make install; \
    cd /; \
    rm -rf /tmp/pgvector; \
    apk del .pgvector-build; \
    rm -rf /var/cache/apk/*

USER postgres
