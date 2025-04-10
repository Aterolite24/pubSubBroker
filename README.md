# PubSubBroker

## Project Overview

This project implements a simple publish-subscribe message broker. It allows clients to publish messages to specific topics, and subscribers can receive these messages in real-time. The broker supports at-least-once delivery and persists messages using MongoDB.

## Features

* **Publishing:** Clients can publish messages to topics via an HTTP POST request.
* **Subscribing:** Clients can subscribe to topics and receive messages via Server-Sent Events (SSE).
* **At-Least-Once Delivery:** Guarantees that messages are delivered at least once to subscribers.
* **Message Acknowledgment:** Subscribers can acknowledge received messages.
* **Message Persistence:** Messages are stored in MongoDB.
* **Historical Message Retrieval:** Subscribers can receive messages published after a specific timestamp.
* **Automatic Message Deletion:** Old messages are automatically deleted from the database.

## File Descriptions

* **`/cmd/main.go`:** The main entry point of the application. Sets up the HTTP router and starts the server.
* **`/httpserver/server.go`:** Defines the HTTP server logic, including route registration and handlers for publishing, subscribing, and acknowledging messages.
* **`/models/ack.go`:** Defines the structure for message acknowledgments.
* **`/models/broker.go`:** Contains the core logic of the message broker, managing subscriptions, publishing, and acknowledgments.
* **`/models/clientData.go`:** Defines the structure for client subscription requests.
* **`/models/config.go`:** Defines the configuration parameters for the broker.
* **`/models/dbMessage.go`:** Defines the structure for messages stored in the database.
* **`/models/envelope.go`:** Defines a wrapper for messages used in at-least-once delivery.
* **`/models/error.go`:** Defines a custom error type for the broker.
* **`/models/pendingMessage.go`:** Defines the structure for tracking messages awaiting acknowledgment.
* **`/models/publishData.go`:** Defines the structure for publishing messages.
* **`/models/subscribeData.go`:** (Appears unused)
* **`/models/subscribeRequest.go`:** Defines the structure for subscribe requests in the test client.
* **`/utils/db.go`:** Contains utility functions for database interactions (MongoDB).
* **`/utils/generate.go`:** Contains utility functions for generating unique IDs.
* **`/tests/client.go`:** A simple Go client for testing the broker.
* **`/.env.sample`:** An example file for environment variable configuration.

## Architecture

### Layer Breakdown

The PubSubBroker follows a layered architecture, separating concerns into distinct components:

#### 1. Presentation Layer (HTTP Interface)

* This layer is responsible for handling communication with clients via HTTP.
* It uses the `github.com/gorilla/mux` router to map incoming HTTP requests to specific handler functions defined in `httpserver/server.go`.
* **`RegisterRoutes`**: Sets up the API endpoints (`/publish`, `/subscribe`, `/ack`) and associates them with their respective handler functions.
* **`handlePublish`**: Receives HTTP POST requests to `/publish`, decodes the JSON payload containing the topic and message, and calls the `Publish` method in the Business Logic Layer.
* **`handleSubscribe`**: Handles HTTP POST requests to `/subscribe`, decodes the JSON payload for the topic and optional timestamp, sets up the Server-Sent Events (SSE) connection, and calls the `Subscribe` method in the Business Logic Layer to register the subscriber and retrieve messages. It then streams messages to the client.
* **`handleAck`**: Handles HTTP POST requests to `/ack`, decodes the JSON payload containing the message ID and topic, and calls the `Acknowledge` method in the Business Logic Layer to process the acknowledgment.

#### 2. Business Logic Layer (Core Broker Functionality)

* This layer contains the core logic of the publish-subscribe system, implemented in `models/broker.go`.
* The `Broker` struct manages the broker's state, including:
    * **`subs`**: A map storing active subscriptions, where keys are topics and values are slices of channels (`chan json.RawMessage`) for each subscriber.
    * **`config`**: Holds the broker's configuration parameters (e.g., timeouts, intervals).
    * **`pendingAcks`**: A map used for at-least-once delivery, tracking messages that have been sent but not yet acknowledged.
* **`Publish`**: Receives a topic and message, persists the message using the Data Access Layer, and then distributes the message to all subscribers of that topic. If at-least-once delivery is enabled, it wraps the message in an `Envelope` and starts an acknowledgment timer.
* **`Subscribe`**: Registers a new subscriber for a given topic by creating a channel and adding it to the `subs` map. It also retrieves historical messages from the Data Access Layer based on the provided timestamp and sends them to the new subscriber.
* **`Acknowledge`**: Processes acknowledgments from subscribers by stopping the acknowledgment timer and removing the message from the `pendingAcks` map.
* This layer utilizes Go's concurrency features (goroutines and channels) to handle multiple publishers and subscribers efficiently.

#### 3. Data Access Layer (Persistence)

* This layer is responsible for interacting with the MongoDB database and is implemented in `utils/db.go`.
* **`GetMongoClient`**: Establishes and returns a singleton instance of the MongoDB client.
* **`insertDatabase`**: Persists a published message into the MongoDB database.
* **`Find`**: Retrieves messages from MongoDB based on specified criteria (e.g., topic and timestamp for historical retrieval).
* **`DeleteRowsOlderThan`** and **`StartRowDeletionJob`**: Implement the logic for periodically deleting old messages from the database based on the configured persistence timeout.

#### 4. Utilities Layer

* The `utils` package provides helper functions used across the application.
* **`GenerateMessageID`**: Generates unique identifiers for messages, which are essential for the at-least-once delivery mechanism.

### Data Flow

* **Publishing:** A client sends a POST request to `/publish`. The Presentation Layer's `handlePublish` decodes the request and calls the Business Logic Layer's `Publish` method. `Publish` then uses the Data Access Layer to persist the message and sends it to relevant subscribers (potentially wrapped in an `Envelope` with an acknowledgment timer).
* **Subscribing:** A client sends a POST request to `/subscribe`. The Presentation Layer's `handleSubscribe` sets up the SSE connection and calls the Business Logic Layer's `Subscribe` method. `Subscribe` creates a subscriber channel, retrieves historical messages from the Data Access Layer, and adds the channel to its subscription registry. Messages are then streamed to the client via the SSE connection.
* **Acknowledging:** A client sends a POST request to `/ack`. The Presentation Layer's `handleAck` decodes the request and calls the Business Logic Layer's `Acknowledge` method, which updates the state of pending acknowledgments.

### Concurrency Model

The broker leverages Go's concurrency model:

* **Goroutines:** Used to handle multiple client requests concurrently, for background tasks like message deletion, and for managing acknowledgment timers.
* **Channels:** Used for communication between different parts of the system, particularly for sending messages from the `Publish` method to the `handleSubscribe` handler for streaming to clients.

### State Management

The broker maintains its state primarily in memory within the `Broker` struct (e.g., active subscriptions in the `subs` map and pending acknowledgments in the `pendingAcks` map). MongoDB is used for persistent storage of messages.

### Technology Choices

* **Go:** Chosen for its strong concurrency support, performance, and suitability for building network applications.
* **Gorilla Mux:** A popular and powerful HTTP router for Go, providing flexibility in defining API endpoints.
* **MongoDB:** A NoSQL database chosen for its flexible schema and scalability, suitable for storing unstructured message data.

## How to Test

1.  **Prerequisites:**
    * Go installed
    * MongoDB running
    * Set up environment variables in a `.env` file (copy from `.env.sample`).

2.  **Run the Broker:**
    ```bash
    go run ./cmd/main.go
    ```

3.  **Run the Test Client:**
    Open another terminal and navigate to the `/tests` directory:
    ```bash
    go run client.go
    ```

4.  **Publish Messages (using `curl` in a new terminal):**
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"topic": "test-topic", "message": {"value": 123}}' http://localhost:9000/publish
    ```
    Observe the output in the test client terminal.

## Potential Future Work

* Implement Acknowledge pooling for handling pending messages.
* Implement synchronization with multiple brokers.
* Explore other message queueing technologies.