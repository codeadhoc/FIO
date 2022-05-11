package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	log.Println("api service starting...")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hello)
	srv := http.Server{Addr: ":8000", Handler: mux}
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Println("shutting down server...")
			srv.Shutdown(ctx)
			<-ctx.Done()
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %s", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	//to simulate a deplyed response
	time.Sleep(20 * time.Second)
	w.WriteHeader(200)
	data := (time.Now()).String()
	log.Println("health ok")
	w.Write([]byte(data))
}
