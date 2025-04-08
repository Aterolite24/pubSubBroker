package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"pubSubBroker/internal/pubsub"
)

type PublishRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

var broker = pubsub.NewBroker()

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/publish", handlePublish).Methods("POST")
	r.HandleFunc("/subscribe/{topic}", handleSubscribe)
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	var msg PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	broker.Publish(msg.Topic, msg.Message)
	w.WriteHeader(http.StatusAccepted)
}

func handleSubscribe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sub := broker.Subscribe(topic)
	defer broker.Unsubscribe(topic, sub)

	for {
		select {
		case msg := <-sub:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}
