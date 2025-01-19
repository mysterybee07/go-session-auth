package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mysterybee07/go-sessions-auth/routes"
)

func main() {

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello World")
	// })
	routes.Setup()
	fmt.Println("Server running on port: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
