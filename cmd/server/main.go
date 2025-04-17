package main

import (
	"circledigital.in/real-state-erp/init"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	log.Println("real-estate-erp-api")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := init.GetApplication()

	log.Printf("Server is listening on PORT: %s", port)
	err = http.ListenAndServe(":"+port, app.GetRouter())
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}