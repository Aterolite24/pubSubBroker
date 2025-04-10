package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pubSubBroker/internal/models"
	"pubSubBroker/internal/utils"

	"github.com/gorilla/mux"
)

var broker *models.Broker

func RegisterRoutes(r *mux.Router) {
	cfg := models.NewConfig(true, false, 10*time.Second, 5*time.Second, 15*time.Second)

	broker = models.NewBroker(cfg)
	utils.StartRowDeletionJob(cfg.DeletionInterval, cfg.PersistenceTimeout)

	r.HandleFunc("/publish", handlePublish).Methods("POST")
	r.HandleFunc("/subscribe", handleSubscribe).Methods("POST")
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

	var body models.ClientData
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	topic := body.Topic

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sub := broker.Subscribe(topic, body.TimeStamp)
	defer broker.Unsubscribe(topic, sub)

	for {
		select {
		case msg := <-sub:
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "%s\n", data)
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
