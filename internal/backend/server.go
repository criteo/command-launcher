package backend

import (
	"fmt"
	"net/http"
)

func StartHttpServer(port int) error {
	http.HandleFunc("/", HelloHandler)
	http.HandleFunc("/health", HealthHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}

// Implement the serve funcation in default backend
func (backend *DefaultBackend) Serve(port int) error {
	return StartHttpServer(port)
}
