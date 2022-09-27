package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type SpaghettiLevel struct {
	Level int `json:"level"`
}

type spaghettiHandler struct {
	sync.Mutex
	sl     SpaghettiLevel
	pass   string
	stream chan int
}

func main() {
	fmt.Println("Starting server...")

	spagLevel := &spaghettiHandler{
		sl:     SpaghettiLevel{Level: 1},
		pass:   os.Getenv("SPAG_PASS"),
		stream: make(chan int),
	}

	if spagLevel.pass == "" {
		panic("Needs password")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", spagLevel.serve)               // REST
	mux.HandleFunc("/stream", spagLevel.streamHandler) // server sent events
	mux.HandleFunc("/auth", spagLevel.auth)            // "login"

	log.Fatal(http.ListenAndServe(getPort(), mux))
}

func (s *spaghettiHandler) auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "authorization")

	username, password, ok := r.BasicAuth()

	switch r.Method {
	case "GET":
		if !ok || username != "Spagett" || password != s.pass {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			return
		}
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func (s *spaghettiHandler) streamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// defer func() {
	// 	if s.stream != nil {
	// 		close(s.stream)
	// 		s.stream = nil
	// 	}
	// }()

	flush, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for {
		select {
		case message := <-s.stream:
			if s.stream != nil {
				fmt.Fprintf(w, "data: %v\n\n", message)
				log.Printf("MESSAGE DISPATCHED: %v", message)
			}
			flush.Flush()

		case <-r.Context().Done():
			log.Println("CONNECTION CLOSED")
			return
		}
	}
}

func (s *spaghettiHandler) serve(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")

	switch r.Method {
	case "GET":
		s.get(w, r)
		return
	case "POST":
		if !ok || username != "Spagett" || password != s.pass {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("hey, scram!"))
			return
		}
		s.post(w, r)

	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("Unfortunately, I am a teapot"))
		return
	}
}

func (s *spaghettiHandler) get(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	jsonBytes, err := json.Marshal(s.sl)
	s.Unlock()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (s *spaghettiHandler) post(w http.ResponseWriter, r *http.Request) {

	content := r.Header.Get("content-type")
	if content != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var spag SpaghettiLevel

	err := json.NewDecoder(r.Body).Decode(&spag)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.Lock()
	if spag.Level > 10 || spag.Level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.sl.Level = spag.Level
	s.Unlock()
	s.stream <- s.sl.Level

}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}
