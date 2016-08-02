package marbl

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/google/martian/log"

	"golang.org/x/net/websocket"
)

type Handler struct {
	mu   sync.RWMutex
	subs map[string]chan<- []byte
}

func NewHandler() *Handler {
	return &Handler{
		subs: make(map[string]chan<- []byte),
	}
}

func (h *Handler) Write(b []byte) (int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, framec := range h.subs {
		go func(framec chan<- []byte) {
			framec <- b
		}(framec)
	}

	return len(b), nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	websocket.Server{Handler: h.streamLogs}.ServeHTTP(rw, req)
}

func (h *Handler) streamLogs(conn *websocket.Conn) {
	defer conn.Close()

	id, err := newID()
	if err != nil {
		log.Errorf("logstream: failed to create ID: %v", err)
		return
	}
	framec := make(chan []byte, 10)

	h.subscribe(id, framec)
	defer h.unsubscribe(id)

	for b := range framec {
		if err := websocket.Message.Send(conn, b); err != nil {
			log.Errorf("logstream: failed to send message: %v", err)
			return
		}
	}
}

func newID() (string, error) {
	src := make([]byte, 8)
	if _, err := rand.Read(src); err != nil {
		return "", err
	}

	return hex.EncodeToString(src), nil
}

func (h *Handler) unsubscribe(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.subs, id)
}

func (h *Handler) subscribe(id string, framec chan<- []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.subs[id] = framec
}
