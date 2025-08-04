# Go Ecommerce

This is a full-stack e-commerce platform built with a Go backend and a Next.js frontend.

## Features

- User authentication (registration and login)
- Browse products
- Add products to a shopping cart
- Checkout and place orders
- View order history
- Company and product reviews

## Tech Stack

### Backend

- **Language:** Go
- **Framework:** Standard library `net/http`
- **Database:** PostgreSQL (or any other SQL database compatible with `database/sql`)
- **Migrations:** `golang-migrate`
- **API Specification:** OpenAPI (Swagger)

### Frontend

- **Framework:** Next.js (React)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **State Management:** React Context

## Getting Started

### Prerequisites

- Go (version 1.20+)
- Node.js (version 18+)
- Docker and Docker Compose (for running a local database)

### Backend Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/go-ecommerce.git
    cd go-ecommerce
    ```

2.  **Install Go dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Run the database:**
    A `docker-compose.yml` file is not provided, but you can use the following command to start a PostgreSQL container:

    ```bash
    docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
    ```

4.  **Run database migrations:**
    You will need to have `migrate` installed. You can find installation instructions here: [https://github.com/golang-migrate/migrate/tree/master/cmd/migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

    ```bash
    migrate -path migrations -database "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable" up
    ```

5.  **Run the backend server:**
    ```bash
    go run cmd/go_ecommerce/main.go
    ```
    The server will start on `http://localhost:8080`.

### Frontend Setup

1.  **Navigate to the `web` directory:**

    ```bash
    cd web
    ```

2.  **Install Node.js dependencies:**

    ```bash
    npm install
    ```

3.  **Run the frontend development server:**
    ```bash
    npm run dev
    ```
    The frontend will be available at `http://localhost:3000`.

## API Documentation

The API is documented using the OpenAPI standard. The `openapi.json` file in the root of the project contains the full API specification. You can use a tool like Swagger UI to view the documentation in a more user-friendly format.

## Testing

To run the backend integration tests, you will need a running test database. The tests are configured to use a separate database to avoid interfering with development data.

```bash
go test ./tests/...
```

## Project Structure

```
.
├── cmd/go_ecommerce/main.go  # Backend application entry point
├── internal/                   # Internal application code
├── migrations/                 # Database migrations
├── pkg/                        # Reusable packages
│   ├── database/               # Database connection
│   ├── handlers/               # HTTP handlers
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Data models
│   ├── repositories/           # Database repositories
│   ├── services/               # Business logic
│   └── utils/                  # Utility functions
├── tests/                      # Integration tests
└── web/                        # Frontend application (Next.js)
    ├── src/
    │   ├── app/                # Next.js App Router
    │   ├── components/         # React components
    │   ├── context/            # React Context providers
    │   └── lib/                # Library functions (e.g., API client)
    └── ...
```
