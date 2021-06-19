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
	ID string `json:"Id,omitempty"`
	Name string `json:"name,omitempty"`
	Price string `json:"price,omitempty"`
}

type Transaction struct {
	ID string `json:"Id,omitempty"`
	BuyerID string `json:"buyer,omitempty"`
	IP string `json:"ip,omitempty"`
	Device string `json:"device,omitempty"`
	ProductIDs []string `json:"products,omitempty"`
}

func main() {
	/*
	* chi router
	*/
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//loadSchema()

	// load data to local database
	r.Get("/",func(w http.ResponseWriter, r *http.Request) {
		
		currentTime := time.Now().Format("2006-01-02")
		/* 
		* get buyers from external endpoint
		*/
		urlBuyers := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date=%s", currentTime)
		respBuyers, err := http.Get(urlBuyers)		
		if err != nil { log.Fatal(err) }		
		defer respBuyers.Body.Close()		
		buyers, err := ioutil.ReadAll(respBuyers.Body)
		if err != nil { log.Fatal(err) }
		/* 
		* get products from external endpoint
		*/
		urlProducts := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/products?date=%s", currentTime)
		respProducts, err := http.Get(urlProducts)		
		if err != nil { log.Fatal(err) }		
		defer respProducts.Body.Close()		
		_, err = ioutil.ReadAll(respProducts.Body)
		if err != nil { log.Fatal(err) }		
		/* 
		* get buyers from external endpoint
		*/
		urlTrans := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions?date=%s", currentTime)
		respTrans, err := http.Get(urlTrans)		
		if err != nil { log.Fatal(err) }		
		defer respTrans.Body.Close()		
		_, err = ioutil.ReadAll(respTrans.Body)
		if err != nil { log.Fatal(err) }

		/*
		* create connection to dgraph
		*/
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		/*
		* mutation
		*/
		var inBuyers []IncomingBuyer
		var outBuyers []DgBuyer
		if err := json.Unmarshal(buyers, &inBuyers); err != nil { log.Fatal(err) }
		
		for _, v := range inBuyers {
			outBuyers = append(outBuyers, DgBuyer{Uid: `_:`+v.Id, Id: v.Id, Name: v.Name, Age: strconv.Itoa(v.Age), DType: []string{"Buyer"} })
		}
		
		dgBuyers,err := json.Marshal(outBuyers)
		if err != nil { log.Fatal(err) }

		mu := &api.Mutation{ CommitNow: true, SetJson: dgBuyers}

		_, err = dgraphClient.NewTxn().Mutate(context.Background(), mu)
		if err != nil { log.Fatal(err) }

		w.Write([]byte("done"))
	})

	http.ListenAndServe(":5000", r)
}

func getData(){
	
}

func productsData(data string) []Product{
	var pList []Product
	line := strings.Split(data,"\n")

	for _, inl := range line {
		inl2 := strings.Split(inl,`'`)

		if len(inl2) == 3 {
			pList = append(pList, Product{ID: inl2[0], Name: inl2[1], Price: inl2[2]})
		}else if len(inl2) == 4 {
			pList = append(pList, Product{ID: inl2[0], Name: inl2[1]+inl2[2], Price: inl2[3]})
		}
	}
	return pList
}

func transData(data string) []Transaction{
	var pTrans []Transaction

	line := strings.Split(data,"#")

	for i, inl := range line {
		if i == 0 {continue}

		newL := strings.Split(inl,"\x00")
		pTrans = append(pTrans, 
			Transaction{
				ID: newL[0], 
				BuyerID: newL[1], 
				IP: newL[2], 
				Device: newL[3],
				ProductIDs: strings.Split(strings.Replace(strings.Replace(newL[4],"(","",1),")","",1), ",")})
	}

	return pTrans
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