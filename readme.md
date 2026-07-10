# Enterprise-Grade E-Commerce Application

A decoupled, high-performance e-commerce catalog and purchase simulation built with **Go** (backend) and **React** (frontend), powered by **SQLite** (persistent embedded database) and packaged as a single-container **Docker** deployment.

## Challenge Verification Details
*   **Example CSV Download Date:** 2026-07-09

---

## Architectural & Design Decisions

### 1. Persistent, Serverless SQL Backend (SQLite)
*   **Decision:** Replaced MongoDB with **SQLite** using the pure-Go driver (`github.com/glebarez/go-sqlite`).
*   **Rationale:** SQLite is serverless and database-file-based. It eliminates the operational overhead of running a separate database container during local development. By utilizing the pure-Go driver, we bypass any CGO compiler compiler constraints, allowing simple, robust multi-stage Docker builds.
*   **Data Integrity:** SQL transactions are utilized for checkout processes, guaranteeing ACID consistency. Product `sku` indexes are enforced at the database schema level.

### 2. Standalone Commands Layout (`cmd/`)
*   **Decision:** Decoupled administrative tasks (database seeding) from the web server runtime by structuring execution commands under a `cmd/` directory.
*   *   `backend/cmd/api/main.go` - The production web API server, completely independent of local disk files.
    *   `backend/cmd/seeder/main.go` - A robust, on-demand command-line seeder utility.
*   **Rationale:** Running data-loading migrations during server boot in distributed environments is dangerous (due to double-runs, locks, and start delays). Decoupling it into a CLI command gives developers and operators full control over when data is initialized or updated.

### 3. CSV Sanitization & Error Reporting
*   **Decision:** Implemented a **"Validate, Sanitize, and Report"** engine instead of a simple bulk insert.
*   **Rationale:** The challenge sample CSV contains multiple security vectors (SQL injections, XSS payloads) and malformed data (negative stock, non-numeric values, duplicates).
    *   **SQL Injection Prevention:** Parametrization using standard SQL place-markers (`?`) blocks SQL injections.
    *   **XSS Mitigation:** HTML escaping (`html.EscapeString`) handles dangerous content (e.g. `<script>` tags) safely before indexing.
    *   **Graceful Reports:** Rather than throwing general exceptions and crashing, the service imports all valid rows and returns a JSON/CLI summary of failed row numbers and their exact validation errors.

### 4. Atomic Inventory Deductions
*   **Decision:** Structured checkout purchases around explicit database transactions (`db.BeginTx`).
*   **Rationale:** To prevent double-spending or race condition stock allocations (e.g. two users buying the last item at the exact same millisecond), the stock quantity is fetched and decremented inside an isolated transaction. If database operations fail or stock falls short, changes are rolled back automatically.

### 5. Repository Pattern & Dependency Injection (Consumer-Defined Interfaces)
*   **Decision:** Replaced the global database client with a decoupled repository architecture where database interfaces are defined by the consumer packages rather than the database implementation package itself.
    *   `handlers.ProductStore` specifies the CRUD and checkout methods the HTTP handlers need.
    *   `services.ProductImporterStore` specifies the smaller subset of write/query methods the CSV importer service requires.
    *   `repository.SQLiteProductRepository` implements these queries concretely and returns a concrete pointer type (`*SQLiteProductRepository`) which implicitly satisfies both consumers.
*   **Rationale:** In Go, interfaces are satisfied implicitly (without an explicit `implements` keyword). Defining interfaces directly in the consuming packages keeps packages fully decoupled, prevents circular imports, and respects the Go proverb: *"The bigger the interface, the weaker the abstraction."* This also simplifies testing by allowing us to design minimal mock structs tailored to specific test scopes.

---

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   ├── api/          # Production HTTP server entrypoint
│   │   └── seeder/       # Standalone CLI seeder command line tool
│   ├── data/             # Raw seeding resources (products_example.csv)
│   ├── handlers/         # Products REST endpoints & purchase handlers
│   ├── models/           # Go struct representations for database rows
│   ├── repository/       # SQLite database connection setup and query implementations
│   └── services/         # CSV parser, validation logic & report engine
├── frontend/             # React SPA (Vite + TypeScript) [Upcoming]
├── docker-compose.yml    # Single-command orchestration
└── README.md             # Project documentation (this file)
```

---

## Local Setup & Execution (Backend)

Ensure you have [Go 1.22+](https://go.dev/dl/) installed.

### 1. Install Dependencies
Navigate to the backend directory and download dependencies:
```bash
cd backend
go mod tidy
```

### 2. Seed the Database
Seed the SQLite database (`ecommerce.db`) using the CSV CLI tool (no parameters needed for default run):
```bash
go run cmd/seeder/main.go
```

### 3. Start the Web Server
Launch the HTTP API server:
```bash
go run cmd/api/main.go
```
The server will run on `http://localhost:8080`.
*   Verify server health: `curl http://localhost:8080/health`

### 4. Running Automated Tests
Run handler integration tests and parser verification test suites:
```bash
go test ./... -v
```

---

## Running with Docker

Start the application inside containers:
```bash
docker compose up --build
```
This builds the multi-stage backend, starts the service, and mounts a persistent volume to preserve the database file across restarts.