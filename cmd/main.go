package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sl "github.com/mashalin/multiplexer/internal/lib/logger"
	"github.com/mashalin/multiplexer/internal/transport/rest"
	urlshandler "github.com/mashalin/multiplexer/internal/urls/http-server"
	urlsservice "github.com/mashalin/multiplexer/internal/urls/service"
)

const ServerPort = ":8082"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	urlsService := urlsservice.New()
	urlsHandler := urlshandler.New(urlsService)

	server := &http.Server{
		Addr:    ServerPort,
		Handler: http.HandlerFunc(rest.LimitMiddleware(urlsHandler.Handle)),
	}

	log.Info("initializing server", slog.String("PORT", ServerPort))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error("failed to start server", sl.Err(err))
		}
	}()

	log.Info("server started")

	<-shutdown
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}
