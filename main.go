package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/mozafar/hello-microservice/data"
	"github.com/mozafar/hello-microservice/handlers"
)

func main() {

	l := log.New(os.Stdout, "hello-service ", log.LstdFlags)
	v := data.NewValidation()

	// Create the handlers
	productsHanler := handlers.NewProducts(l, v)

	// Create router and handlers
	router := mux.NewRouter()

	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", productsHanler.ListAll)
	getRouter.HandleFunc("/products/{id:[0-9]+}", productsHanler.ListSingle)

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/products", productsHanler.Create)
	postRouter.Use(productsHanler.MiddlewareValidateProduct)

	putRouter := router.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/products/{id:[0-9]+}", productsHanler.Update)
	putRouter.Use(productsHanler.MiddlewareValidateProduct)

	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)

	getRouter.Handle("/docs", sh)
	getRouter.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	// Create a server
	server := http.Server{
		Addr:         "127.0.0.1:9090",
		Handler:      router,
		ErrorLog:     l,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		l.Printf("Starting server on %s\n", server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			l.Fatal(err)
			os.Exit(1)
		}
	}()

	osSignalChannel := make(chan os.Signal, 1)
	signal.Notify(osSignalChannel, os.Interrupt)

	osSignal := <-osSignalChannel

	l.Printf("Received: %s,Gracefully shutting down the server!\n", osSignal)

	ct, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ct)
}
