package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"pubSubBroker/internal/httpserver"
)

func main() {
	r := mux.NewRouter()
	httpserver.RegisterRoutes(r)

	fmt.Println("pubSubBroker Broker is running on port 9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
