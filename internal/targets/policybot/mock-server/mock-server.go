package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/api/app", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("{}"))
		if err != nil {
			return
		}
	})

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		return
	}
}
