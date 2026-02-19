# Heimdall

Autonomous AI monitoring agent for production applications.

Heimdall replaces passive, noisy alerting with an intelligent agent that watches your systems 24/7, investigates anomalies autonomously, and surfaces only what matters.

## Quick Start

```bash
# Frontend
make dev-frontend

# Backend
make dev-backend

# Both concurrently
make dev
```

## Project Structure

```
heimdall/
├── frontend/    # Vue 3 + Vite + TypeScript
├── backend/     # Go backend
├── docs/        # Documentation
└── Makefile     # Root-level task runner
```

See [docs/vision.md](docs/vision.md) for the full product vision and [docs/plans/blueprint.md](docs/plans/blueprint.md) for the technical blueprint.
