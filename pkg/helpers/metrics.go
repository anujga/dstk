package helpers

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

func ExposePrometheus(address string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			zap.S().Errorw("error in prometheus endpoint",
				"err", err)
		}
	}()
	return server

}
