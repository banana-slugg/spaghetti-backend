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
	sl   SpaghettiLevel
	pass string
}

func main() {
	fmt.Println("Starting server...")

	spagLevel := &spaghettiHandler{sl: SpaghettiLevel{Level: 1}, pass: os.Getenv("SPAG_PASS")}

	if spagLevel.pass == "" {
		panic("Needs password")
	}

	http.HandleFunc("/spaghetti", spagLevel.serve)

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func (s *spaghettiHandler) serve(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	switch r.Method {
	case "GET":
		s.get(w, r)
		return
	case "POST":
		if !ok || username != "Spagett" || password != s.pass {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("hey, scram!"))
			return
		} else {
			s.post(w, r)
			return
		}
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
		w.Write([]byte(err.Error()))
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
		w.Write([]byte(err.Error()))
		return
	}

	s.Lock()
	s.sl.Level = spag.Level
	defer s.Unlock()
}
