package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Handles the connection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, _ := reader.ReadString('\n')
	parts := strings.Split(request, " ")

	for _, part := range parts {
		fmt.Println(part)
	}

	if parts[1] == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.Contains(parts[1], "/echo/") {
		var replaced string
		replaced = strings.Replace(parts[1], "/echo/", "", -1)
		fmt.Printf("Replaced: %s\n", replaced)

		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(replaced), replaced)
		fmt.Println(response)
		conn.Write([]byte(response))
	} else {
		conn.Write([]byte("HTTP/1.1 404 NOT FOUND\r\n\r\n"))
	}
}

func main() {

	fmt.Println("Listening server...")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
