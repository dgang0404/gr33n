# TimescaleDB + pgvector + PostGIS for local Docker Compose (Phase 24 RAG).
#
# Upstream timescale/timescaledb OSS images are Alpine-based (apk), not Debian.
# Stock Timescale does not ship pgvector; we build it from source once at image build time.
#
# Use LLVM 19 toolchains only: Timescale's postgres is built with clang-19 / llvm-lto under
# /usr/lib/llvm19/. The stock `clang`/`llvm` apk metapackages pull LLVM 21 (huge); they are
# redundant here and have triggered flaky extractions (I/O / integrity) on some hosts.
FROM timescale/timescaledb:latest-pg16

USER root

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
        musl-dev \
        clang19 \
        llvm19 \
        postgresql16-dev; \
    git clone --branch v0.8.0 --depth 1 https://github.com/pgvector/pgvector.git /tmp/pgvector; \
    cd /tmp/pgvector; \
    command -v clang-19 >/dev/null 2>&1 || { echo "clang-19 missing (apk clang19)"; exit 1; }; \
    test -x /usr/lib/llvm19/bin/llvm-lto || { echo "llvm-lto missing (apk llvm19)"; exit 1; }; \
    make OPTFLAGS=""; \
    make install; \
    cd /; \
    rm -rf /tmp/pgvector; \
    apk del .pgvector-build; \
    rm -rf /var/cache/apk/*

# PostGIS: gr33n schema uses GEOMETRY (farms/zones). Alpine's `apk add postgis` targets the
# distro PostgreSQL (currently PG18) and cannot load into Timescale's bundled PG16 — you get
# "PostGIS built for PostgreSQL 18.x cannot be loaded in PostgreSQL 16.x". Build PostGIS from
# source against Timescale's pg_config. Flags: --without-raster (no GDAL), --without-protobuf
# (no protobuf-c dev package; disables MVT/Geobuf helpers — fine for farm/zones GEOMETRY).
RUN set -eux; \
    VER="v$(cut -d. -f1,2 /etc/alpine-release)"; \
    if ! grep -q 'alpine/.*/community' /etc/apk/repositories 2>/dev/null; then \
      echo "https://dl-cdn.alpinelinux.org/alpine/${VER}/community" >> /etc/apk/repositories; \
    fi; \
    apk update; \
    apk add --no-cache --virtual .postgis-build \
        wget ca-certificates \
        build-base automake autoconf libtool pkgconf \
        flex bison perl \
        geos-dev proj-dev \
        json-c-dev libxml2-dev sqlite-dev \
        clang19 llvm19 \
        ; \
    export PATH="/usr/local/bin:${PATH}"; \
    POSTGIS_VER=3.4.3; \
    cd /tmp; \
    wget -q "https://download.osgeo.org/postgis/source/postgis-${POSTGIS_VER}.tar.gz"; \
    tar -xzf "postgis-${POSTGIS_VER}.tar.gz"; \
    cd "postgis-${POSTGIS_VER}"; \
    ./configure \
        --with-pgconfig=/usr/local/bin/pg_config \
        --without-raster \
        --without-protobuf; \
    if ! command -v clang-19 >/dev/null 2>&1; then \
      apk add --no-cache clang19 2>/dev/null || true; \
    fi; \
    if ! command -v clang-19 >/dev/null 2>&1; then \
      for c in /usr/lib/llvm19/bin/clang /usr/bin/clang-21 /usr/bin/clang; do \
        if [ -x "$c" ]; then ln -sf "$c" /usr/bin/clang-19; break; fi; \
      done; \
    fi; \
    command -v clang-19 >/dev/null 2>&1 || { echo "clang-19 missing (PGXS bitcode for PostGIS)"; exit 1; }; \
    test -x /usr/lib/llvm19/bin/llvm-lto || { echo "llvm-lto missing (apk llvm19)"; exit 1; }; \
    make -j"$(nproc)"; \
    make install; \
    cd /; \
    rm -rf "/tmp/postgis-${POSTGIS_VER}"*; \
    apk del .postgis-build; \
    apk add --no-cache geos proj json-c libxml2 sqlite-libs; \
    rm -rf /var/cache/apk/*

USER postgres
