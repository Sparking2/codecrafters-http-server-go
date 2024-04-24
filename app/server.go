package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type request struct {
	URI    string
	Agent  string
	Method string
	Body   string
}

type route struct {
	pattern *regexp.Regexp
	handler func(*net.Conn, request)
	method  string
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler func(conn *net.Conn, request request), method string) {
	h.routes = append(h.routes, &route{pattern, handler, method})
}

func newRequest(conn *net.Conn) request {
	req := request{
		URI: "/",
	}

	reader := bufio.NewReader(*conn)
	buffer := make([]byte, 4096)
	read, err := reader.Read(buffer)
	if err != nil {
		fmt.Println(err.Error())
		req.URI = "/404"
		return req
	}
	requestData := string(buffer[:read])

	splitRequestedData := strings.Split(requestData, "\n")
	readyToReadBody := false

	for _, s := range splitRequestedData {
		line := strings.TrimSuffix(s, "\n")
		//fmt.Println(line)
		if strings.Contains(line, "GET") {
			split := strings.Split(line, " ")
			if len(split) > 3 || len(split) < 2 {
				continue
			}
			req.URI = split[1]
			req.Method = "GET"
		}
		if strings.Contains(line, "POST") {
			split := strings.Split(line, " ")
			if len(split) > 3 || len(split) < 2 {
				continue
			}
			req.URI = split[1]
			req.Method = "POST"
		}

		if strings.Contains(line, "User-Agent:") {
			req.Agent = strings.Replace(line, "User-Agent: ", "", -1)
		}

		if len(line) == 1 {
			readyToReadBody = true
			continue
		}

		if readyToReadBody {
			req.Body = line
		}
	}

	//fmt.Println(req)

	return req
}

func routing(conn net.Conn, handler *RegexpHandler) {
	//defer conn.Close()

	request := newRequest(&conn)

	var matched = false

	for _, route := range handler.routes {
		if !route.pattern.Match([]byte(request.URI)) {
			continue
		}
		if route.method != request.Method {
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

func agentHandler(conn *net.Conn, request request) {
	contentLength := strconv.Itoa(len(request.Agent) - 1)
	formatedString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %s\r\n\r\n%s", contentLength, request.Agent)
	(*conn).Write([]byte(formatedString))
}

func filesHandler(conn *net.Conn, request request) {
	fmt.Printf("Input directory: %s\n", *directoryPtr)

	cleanUri := strings.Replace(request.URI, "/files/", "", -1)

	data, err := os.ReadFile(*directoryPtr + cleanUri)
	if err != nil {
		notFoundHandler(conn)
		return
	}
	readContent := string(data)

	contentLength := strconv.Itoa(len(readContent))

	formatedString := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %s\r\n\r\n%s", contentLength, readContent)
	(*conn).Write([]byte(formatedString))
}

func fileCreation(conn *net.Conn, request request) {
	fmt.Printf("Output directory: %s\n", *directoryPtr)

	cleanUri := strings.Replace(request.URI, "/files/", "", -1)

	err := os.WriteFile(*directoryPtr+cleanUri, []byte(request.Body+"\r\n"), 0644)
	if err != nil {
		fmt.Println("Failed to write file", err.Error())
	}

	formatedString := fmt.Sprintf("HTTP/1.1 201 Created\r\nContent-Type: text/plain\nContent-Length: 0\r\n")
	(*conn).Write([]byte(formatedString))
}

var directoryPtr *string

func main() {
	directoryPtr = flag.String("directory", ".", "Directory to serve files from")
	flag.Parse()

	homeRegex, _ := regexp.Compile(`^/$`)
	echoRegex, _ := regexp.Compile(`/echo/.*`)
	agentRegex, _ := regexp.Compile("user-agent")
	fileRegex, _ := regexp.Compile(`/files/.*`)

	handler := &RegexpHandler{}

	handler.Handler(homeRegex, okHandler, "GET")
	handler.Handler(echoRegex, echoHandler, "GET")
	handler.Handler(agentRegex, agentHandler, "GET")
	handler.Handler(fileRegex, filesHandler, "GET")
	handler.Handler(fileRegex, fileCreation, "POST")

	fmt.Println("Starting Server")
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	fmt.Println("Server started at http://localhost:4221")
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

// local test: curl -vvv -d "hello world" localhost:4221/files/readme.txt
