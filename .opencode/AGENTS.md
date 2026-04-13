# Source of Truth: Spec-Driven Development (SDD) Workflow

## Tech Stack
- **Backend:** Go v1.20+ (Specifically running v1.23.0, utilizing `net/http` standard library)
- **Database:** PostgreSQL via `database/sql` (Note: We use `db` tags for mapping, NOT GORM naming conventions).
- **Frontend:** Next.js 

## Architecture Rules
**Strict 3-Tier Pattern enforced:**
1. **Handler (`pkg/handlers/`)**: Responsible for HTTP routing, request parsing, checking CSRF, and returning JSON responses.
2. **Service (`pkg/services/`)**: Responsible for core business logic, validations, and orchestrating repositories.
3. **Repository (`pkg/repositories/`)**: Responsible entirely for persistent data access and executing SQL queries.

## Coding Standards
- **Authentication:** Must implement and respect JWTs stored securely in `HttpOnly` cookies.
- **Security:** `X-CSRF-Token` headers are mandated for state-mutating operations to prevent cross-site request forgery.

## Role Constraints
- **Producer Role:** Any user configured with `IsProducer: true` retains strict ownership over the products they create. By extension, this role grants the authorization needed to create products, fulfil orders, and query producer-specific domain data.

## Dependency Rule
**STRICT LIMITATION:** Do NOT add any external libraries to the project beyond what is already tracked natively in `go.mod` and `package.json` (such as the existing `google/uuid` module). Rely strictly on current modules or the Go Standard Library.

## Project Directory Structure
All future code must align with this established physical structure:

```text
/go_ecommerce
├── .opencode/           # Agent source-of-truth instructions
├── cmd/                 # Application Entrypoints (e.g., main.go)
├── internal/            # Private application code
├── migrations/          # Raw SQL migration files  
├── pkg/                 # Go Application Code
│   ├── handlers/        # API Controllers
│   ├── models/          # Structs, DB schemas, and DTOs
│   ├── repositories/    # Database Logic
│   ├── services/        # Business Logic
│   └── utils/           # Shared utility tools
├── tests/               # Test suites (Integration, Unit)
└── web/                 # Next.js Frontend Application
```

## Testing Rules (CRITICAL)
1. **Verify Before Submit:** Before completing any task, the agent must check `doc/specs/TESTING.md`.
2. **Integration Checks:** Ensure that matching integration tests in `/tests/` are updated or created.
3. **Common Helpers:** Utilize `commons.go` for `TestAuthData` and CSRF handling in all new tests.
4. **Execution:** Always run `go test ./tests/...` to verify logic and security middleware.

## Project Knowledge Base
For implementation details, always refer to the following specifications:
- **System Architecture:** @doc/specs/SYSTEM.md
- **Database Schema:** @doc/specs/DATA_MODEL.md
- **API Contracts:** @doc/specs/API_SPEC.md
- **Business Logic:** @doc/specs/USER_FLOWS.md
- **Testing Procedures:** @doc/specs/TESTING.md

## Definition of Done (DoD)
Before completing a task, verify:
1. **Sync Check:** Are changes reflected in `openapi.json`, `API_SPEC.md` (API) and `DATA_MODEL.md` (DB)?
2. **Security:** If a new endpoint was added, does it have the `auth_middleware` and a CSRF check?
3. **Architecture:** Is the business logic in the `Service` and NOT the `Handler`?
4. **Tests:** Did you run `go test ./tests/...`?
