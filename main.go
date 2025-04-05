package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
	http.HandleFunc("/", handleReq)
	http.HandleFunc("/admin", handleAdmin)
	http.HandleFunc("/org/{id}/admin", handleOrgAdmin)
	http.HandleFunc("/org/{id}/user", handleOrgUser)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
		return
	}

}