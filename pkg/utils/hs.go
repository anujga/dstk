package utils

import (
	"errors"
	"go.uber.org/zap"
	"net/http"
)

func HttpServer(handlerMap map[string]http.Handler, port string) chan error {
	ch := make(chan error, 1)
	defer close(ch)
	if handlerMap == nil || len(handlerMap) == 0 {
		ch <- errors.New("no handlers")
		return ch
	}
	go func() {
		for pattern, handler := range handlerMap {
			http.Handle(pattern, handler)
		}
		err := http.ListenAndServe(port, nil)
		zap.S().Infow("stopped", "err", err)
		ch <- err
	}()
	return ch
}
