package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Println("Usage: client <cert.pem> <key.pem>")
		return nil
	}
	certFile := os.Args[1]
	keyFile := os.Args[2]

	certData, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("read cert: %w", err)
	}
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("read key: %w", err)
	}

	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.X509KeyPair(certData, keyData)
	if err != nil {
		return fmt.Errorf("load cert: %w", err)
	}
	tlsConfig.InsecureSkipVerify = true

	d := webtransport.Dialer{
		RoundTripper: &http3.RoundTripper{
			TLSClientConfig: tlsConfig,
		},
	}

	rsp, conn, err := d.Dial(ctx, "https://localhost/webtransport", nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	if rsp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", rsp.StatusCode)
	}

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return fmt.Errorf("accept stream: %w", err)
	}

	fmt.Println("stream:", stream)

	for {
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		fmt.Println("read:", string(buf[:n]))
	}
}
