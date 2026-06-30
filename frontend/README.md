# WealthLens Frontend

Next.js App Router frontend for the WealthLens deterministic portfolio tracker.

## Local development

From the repository root:

```bash
cp frontend/.env.example frontend/.env.local
make db-up
make migrate-up
make run
```

In a second terminal:

```bash
make frontend-dev
```

The frontend runs at `http://localhost:3000` and communicates with the Go API
from the Next.js server. Access and refresh tokens are stored in HTTP-only
cookies and are not exposed to client-side JavaScript.

## Validation

```bash
make frontend-check
```
