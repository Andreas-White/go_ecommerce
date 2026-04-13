# Testing Specifications

This document outlines the testing architecture for the `go_ecommerce` backend. Agents and developers should refer to this guide to understand how to correctly initialize the test suite, navigate the security middlewares (JWT and CSRF) during tests, and verify their logic blocks without polluting the production database.

## Test Setup & Migrations (`main_integration_test.go`)
The testing suite relies on a dedicated isolated database to prevent overlaps with local development data.
- **`TestMain(m *testing.M)`**: This acts as the universal entry point for all tests executed in the package.
- **Auto-migrations**: `TestMain` automatically loads test configurations (`LoadTestConfig`), spins up a connection via the `golang-migrate/migrate` library pointing to the `file://../migrations` directory, and runs `mig.Up()`. This guarantees the test database perfectly mirrors the production schema before any test executes.

## Auth Helpers (`commons.go`)
Because the application strictly enforces `AuthenticateJWT` and `CSRFMiddleware` wrappers around secure routes, standard HTTP simulation fails without handling cookies and headers. We use the **`TestAuthData`** struct to bypass and seamlessly inject state:

```go
type TestAuthData struct {
	JWTToken  string
	CSRFToken string
	Cookies   []*http.Cookie
}
```

- **`registerTestUserAuth` / `loginUserAndGetTokenAuth`**: These helpers should be heavily utilized. They programmatically hit the registration/login endpoints, intercept the `set-cookie` header arrays sent back by the server, and populate a returned `*TestAuthData` pointer. 
- You can pass the populated `*TestAuthData` to any state-mutating test request using the helper `addAuthHeaders(req, authData)`. It securely applies the expected `X-CSRF-Token` header and necessary HttpOnly cookies native to the simulated request.

## CSRF Flow in Tests
To combat CSRF failures on strict routes, you cannot blindly fire `POST` or `PUT` requests in testing:
1. Every mutating request natively expects an `X-CSRF-Token` mapped to its corresponding session cookie.
2. The testing suite establishes this by internally calling `getCSRFTokenForEndpoint` (which sends a `GET` lookup against the target payload or the general `/csrf` dummy block). 
3. This fetches the `csrf_token` cookie safely, allowing the subsequent `POST` mutation test to succeed dynamically.

## Standard Verification Command
When finalizing any feature implementation inside the User, Auth, Order, or Product scopes, you must verify against regressions. Ensure all logic satisfies previous constraints by running the integrated suite from the root level:

```bash
go test ./tests/...
```
