# Go Web Service Example

This is an example Go web service demonstrating a robust project structure with a custom framework for configuration, logging, metrics, and web serving. It includes both a REST API and a server-side rendered UI.

## Features

- **Dual-Layer Interface**: Provides both a RESTful API and an HTMX-powered web UI.
- **Robust Framework**:
    - **Configuration**: Multi-source configuration using flags, environment variables, and YAML files (Viper).
    - **Observability**: Built-in Prometheus metrics, structured logging (slog), and health checks.
    - **Middleware**: Extensible middleware for logging, metrics, and header propagation.
    - **Graceful Shutdown**: Handles OS signals for clean termination.
    - **Profiling**: Integrated pprof support for performance analysis.
- **Modern UI**: Uses Go templates, Tailwind CSS, and HTMX for a dynamic experience without complex frontend frameworks.

## Architecture & Patterns

This project follows several key design patterns to ensure scalability, maintainability, and a high-quality developer experience.

### 1. Custom Internal Framework
Common infrastructure concerns (logging, configuration, metrics, web serving) are abstracted into the `internal/framework` package.
- **Benefit**: Ensures consistency across different parts of the application and keeps business logic clean of infrastructure boilerplate.

### 2. Functional Options Pattern
Used extensively in the initialization of components like the Web Server and REST Client (e.g., `web.NewServer(web.WithPort(...))`).
- **Benefit**: Provides a clean, extensible API for component configuration without bloating constructor signatures.

### 3. Middleware Pattern
Applied to both the incoming HTTP server requests and outgoing REST client requests.
- **Benefit**: Allows for composable, cross-cutting concerns such as observability (metrics/logging), header propagation, and security to be applied uniformly.

### 4. Dependency Injection & Separation of Concerns
The project structure clearly separates the transport layer (`internal/api`), business logic (`internal/service`), and external integrations (`internal/joker`).
- **Benefit**: Enhances testability by allowing components to be easily mocked and ensures that changes to one layer don't leak into others.

### 5. Graceful Shutdown
The application listens for OS signals (SIGTERM, SIGINT) to perform a clean termination.
- **Benefit**: Ensures that in-flight requests are completed and resources (like database connections or file handles) are closed properly, which is critical for reliability in containerized environments like Kubernetes.

### 6. Modern Monolith with HTMX
Combines server-side Go templates with HTMX for dynamic user interactions.
- **Benefit**: Delivers a modern, "SPA-like" user experience without the build complexity or overhead of a heavy JavaScript framework.

## Configuration

The application uses a hierarchical configuration system. Precedence (highest to lowest): Flags > Environment Variables > Config File.

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port to listen on for the main web server. | `8080` |
| `--debug-port` | If set, enables pprof on this port. | (empty) |
| `--profiles` | Comma-separated list of profiles (e.g., `dev`). Loads `config-<profile>.yaml`. | (empty) |
| `--log-level` | Log level: `debug`, `info`, `warn`, `error`. | `info` |
| `--log-json` | Enable structured JSON logging. | `false` |
| `--log-source` | Include source file and line number in logs. | `false` |

### Environment Variables

Environment variables correspond to flags but are uppercase and use underscores (e.g., `PORT`, `LOG_LEVEL`, `PROFILES`).

### Config Files

The application looks for `config.yaml` in the current directory and `../../`. If profiles are active, it also merges `config-<profile>.yaml`.

Example `config.yaml`:
```yaml
PORT: 9090
JOKES:
  LOGGING: DEBUG
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/helloworld` | Simple greeting. |
| `POST` | `/helloworld` | Greeting with a custom name (JSON body: `{"name": "..."}`). |
| `GET` | `/dailyjoke` | Fetches a joke from the downstream joke service. |

## UI Components

| Path | Description |
|------|-------------|
| `/` | Redirects to `/home`. |
| `/home` | The main dashboard/landing page. |
| `/static/*` | Serves static assets (CSS, JS). |

## Observability

- **Health Checks**:
    - `/healthz`: Liveness probe.
    - `/readyz`: Readiness probe.
- **Metrics**: 
    - `/metrics`: Prometheus metrics endpoint.
    - `/metrics/docs`: Auto-generated documentation for registered metrics.
- **Profiling**: Enabled via `--debug-port`.

## Getting Started

### Prerequisites

- Go 1.22+
- Make (optional)

### Running Locally

```bash
# Using Go directly
go run cmd/main.go --port 8080

# Using a profile
go run cmd/main.go --profiles dev

# Using Make
make run

# Using Docker
make docker/build
make docker/run
```

### Testing

```bash
make test
```
