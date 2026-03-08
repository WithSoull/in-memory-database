package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
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

	fmt.Printf("Connected to %s (type EXIT to quit)\n", *address)

	stdin := bufio.NewReader(os.Stdin)
	server := bufio.NewReader(conn)

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
