package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func server(callback handler, cf func() chan interface{}) chan error {
	var err error
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		var err error
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		req.C = cf()
		payload, err := callback(&req)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(payload))
	})
	ch := make(chan error, 1)
	go func() {
		err = http.ListenAndServe(":8080", nil)
		zap.S().Infow("stopped", "err", err)
		ch <- err
	}()
	return ch
}
