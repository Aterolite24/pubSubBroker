package conn

import (
	"bufio"
	"log"
	"net"
)

// Client represents a connected client session.
type Client struct {
	conn     net.Conn
	id       string
	topics   map[string]bool
	incoming chan string
	outbound chan string
}

// HandleNewClient initializes a client and begins reading messages.
func HandleNewClient(conn net.Conn) {
	client := &Client{
		conn:     conn,
		id:       conn.RemoteAddr().String(),
		topics:   make(map[string]bool),
		incoming: make(chan string),
		outbound: make(chan string, 10), // buffer for outbound messages
	}

	log.Printf("New client connected: %s\n", client.id)

	go client.readLoop()
	go client.writeLoop()

	// Echo back all incoming messages for now
	for msg := range client.incoming {
		log.Printf("Received from %s: %s", client.id, msg)
		client.Send("ACK: " + msg)
	}
}

func (c *Client) readLoop() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		msg := scanner.Text()
		parsed, err := ParseMessage(msg) // ← fixed function name
		if err != nil {
			log.Printf("Invalid message from %s: %v", c.id, err)
			continue
		}
		c.handleMessage(parsed)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Client %s read error: %v", c.id, err)
	}
	close(c.incoming)
	c.conn.Close()
}

func (c *Client) writeLoop() {
	for msg := range c.outbound {
		_, err := c.conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Printf("Write error to %s: %v", c.id, err)
			break
		}
	}
	log.Printf("Stopping writer for client %s", c.id)
	c.conn.Close()
}

// Send sends a message to the client's outbound channel.
func (c *Client) Send(msg string) {
	select {
	case c.outbound <- msg:
	default:
		log.Printf("Dropping message to %s (channel full)", c.id)
	}
}

// SendJSON encodes and sends a Message as JSON.
func (c *Client) SendJSON(msg *Message) {
	encoded, err := EncodeMessage(msg)
	if err != nil {
		log.Printf("Encoding error: %v", err)
		return
	}
	c.Send(encoded)
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Action {
	case "SUB":
		c.topics[msg.Topic] = true
		c.SendJSON(&Message{
			Action:  "ACK",
			Topic:   msg.Topic,
			Payload: "Subscribed successfully",
		})
	case "PUB":
		// Just echo back to sender for now
		c.SendJSON(&Message{
			Action:  "ACK",
			Topic:   msg.Topic,
			Payload: "Received: " + msg.Payload,
		})
	default:
		c.SendJSON(&Message{
			Action:  "ERR",
			Payload: "Unknown action: " + msg.Action,
		})
	}
}