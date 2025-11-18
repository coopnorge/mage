package main

import (
	"errors"
	"log"
	"net/http"
)

// -----------------------------------------------------------------------------
// Start a mock service endpoint to act as GitHub api.
// It responds with a successful empty body for all requests; which is enough
// for policy-bot to start normally.
// -----------------------------------------------------------------------------

func startMockServer() *http.Server {
	srv := &http.Server{
		Addr: ":9090",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := w.Write([]byte(`{}`)); err != nil {
				log.Printf("mock-server write error: %v", err)
			}
		}),
	}

	go func() {
		log.Println("mock-server listening on :9090")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("mock-server failed: %v", err)
		}
	}()

	return srv
}
