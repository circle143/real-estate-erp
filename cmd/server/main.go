package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
)

func handleReq(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "hello world!")
	if err != nil {
		log.Println("error sending response")
		return
	}
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "circle admin page!")
	if err != nil {
		log.Println("error sending response")
		return
	}
}

func handleOrgAdmin(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	responseString := fmt.Sprintf("org: %v,\nAdmin page!", idString)

	_, err := io.WriteString(w, responseString)
	if err != nil {
		log.Println("error sending response")
		return
	}
}

func handleOrgUser(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	responseString := fmt.Sprintf("org: %v,\nUser page!", idString)

	_, err := io.WriteString(w, responseString)
	if err != nil {
		log.Println("error sending response")
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleReq)
	http.HandleFunc("/admin", handleAdmin)
	http.HandleFunc("/org/{id}/admin", handleOrgAdmin)
	http.HandleFunc("/org/{id}/user", handleOrgUser)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
		return
	}

}
