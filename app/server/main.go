package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		os.Exit(1)
	}
}

func run() error {

	if len(os.Args) < 2 {
		fmt.Println("Usage: server <cert.pem> <key.pem>")
		return nil
	}
	certFile := os.Args[1]
	keyFile := os.Args[2]

	// create a new webtransport.Server, listening on (UDP) port 443
	s := webtransport.Server{
		H3: http3.Server{Addr: ":443"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Create a new HTTP endpoint /webtransport.
	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("got request")
		conn, err := s.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		clientHandler(conn)
	})

	fmt.Println("listening on 443")
	err := s.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		return err
	}

	return nil
}

func clientHandler(conn *webtransport.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("accepting")

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		fmt.Println("Failed to accept stream:", err)
		return
	}

	err = streamHandler(stream)
	if err != nil {
		fmt.Println("Failed to handle stream:", err)
		return
	}

}

func streamHandler(stream webtransport.Stream) error {
	defer stream.Close()

	fmt.Println("Handling stream", stream.StreamID())

	go func() {
		for {
			_, err := stream.Write([]byte("Hello World!"))
			if err != nil {
				fmt.Println("write: ", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		buf := make([]byte, 1024)
		_, err := stream.Read(buf)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		err = handleRequest(stream, bytes.NewReader(buf))
		if err != nil {
			return fmt.Errorf("handleRequest: %w", err)
		}
	}

}

func handleRequest(stream webtransport.Stream, r io.ReadSeeker) error {
	// read the request
	req, err := http.ReadRequest(bufio.NewReader(r))
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}

	// print the request
	fmt.Println("Received request:", req)

	_, err = stream.Write([]byte("Hello World!"))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
