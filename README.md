
# video-orchestrator

HTTP server for MP4 upload and HLS file serving, backed by PostgreSQL.

## Prerequisites

Docker, Git, Go 1.26. Copy `.env.example` to `.env` and fill in the values.

## How it works

1. Client uploads a MP4 via `POST /jobs`
2. The job is saved in PostgreSQL with status `PENDING`
3. `video_segmenter` picks up the file and generates HLS segments
4. The watcher detects `index.m3u8` and marks the job as `DONE`
5. HLS files are served via `GET /streams/video/:id/:filename`

## Installing

```bash
git clone git@github.com:M2G/video-orchestrator.git
cd video-orchestrator
cp .env.example .env
docker compose up
```

## Build

```bash
make build
```

## Running the tests

```bash
go test ./internal/...
```

## API Endpoints

### Upload a video
- Path: `/jobs`
- Method: `POST`
- Body: `multipart/form-data` with a `.mp4` file
- Response: `201`

### Get job status
- Path: `/jobs/{id}`
- Method: `GET`
- Response: `200`

### Serve HLS file
- Path: `/streams/video/{id}/{filename}`
- Method: `GET`
- Response: `200`

## Environment variables

- `DATABASE_URL` : PostgreSQL connection string
- `UPLOAD_DIR` : Directory where MP4 files are saved
- `STREAMS_DIR` : Directory watched for HLS output
- `HTTP_ADDR` : HTTP server address (default `:8181`)
- `LOG_LEVEL` : Log level: debug, info, warn, error (default `info`)

## Stack

- Go 1.26
- PostgreSQL 18
- sqlc + pgx/v5
- logrus