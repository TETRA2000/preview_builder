# Preview Builder

Automated Docker-based preview environment generator for GitHub pull requests. Monitors open PRs, builds Docker containers for each one, and deploys them as live preview sites accessible via unique HTTP ports.

Originally built to serve Angular.io documentation previews.

## How It Works

1. Fetches all open PRs from the configured GitHub repository
2. Compares each PR's latest commit SHA against the last recorded build in a local SQLite database
3. Skips PRs that are unmergeable or haven't changed since the last build
4. For changed PRs: checks out the PR branch, merges with master, and builds a Docker image
5. Starts a container with port mappings based on the PR number:
   - **HTTP (built site):** `10000 + PR number` (e.g., PR #42 at `localhost:10042`)
   - **Dev server:** `30000 + PR number` (e.g., PR #42 at `localhost:30042`)

## Prerequisites

- Docker
- Go 1.6+ (for building from source)
- A GitHub OAuth token with repo access

## Project Structure

```
├── src/
│   ├── main.go                        # Entry point
│   └── preview_builder/
│       ├── preview_builder.go         # Core logic (GitHub, Docker, DB)
│       └── vendor/                    # Vendored Go dependencies
├── angular_io/
│   ├── Dockerfile                     # PR-specific image (builds Angular docs)
│   ├── Dockerfile.base                # Base image (Node.js 6.10 + Apache HTTPd 2.2)
│   ├── build_pr_image.sh             # Shell script to build a PR image
│   ├── build_base_image.sh           # Shell script to build the base image
│   ├── httpd-foreground              # Apache startup script
│   ├── repo/                          # Git submodule: angular.io repository
│   └── angular/                       # Git submodule: angular repository
├── setup/
│   └── init_db.sql                    # SQLite schema
├── Dockerfile                         # Main app container (Go + Docker client)
├── docker-compose.yml                 # Docker Compose orchestration
└── .travis.yml                        # Travis CI config
```

## Setup

### 1. Configure GitHub Access

Edit `src/preview_builder/preview_builder.go` and set your OAuth token in `CreateGithubClient()`:

```go
ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: "your-token-here"},
)
```

Also update the repository owner and name in `GetPRList()`, `GetListCommits()`, and `main.go` (currently hardcoded as `"TETRA2000"` / `"reponame"`).

### 2. Initialize the Database

```bash
sqlite3 preview_builder_data.db < setup/init_db.sql
```

### 3. Build the Base Image

```bash
cd angular_io && ./build_base_image.sh
```

## Running

### With Docker Compose

```bash
docker-compose up
```

This mounts the Docker socket so the app can build and run sibling containers.

### From Source

```bash
cd src
GOPATH=$(dirname $PWD) go build -o ../preview_builder main.go
cd ..
./preview_builder
```

## Database Schema

Single table tracking PR build state:

```sql
CREATE TABLE pull_request (
    number INTEGER PRIMARY KEY,
    latest_commit_sha1 VARCHAR(255),
    updated_at TEXT
);
```

## Configuration

| Setting | Location | Default |
|---|---|---|
| GitHub owner/repo | `preview_builder.go`, `main.go` | `TETRA2000/reponame` |
| OAuth token | `preview_builder.go` `CreateGithubClient()` | empty |
| SQLite database path | `preview_builder.go` `OpenSqliteDb()` | `./preview_builder_data.db` |
| Build timeout | `preview_builder.go` `BuildPreviewImage()` | 15 minutes |
| HTTP port range | `preview_builder.go` `StartPreviewContainer()` | `10000 + PR#` |
| Dev server port range | `preview_builder.go` `StartPreviewContainer()` | `30000 + PR#` |

## CI

Travis CI is configured to build and test against Go 1.6.4, latest 1.x, and master.
