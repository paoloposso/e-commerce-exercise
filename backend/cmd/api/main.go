package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"ntd/backend/handlers"
	"ntd/backend/payment"
	"ntd/backend/repository"
	"ntd/backend/services"
)

//go:embed all:dist
var frontendAssets embed.FS

// corsMiddleware enables cross-origin requests from local development clients.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/ecommerce.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbHandle, err := repository.ConnectDB(dbPath)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbHandle.Close()

	productRepo := repository.NewSQLiteProductRepository(dbHandle)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	orderService := services.NewOrderService(productRepo, &payment.MockBroker{})
	orderHandler := handlers.NewOrderHandler(orderService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	mux.HandleFunc("GET /api/products", productHandler.ListAndSearchProducts)
	mux.HandleFunc("POST /api/products", productHandler.CreateProduct)
	mux.HandleFunc("PUT /api/products/{id}", productHandler.UpdateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", productHandler.DeleteProduct)
	mux.HandleFunc("POST /api/products/import", productHandler.ImportProducts)
	mux.HandleFunc("POST /api/purchase", orderHandler.PurchaseProduct)

	distFS, err := fs.Sub(frontendAssets, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if _, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}

			content, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(content)
		})
	} else {
		log.Printf("Warning: Embedded frontend dist folder not found: %v", err)
	}

	log.Printf("Server listening on port %s", port)
	err = http.ListenAndServe(":"+port, corsMiddleware(mux))
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
