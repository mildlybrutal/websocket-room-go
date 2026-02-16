package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	connections = flag.Int("conn", 1000, "number of concurrent connections")
	msgRate     = flag.Int("rate", 1, "messages per second per client")
)

func main() {
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}

	var wg sync.WaitGroup
	var activeConns int32
	var msgReceived int64

	fmt.Printf("Starting load test with %d connections...\n", *connections)

	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Unique connection URL
			connUrl := fmt.Sprintf("%s?id=bench_%d", u.String(), id)
			c, _, err := websocket.DefaultDialer.Dial(connUrl, nil)
			if err != nil {
				log.Printf("Handshake failed: %v", err)
				return
			}
			defer c.Close()
			atomic.AddInt32(&activeConns, 1)

			// Join Room
			joinMsg := map[string]string{"type": "join_room", "room": "benchmark"}
			c.WriteJSON(joinMsg)

			// Reader Routine (Count received messages)
			go func() {
				for {
					_, _, err := c.ReadMessage()
					if err != nil {
						return
					}
					atomic.AddInt64(&msgReceived, 1)
				}
			}()

			// Writer Routine (Send messages)
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			// Only send messages if we want to stress the broadcast
			for range ticker.C {
				msg := map[string]interface{}{
					"type":    "room_message",
					"room":    "benchmark",
					"content": fmt.Sprintf("bench_msg_%d", id),
				}
				if err := c.WriteJSON(msg); err != nil {
					return
				}
			}
		}(i)

		// Small delay to prevent overwhelming the OS socket limit instantly
		time.Sleep(2 * time.Millisecond)
	}

	// Monitoring Loop
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		current := atomic.LoadInt64(&msgReceived)
		atomic.StoreInt64(&msgReceived, 0) // Reset counter
		active := atomic.LoadInt32(&activeConns)
		fmt.Printf("Active: %d | Throughput: %d msgs/sec\n", active, current)
	}
}
