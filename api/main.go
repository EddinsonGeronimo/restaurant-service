package main

import (
	_ "encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_"github.com/dgraph-io/dgo"
	_"github.com/dgraph-io/dgo/protos/api"
	_"google.golang.org/grpc"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	// Load data to local database
	r.Get("/customers",func(w http.ResponseWriter, r *http.Request) {
		
		currentTime := time.Now().Format("2006-01-02")

		urlCustomers := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date=%s", currentTime)

		resp, err := http.Get(urlCustomers)
		
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
		
		defer resp.Body.Close()
		
		body, err := ioutil.ReadAll(resp.Body)

		w.Write(body)
	})

	http.ListenAndServe(":3000", r)
}