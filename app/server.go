package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func errorHandler(w http.ResponseWriter, _ *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		_, err := fmt.Fprint(w, "404 - Not Found")
		if err != nil {
			return
		}
	}
}

func homeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			errorHandler(w, r, http.StatusNotFound)
			return
		}
		_, err := fmt.Fprintf(w, "Hello, World!")
		if err != nil {
			return
		}
	})
}

func echoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanURI := strings.Replace(r.RequestURI, "/echo/", "", -1)
		contentLength := strconv.Itoa(len(cleanURI))
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", contentLength)
		_, err := fmt.Fprintf(w, cleanURI)
		if err != nil {
			return
		}
	})
}

func agentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent := r.Header.Get("User-Agent")
		contentLength := strconv.Itoa(len(agent))

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", contentLength)
		_, err := fmt.Fprintf(w, agent)
		if err != nil {
			return
		}
	})
}

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(w http.ResponseWriter, r *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func main() {
	handler := &RegexpHandler{}

	echoRegex, _ := regexp.Compile(`/echo/.*`)
	handler.Handler(echoRegex, echoHandler())

	userAgentRegex, _ := regexp.Compile("user-agent")
	handler.Handler(userAgentRegex, agentHandler())

	homeRegex, _ := regexp.Compile("/")
	handler.Handler(homeRegex, homeHandler())

	http.ListenAndServe("0.0.0.0:4221", handler)
}
