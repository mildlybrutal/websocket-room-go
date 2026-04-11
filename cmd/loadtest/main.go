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
	connections = flag.Int("conn", 500, "number of concurrent connections")
	msgRate     = flag.Int("rate", 1, "messages per second per client")
	roomID      = flag.String("room", "1", "room ID to join (must be numeric)")
	host        = flag.String("host", "localhost:8080", "server host:port")
	duration    = flag.Int("duration", 60, "test duration in seconds")
)

type stats struct {
	activeConns int32
	failedConns int32
	msgSent     int64
	msgReceived int64
	errors      int64
}

func main() {
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *host, Path: "/ws"}

	var wg sync.WaitGroup
	var s stats

	stop := make(chan struct{})

	fmt.Printf("=== Load Test Starting ===\n")
	fmt.Printf("Target:      %s\n", u.String())
	fmt.Printf("Connections: %d\n", *connections)
	fmt.Printf("Room ID:     %s\n", *roomID)
	fmt.Printf("Msg Rate:    %d msg/sec per client\n", *msgRate)
	fmt.Printf("Duration:    %ds\n", *duration)
	fmt.Println("=========================")

	// Stop after duration
	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		close(stop)
	}()

	// Spawn clients
	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runClient(id, u, &s, stop)
		}(i)

		// Stagger connections — 2ms apart prevents OS socket exhaustion
		time.Sleep(2 * time.Millisecond)
	}

	// Stats reporter
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var prevSent, prevReceived int64

		for {
			select {
			case <-ticker.C:
				active := atomic.LoadInt32(&s.activeConns)
				failed := atomic.LoadInt32(&s.failedConns)
				sent := atomic.LoadInt64(&s.msgSent)
				received := atomic.LoadInt64(&s.msgReceived)
				errs := atomic.LoadInt64(&s.errors)

				sentRate := sent - prevSent
				recvRate := received - prevReceived
				prevSent = sent
				prevReceived = received

				fmt.Printf(
					"Active: %d | Failed: %d | Sent: %d/s | Received: %d/s | Errors: %d\n",
					active, failed, sentRate, recvRate, errs,
				)
			case <-stop:
				return
			}
		}
	}()

	wg.Wait()

	fmt.Println("\n=== Final Results ===")
	fmt.Printf("Active Connections:  %d\n", atomic.LoadInt32(&s.activeConns))
	fmt.Printf("Failed Connections:  %d\n", atomic.LoadInt32(&s.failedConns))
	fmt.Printf("Total Sent:         %d\n", atomic.LoadInt64(&s.msgSent))
	fmt.Printf("Total Received:     %d\n", atomic.LoadInt64(&s.msgReceived))
	fmt.Printf("Total Errors:       %d\n", atomic.LoadInt64(&s.errors))
}

func runClient(id int, u url.URL, s *stats, stop chan struct{}) {
	connURL := fmt.Sprintf("%s?id=bench_%d", u.String(), id)

	conn, _, err := websocket.DefaultDialer.Dial(connURL, nil)
	if err != nil {
		atomic.AddInt32(&s.failedConns, 1)
		log.Printf("[client %d] dial failed: %v", id, err)
		return
	}
	defer func() {
		atomic.AddInt32(&s.activeConns, -1)
		conn.Close()
	}()

	atomic.AddInt32(&s.activeConns, 1)

	// Join room
	err = conn.WriteJSON(map[string]string{
		"type": "join_room",
		"room": *roomID,
	})
	if err != nil {
		atomic.AddInt64(&s.errors, 1)
		log.Printf("[client %d] join failed: %v", id, err)
		return
	}

	// Reader goroutine
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			atomic.AddInt64(&s.msgReceived, 1)
		}
	}()

	// Writer — sends at configured rate
	interval := time.Second / time.Duration(*msgRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			<-readerDone
			return

		case <-readerDone:
			// Connection dropped from server side
			atomic.AddInt64(&s.errors, 1)
			return

		case <-ticker.C:
			err := conn.WriteJSON(map[string]any{
				"type":    "room_message",
				"room":    *roomID,
				"content": fmt.Sprintf("bench msg from client %d", id),
			})
			if err != nil {
				atomic.AddInt64(&s.errors, 1)
				return
			}
			atomic.AddInt64(&s.msgSent, 1)
		}
	}
}
