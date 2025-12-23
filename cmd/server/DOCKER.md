# ccNexus Server Docker

## Build
From repo root:

```bash
docker build -f cmd/server/Dockerfile -t ccnexus-server:local .
```

## Run (docker)

```bash
docker run --rm -p 3000:3000 \
  -e CCNEXUS_DATA_DIR=/data \
  -v ccnexus-data:/data \
  ccnexus-server:local
```

If you see an SQLite error like "unable to open database file", it's almost always a permission problem on the mounted `/data` volume. The provided image runs as root by default to avoid this.

WebUI: http://localhost:3000/ui/

## Run (docker compose)
From repo root:

```bash
docker compose -f cmd/server/docker-compose.yml up --build
```

Data persists in the named volume `ccnexus-data`.

## Environment variables
- `CCNEXUS_DATA_DIR`: data dir (default in container: `/data`)
- `CCNEXUS_DB_PATH`: optional absolute db path (default: `${CCNEXUS_DATA_DIR}/ccnexus.db`)
- `CCNEXUS_PORT`: override listen port (default: `3000`)
- `CCNEXUS_LOG_LEVEL`: override log level
