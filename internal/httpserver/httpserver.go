// Package httpserver httpserver и его хендлеры
package httpserver

import (
	"context"
	"errors"
	"log"
	"net/http"
)

type HTTPServer struct {
	*http.Server
}

func NewHTTPServer(
	handler *HTTPServerHandlers, addr string) *HTTPServer {

	server := &http.Server{
		Addr:    addr,
		Handler: Router(handler),
	}

	return &HTTPServer{server}

}

func (httpServer *HTTPServer) Run() error {

	var err error

	err = httpServer.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {

		return err
	}

	return nil

}

func (httpServer *HTTPServer) FuncShutdown() func(ctx context.Context) error {

	return func(ctx context.Context) error {

		log.Println("HTTP server stopping")
		err := httpServer.Shutdown(ctx)
		if err != nil {
			log.Printf("HTTP server stopped with error %s", err.Error())
		} else {
			log.Println("HTTP server stopped")
		}
		return err
	}
}
