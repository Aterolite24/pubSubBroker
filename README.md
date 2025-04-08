# 🚀 Persistent Pub–Sub Message Broker

A lightweight, high-performance, and configurable message broker inspired by NATS, supporting multiple delivery guarantees and retention policies.

---

## 📌 Project Overview

This system implements a **publish–subscribe** messaging model where:

- Publishers send messages to named **topics**
- Subscribers receive messages from the topics they subscribe to
- The broker is **persistent**, supports **configurable delivery guarantees**, and **retention policies**

---

## ✨ Features

- [x] Pub–Sub Channel Decoupling
- [x] Max-once Delivery Guarantee
- [ ] At-least-once Delivery Guarantee
- [ ] Message Persistence (In-memory + Disk)
- [ ] Time-based & Capability-based Retention
- [ ] Topic Routing & Filtering
- [ ] Scalable Design for Clustering
- [ ] Simple API Protocol (JSON / Protobuf)
- [ ] Real-time Logging & Monitoring

---

## 📁 Project Structure

```bash
.
├── broker/                 # Core pub-sub broker logic
│   ├── connection/         # Client connection handling
│   ├── delivery/           # Delivery guarantees (acks, retries)
│   ├── retention/          # TTL and capability-based logic
│   ├── routing/            # Topic-based message routing
│   └── storage/            # Persistence layer (WAL or DB)
├── api/                    # External API layer (TCP/WebSocket)
├── cli/                    # CLI tools (broker control, status)
├── config/                 # Configuration files and loaders
├── scripts/                # Dev tools, setup scripts
├── tests/                  # Unit and integration tests
├── docs/                   # Diagrams, specs, design notes
├── .gitignore
├── README.md
└── requirements.txt or go.mod
