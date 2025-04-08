package main

import (
	"log"
	"gobroker/internal/conn"
)

func main() {
	log.Println("Starting Gobroker TCP server...")
	conn.StartTCPServer(":9000")
}
