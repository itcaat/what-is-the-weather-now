package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/itcaat/what-is-the-weather-now/server"
)

func main() {
	http.HandleFunc("/", server.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started on port:", port)
	http.ListenAndServe(":"+port, nil)
}
