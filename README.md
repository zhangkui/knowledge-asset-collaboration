# Enterprise Knowledge Asset & Document Collaboration Platform

Go REST API plus Vue 3 frontend for enterprise knowledge spaces, versioned documents, collaboration, review, permissions, search, sharing, notifications, recycle bin and audit reporting.

Run backend with `go run .`; health endpoint is `GET /healthz`. Run infrastructure with `docker compose up -d`. Default local administrator: `admin@example.com` / `ChangeMe!123`.

List responses use `{items,page,page_size,total}` and errors use `{error:{code,message}}`. OpenAPI is in `docs/openapi.yaml`.
