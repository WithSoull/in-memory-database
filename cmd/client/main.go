package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const defaultAddress = "127.0.0.1:3223"

func main() {
	address := flag.String("address", defaultAddress, "server TCP address")
	flag.Parse()

	conn, err := net.Dial("tcp", *address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	stdin := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

	// Verify the connection is alive and the server has capacity.
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprintln(conn, "PING"); err != nil {
		fmt.Fprintf(os.Stderr, "server unavailable: %v\n", err)
		os.Exit(1)
	}
	pong, err := server.ReadString('\n')
	conn.SetDeadline(time.Time{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "server rejected connection: connection limit reached")
		os.Exit(1)
	}
	if strings.TrimSpace(pong) != "PONG" {
		fmt.Fprintf(os.Stderr, "unexpected server response: %q\n", strings.TrimSpace(pong))
		os.Exit(1)
	}

	fmt.Printf("Connected to %s (type EXIT to quit)\n", *address)

	for {
		fmt.Print("> ")

		line, err := stdin.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "EXIT") {
			fmt.Println("Bye!")
			break
		}

		if _, err := fmt.Fprintln(conn, line); err != nil {
			fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			break
		}

		response, err := server.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "receive error: %v\n", err)
			break
		}

		fmt.Print(response)
	}
}
