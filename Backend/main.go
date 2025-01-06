package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/config"
	"backend/middleware"
	"backend/routes"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Import routes

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// App config
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	// Database connection
	config.ConnectDB()

	// Initialize the router
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.JSONMiddleware)
	router.Use(middleware.CorsMiddleware)

	// API Endpoints
	router.Handle("/api/food", routes.FoodRoutes()).Methods("GET", "POST")
	router.Handle("/api/user", routes.UserRoutes()).Methods("GET", "POST")
	router.Handle("/api/cart", routes.CartRoutes()).Methods("GET", "POST")
	router.Handle("/api/order", routes.OrderRoutes()).Methods("GET", "POST")

	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./uploads"))))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Working"))
	})

	fmt.Printf("Server Started on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
