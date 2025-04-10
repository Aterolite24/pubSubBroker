package main

import (
	"fmt"
	"log"
	"net/http"

	"pubSubBroker/internal/httpserver"

	
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	httpserver.RegisterRoutes(r)

	fmt.Println("Broker is running on port 9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
