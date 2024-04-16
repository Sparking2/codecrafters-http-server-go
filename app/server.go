package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type request struct {
	URI string
}

type route struct {
	pattern *regexp.Regexp
	handler func(*net.Conn, request)
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler func(conn *net.Conn, request request)) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func newRequest(conn *net.Conn) request {
	req := request{
		URI: "/",
	}

	reader := bufio.NewReader(*conn)
	for {
		line, err := reader.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				req.URI = "/500"
				break
			}
		}
		line = strings.TrimSuffix(line, "\n")
		fmt.Println(line)
		if strings.Contains(line, "GET") {
			split := strings.Split(line, " ")
			if len(split) > 3 || len(split) < 2 {
				continue
			}
			req.URI = split[1]
		}

		// TODO: add data to request
		if err != nil {
			if err == io.EOF {
				req.URI = "/500"
				break
			}
		}
		if len(line) == 1 {
			break
		}
	}

	fmt.Println(req)

	return req
}

func routing(conn net.Conn, handler *RegexpHandler) {
	defer conn.Close()

	request := newRequest(&conn)

	var matched = false

	for _, route := range handler.routes {
		if !route.pattern.Match([]byte(request.URI)) {
			continue
		}
		matched = true
		route.handler(&conn, request)
		break
	}

	if !matched {
		notFoundHandler(&conn)
	}
}

func notFoundHandler(conn *net.Conn) {
	_, err := (*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	if err != nil {
		fmt.Println("Failed to write in connection", err.Error())
		return
	}
}

func okHandler(conn *net.Conn, _ request) {
	(*conn).Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func echoHandler(conn *net.Conn, request request) {

	cleanEcho := strings.Replace(request.URI, "/echo/", "", -1)
	contentLength := strconv.Itoa(len(cleanEcho))

	formatedString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %s\r\n\r\n%s", contentLength, cleanEcho)
	(*conn).Write([]byte(formatedString))
}

func main() {
	homeRegex, _ := regexp.Compile(`^/$`)
	echoRegex, _ := regexp.Compile(`/echo/.*`)

	handler := &RegexpHandler{}

	handler.Handler(homeRegex, okHandler)
	handler.Handler(echoRegex, echoHandler)

	fmt.Println("Starting Server")
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer listener.Close()

	fmt.Println("Listening on port 4221")
	for {
		conn, _ := listener.Accept()
		go routing(conn, handler)
	}

}
