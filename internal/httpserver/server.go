package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pubSubBroker/internal/models"

	"github.com/gorilla/mux"
)

var broker *models.Broker

func RegisterRoutes(r *mux.Router) {
	cfg := models.NewConfig(true, false, 5*time.Second)
	broker = models.NewBroker(cfg)

	r.HandleFunc("/publish", handlePublish).Methods("POST")
	r.HandleFunc("/subscribe/{topic}", handleSubscribe)
	r.HandleFunc("/ack", handleAck).Methods("POST")
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	var msg models.PublishData
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := broker.Publish(msg.Topic, msg.Message); err != nil {
		http.Error(w, "Broker error: "+err.Error(), http.StatusBadRequest)
		return
	}
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
			subData := models.SubscribeData{
				Topic:   topic,
				Message: msg,
			}

			data, err := json.Marshal(subData)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "%s", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

func handleAck(w http.ResponseWriter, r *http.Request) {
	var ack models.Ack

	if err := json.NewDecoder(r.Body).Decode(&ack); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := broker.Acknowledge(ack.Topic, ack.MessageID); err != nil {
		http.Error(w, "Ack error: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
