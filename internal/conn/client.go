package conn

import (
	"bufio"
	"fmt"
	"log"
	"net"
	// "strings"
)

// Client represents a connected client session.
type Client struct {
	conn     net.Conn
	id       string
	topics   map[string]bool
	incoming chan string
}

// HandleNewClient initializes a client and begins reading messages.
func HandleNewClient(conn net.Conn) {
	client := &Client{
		conn:     conn,
		id:       conn.RemoteAddr().String(),
		topics:   make(map[string]bool),
		incoming: make(chan string),
	}

	log.Printf("New client connected: %s\n", client.id)

	go client.readLoop()
	go client.writeLoop()

	// For now, simulate echo
	for msg := range client.incoming {
		log.Printf("Received from %s: %s", client.id, msg)
		client.Send("ACK: " + msg)
	}
}

func (c *Client) readLoop() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		msg := scanner.Text()
		c.incoming <- msg
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Client %s read error: %v", c.id, err)
	}
	close(c.incoming)
	c.conn.Close()
}

func (c *Client) writeLoop() {
	// Placeholder — in future we can use outgoing message queue
}

func (c *Client) Send(msg string) {
	_, err := fmt.Fprintln(c.conn, msg)
	if err != nil {
		log.Printf("Send error to %s: %v", c.id, err)
	}
}