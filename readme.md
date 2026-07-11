# E-Commerce Application

E-Commerce platform built with **Go** and **React (Vite)**. 
It uses a **Single-Binary Deployment** model, meaning the frontend UI is baked directly into the backend server for a frictionless single-container deployment!

> The example CSV file was downloaded on **2026-07-09**.

---

## 🚀 How to Execute (Start to End)

The easiest way to run the entire application (Database, API, and Frontend UI) is using Docker.

### 1. Start the System
Ensure you have Docker installed, then run:
```bash
docker compose up --build
```
This single command will:
1. Build the React frontend.
2. Compile the Go backend (embedding the React UI inside it).
3. Start the server at `http://localhost:8080/`.

### 2. Navigate the Application
Open your browser and navigate to **http://localhost:8080/**. The frontend emulates a clean separation of concern via three tabs:

#### 🛍️ Shop View (Customer)
- **Browse & Search:** Search the product catalog dynamically using the search input in the header.
- **Purchase Products:** Use the quantity steppers (`+` / `-`) on any card and click **Buy Now** to trigger checkout (processed instantly via a mock billing broker using concurrency-safe optimistic locking).

#### 🗂️ Inventory View (Administrator)
- **Bulk CSV Import:** Click **Upload CSV** in the header and select `backend/data/products_example.csv` to populate the database.
- **Manual Product CRUD:**
  - Click **+ Add Product** to create a product. The price field uses a **currency mask** (typing digits automatically formats cents, e.g. `4599` displays as `45.99`, so you don't need to type `.`).
  - Validation errors (e.g. negative price, invalid SKU characters, blank name) display directly as **inline custom warnings** below each input.
  - Click **Edit** (✏️) or **Delete** (🗑️) on any product card to update details or purge products.

#### 📜 Order History View (Ledger/Audit)
- Inspect all simulated transaction records (date, order UUID, customer token, SKU, quantity, and total price).

---

## 🧪 Running Tests

The application has unit and integration test suites for both the backend (Go) and the frontend (React).

### Backend Tests (Go)
To run the Go tests, navigate to the `backend` folder and run the standard Go test command:
```bash
cd backend
go test ./...
```

### Frontend Tests (Vitest & React Testing Library)
To run the frontend tests, navigate to the `frontend` folder. You can run them once or in watch mode:
*   **Run once:**
    ```bash
    cd frontend
    npm run test:run
    ```
*   **Watch mode (interactive):**
    ```bash
    cd frontend
    npm run test
    ```

---

## 🏛️ Architectural Decisions

### 1. HTTP Router: Standard Library `ServeMux` (Go 1.22+)
**Why Mux?**
We rely entirely on Go's built-in `http.ServeMux` rather than third-party frameworks like Gin or Fiber.
*   **Reason:** Go 1.22 introduced native path parameter routing (e.g., `PUT /api/products/{id}`). This eliminated the need for bloated external libraries, resulting in faster compilation, zero router dependencies, and highly idiomatic code.

### 2. Safe Checkouts: Optimistic Locking & Retries
**Why Retries?**
When multiple users try to buy the exact same product at the exact same millisecond, race conditions can occur (double-spending stock). 
*   **Reason:** We implemented **Optimistic Locking** using a `version` column in SQLite. During checkout, if a version conflict is detected (meaning someone else just bought it), the `OrderService` automatically catches the error, waits 50 milliseconds, and **Retries** the transaction. This guarantees 100% stock accuracy without heavy table locks.

### 3. Idempotency Keys
**Why Idempotency?**
If a user clicks "Buy Now" and their WiFi drops before they get the success message, they might click it again and get charged twice.
*   **Reason:** The frontend generates a unique `idempotency_key` (UUID) for every purchase attempt. The backend checks this key; if it has seen it before, it safely returns the exact same successful order response without charging the card again or deducting stock.

### 4. Single-Binary Embedded React
**Why Embed?**
*   **Reason:** We utilize Go's `//go:embed` directive. The Dockerfile builds the React frontend and deposits it into the Go source tree. The Go compiler then bakes the HTML/CSS/JS directly into the binary. This eliminates the need for an Nginx proxy or a complex two-container Docker setup.

---

## 🔌 API Endpoints

The backend exposes a highly specialized, minimalist JSON API:

| Method | Endpoint | Description |
| **GET** | `/api/products` | Lists the catalog. Supports `?q=` (search) and `?category=` filters. |
| **POST** | `/api/products/import` | Accepts a Multipart Form (CSV file) to bulk upsert products. |
| **POST** | `/api/products` | Creates a single product manually. |
| **PUT** | `/api/products/{id}` | Updates an existing product. |
| **DELETE** | `/api/products/{id}` | Deletes a product from the catalog. |
| **POST** | `/api/purchase` | Simulates a checkout. Requires `sku`, `quantity`, `expected_price`, `customer_id`, and `idempotency_key`. |
| **GET** | `/api/orders` | Retrieves the global order history. |

---

## 🛡️ Security Decisions

### 1. Cross-Site Scripting (XSS) Prevention
If a user imports a CSV or uses the API to upload a product named `<script>alert('xss')</script>`, how is it handled?
*   **Reasoning:** The Go backend accepts and stores the raw, unescaped string directly in the database. We explicitly **avoid** running `html.EscapeString` on the backend to prevent "double escaping" bugs. Instead, we rely entirely on **React's native JSX sanitization**. Whenever React binds dynamic data (e.g., `<p>{product.name}</p>`), it safely encodes HTML entities on the fly, rendering the exact text safely without executing malicious scripts.

### 2. SQL Injection Prevention
*   **Reasoning:** The application interacts with SQLite exclusively through the `database/sql` standard library using strict **Parameterized Queries**. By passing values as independent arguments to placeholders (`?`), the database driver guarantees that user input is treated as literal data, making SQL injection mathematically impossible.

---

## 🖥️ Frontend Modules (React)

The frontend is a lightweight Single Page Application (SPA) utilizing Pure Vanilla CSS for premium styling. It is deeply segregated into reusable components:

1. **State Management (`hooks/useAppStore.ts`)**
   - A centralized custom hook that handles API interactions, state tracking, debounce-searching, and background background loading (to prevent UI blinking).
2. **API Integration Layer (`services/api.ts`)**
   - A minimalist, zero-dependency `fetch` wrapper strictly typed to match the backend JSON schemas.
3. **Component Architecture (`components/`)**
   - `Header.tsx`: Contains the navigation tabs, robust search bar, and CSV bulk importer.
   - `Catalog.tsx` & `ProductCard.tsx`: Renders the grid of products with quantity steppers and purchase buttons.
   - `OrdersTable.tsx`: Displays the transaction ledger.
   - `ProductModal.tsx`: The overlay form for manual CRUD operations.