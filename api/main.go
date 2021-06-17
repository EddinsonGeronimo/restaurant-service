package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"log"
	"context"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
)

type Person struct {
	Uid  string `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
}

func main() {
	
	/*
	* chi router
	*/
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	// load data to local database
	r.Get("/customers",func(w http.ResponseWriter, r *http.Request) {
		
		/* 
		* get customers from external endpoint
		*/
		currentTime := time.Now().Format("2006-01-02")

		urlCustomers := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date=%s", currentTime)

		resp, err := http.Get(urlCustomers)
		
		if err != nil { log.Fatal(err) }
		
		defer resp.Body.Close()
		
		_, err = ioutil.ReadAll(resp.Body)//body, err := 

		/*
		* open connection to local database
		*/
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
		
		if err != nil { log.Fatal(err) }
		
		defer conn.Close()
		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		// set schema
		op := &api.Operation{
			Schema: `name: string @index(exact) .`,
		}
		
		ctx := context.Background()

		err = dgraphClient.Alter(ctx, op)

		if err != nil { log.Fatal(err) }

		txn := dgraphClient.NewTxn() // create transaction

		defer txn.Discard(ctx)

		// data to load
		p := Person{
			Uid:  "_:alice",
			Name: "Alice",
		}

		pb, err := json.Marshal(p) // json encoding of data
		
		if err != nil { log.Fatal(err) }
		
		// mutation
		mu := &api.Mutation{
			SetJson: pb,
		}

		_, err = txn.Mutate(ctx, mu)

		if err != nil { log.Fatal(err) }

		q := `query all($a: string) {
			all(func: eq(name, $a)) {
			  name
			}
		}`
		
		new_resp, err := txn.QueryWithVars(ctx, q, map[string]string{"$a": "Alice"}) // query

		err = txn.Commit(ctx)

		if err != nil { log.Fatal(err) }

		w.Write(new_resp.Json)
	})

	http.ListenAndServe(":3000", r)
}