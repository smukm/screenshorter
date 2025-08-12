package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"screenshorter/internal/handlers"
	"screenshorter/internal/service"
	"screenshorter/pkg/httpserver"
	"syscall"
)

func main() {

	gz := service.NewPlaywrite()
	s := service.NewService(gz)
	h := handlers.NewHandler(s)
	srv := httpserver.NewServer()
	go func() {
		if err := srv.Run("8033", h.InitRoutes()); err != nil {
			log.Fatalf("Error running server: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")
	if err := srv.Stop(context.Background()); err != nil {
		log.Printf("Shutting down error %s", err.Error())
	}
}
