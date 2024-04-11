package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func responseRoot(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		fmt.Println("Failed to accept connection", err.Error())
		return
	}
}

func responseUserAgent(conn net.Conn) {
	fmt.Println("user-agent found")
	var userAgent string = "temp-agent\n"
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to accept connection", err.Error())
		return
	}
}

func responseEcho(conn net.Conn, path string) {
	var replaced string
	replaced = strings.Replace(path, "/responseEcho/", "", -1)
	fmt.Printf("Replaced: %s\n", replaced)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(replaced), replaced)
	fmt.Println(response)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to accept connection", err.Error())
		return
	}
}

func responseNotFound(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	if err != nil {
		fmt.Println("Failed to accept connection", err.Error())
		return
	}
}

// HttpRequest struct for grouping together information about the request
type HttpRequest struct {
	method string
	path   string
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection: ", err.Error())
		}
	}(conn)

	reader := bufio.NewReader(conn)
	request, _ := reader.ReadString('\n')

	requestInfo := strings.Split(request, " ")

	httpRequest := HttpRequest{method: requestInfo[0], path: requestInfo[1]}
	fmt.Println(httpRequest)

	if httpRequest.path == "/" {
		responseRoot(conn)
	} else if strings.Contains(httpRequest.path, "/echo/") {
		responseEcho(conn, httpRequest.path)
	} else if strings.Contains(httpRequest.path, "user-agent") {
		responseUserAgent(conn)
	} else {
		responseNotFound(conn)
	}
}

func main() {
	fmt.Println("Starting server...")
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	fmt.Println("Listening on port 4221")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
