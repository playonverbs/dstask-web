package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/naggie/dstask"
)

var (
	conf    dstask.Config
	state   dstask.State
	ctx     dstask.Query
	apiFlag = flag.Bool("api", false, "enable/disable api endpoints")
)

func main() {
	flag.Parse()

	conf = dstask.NewConfig()
	state = dstask.LoadState(conf.StateFile)
	ctx = state.Context

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", IndexHandler).Methods(http.MethodGet)
	r.HandleFunc("/task/add", TaskAddHandler).Methods(http.MethodPost)
	r.HandleFunc("/task/{uuid}", TaskIndexHandler).Methods(http.MethodGet)

	if *apiFlag {
		s := r.PathPrefix("/api").
			Methods(http.MethodGet, http.MethodPost).
			Subrouter()
		s.Use(APIMiddleware)

		s.HandleFunc("/next", APINextHandler)
		s.HandleFunc("/task", APINextHandler)
		s.HandleFunc("/task/{id}", APITaskHandler)
		s.HandleFunc("/add", APIAddHandler)
	}

	srv := &http.Server{
		Addr:         "0.0.0.0:1313",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		log.Println("listening on", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

func APIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(rw, r)
	})
}
