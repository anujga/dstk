package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

func server() chan error {
	var err error
	ch := make(chan error, 1)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(":8080", nil)
		zap.S().Infow("stopped", "err", err)
		ch <- err
	}()
	return ch
}
