package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	_"regexp"
	_"bytes"
	"strconv"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"google.golang.org/grpc"
)

type IncomingBuyer struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}

type DgBuyer struct {
	Uid string `json:"uid,omitempty"`
	Id string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Age string `json:"age,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type Product struct {
	Uid string `json:"uid,omitempty"`
	Id string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Price string `json:"price,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type Transaction struct {
	Uid string `json:"uid,omitempty"`
	Id string `json:"id,omitempty"`
	Buyer string `json:"buyer,omitempty"`
	Ip string `json:"ip,omitempty"`
	Device string `json:"device,omitempty"`
	Products []string `json:"products,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

func main() {
	/*
	* chi router
	*/
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//loadSchema()

	// endpoint: load data to dgraph
	r.Get("/sync",func(w http.ResponseWriter, r *http.Request) {
		
		currentTime := time.Now().Format("2006-01-02")

		buyers := getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date", currentTime)
		prodData := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/products?date", currentTime))
		transData := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions?date", currentTime))

		/*
		* create connection to dgraph
		*/
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		/*
		* encode buyers
		*/
		var inBuyers []IncomingBuyer
		var outBuyers []DgBuyer
		if err := json.Unmarshal(buyers, &inBuyers); err != nil { log.Fatal(err) }
		
		for _, v := range inBuyers {
			outBuyers = append(outBuyers, DgBuyer{Uid: `_:`+v.Id, Id: v.Id, Name: v.Name, Age: strconv.Itoa(v.Age), DType: []string{"Buyer"} })
		}
		/*
		* process products
		*/
		var prodList []Product
		pLine := strings.Split(prodData,"\n")

		for _, inl := range pLine {
			inl2 := strings.Split(inl,`'`)

			if len(inl2) == 3 {
				prodList = append(prodList, Product{Uid: `_:`+inl2[0], Id: inl2[0], Name: inl2[1], Price: inl2[2], DType: []string{"Product"} })
			}else if len(inl2) == 4 {
				prodList = append(prodList, Product{Uid: `_:`+inl2[0], Id: inl2[0], Name: inl2[1]+inl2[2], Price: inl2[3], DType: []string{"Product"} })
			}
		}
		/*
		* process transactions
		*/
		var pTrans []Transaction
		tLine := strings.Split(transData,"#")

		for i, inl := range tLine {
			if i == 0 {continue}

			newL := strings.Split(inl,"\x00")

			// get buyer


			pTrans = append(pTrans, Transaction{
				Uid: `_:`+newL[0],
				Id: newL[0], 
				Buyer: newL[1], 
				Ip: newL[2], 
				Device: newL[3],
				Products: strings.Split(strings.Replace(strings.Replace(newL[4],"(","",1),")","",1), ","),
				DType: []string{"Transaction"} })
		}

		/*dgBuyers,err := json.Marshal(outBuyers)
		if err != nil { log.Fatal(err) }

		mu := &api.Mutation{ CommitNow: true, SetJson: dgBuyers}

		_, err = dgraphClient.NewTxn().Mutate(context.Background(), mu)
		if err != nil { log.Fatal(err) }*/

		w.Write([]byte("done"))
	})

	http.ListenAndServe(":5000", r)
}

func getData(url string, currentTime string) []byte { 
	querystr := fmt.Sprintf("%s=%s", url, currentTime)
	resp, err := http.Get(querystr)		
	if err != nil { log.Fatal(err) }		
	defer resp.Body.Close()		
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { log.Fatal(err) }
	return data
}

func loadSchema(){
	
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
	if err != nil { log.Fatal(err) }		
	defer conn.Close()		
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	op := &api.Operation{}
	op.Schema = `
	buyerId: string @index(exact) .
	name: string @index(exact) .
	age: string .
	type Buyer {
		buyerId: string
		name: string
		age: string
	}`
	
	if err := dgraphClient.Alter(context.Background(), op); err != nil {
		log.Fatal(err)
	}
}