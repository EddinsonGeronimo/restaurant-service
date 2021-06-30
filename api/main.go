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
	"os"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"github.com/go-chi/cors"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Control-Allow-Origin"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  }))

	if len(os.Args) > 1 && os.Args[1] == "--setup-schema" {
		dgraphClient := newClient()
		
		if err := dgraphClient.Alter(context.Background(), &api.Operation{Schema: SCHEMA}); err != nil {
			log.Fatal(err)
		}	
	}

	r.Route("/sync", func(r chi.Router) {
		r.Get("/",sync)
	})

	r.Route("/buyers", func(r chi.Router) {
		r.Get("/search", searchBuyers)

		r.Route("/{buyerId}", func(r chi.Router) {
			r.Get("/", getBuyer)
		})
	})

	http.ListenAndServe(":4000", r)
}

func sync(w http.ResponseWriter, r *http.Request){
	date := r.URL.Query().Get("date")

	dgraphClient := newClient()

	// drop all data before loading new data to dgraph
	if err := dgraphClient.Alter(context.Background(), &api.Operation{DropOp: api.Operation_DATA}); err != nil {
		log.Fatal(err)
		http.Error(w,http.StatusText(http.StatusInternalServerError),500)
	}

	if len(date) == 0 {
		date = time.Now().Format("2006-01-02")
	}

	c := make(chan itemData)
	defer close(c)
	go getData(AWS_ENDPOINT + "products?date", date, "products", c)
	go getData(AWS_ENDPOINT + "transactions?date", date,"transactions", c)
	go getData(AWS_ENDPOINT + "buyers?date", date,"buyers", c)
	
	var prodData, buyers, transData string

	result := make([]itemData, 3)

	for i,_ := range result{
		result[i] = <-c
		if result[i].item == "products"{
			prodData = result[i].data
		} else if result[i].item == "transactions"{
			transData = result[i].data
		} else if result[i].item == "buyers"{
			buyers = result[i].data
		}
	}

	/*
	* process buyers
	*/
	var awsBuyers []struct{
		Id string `json:"id"`
		Name string `json:"name"`
		Age int `json:"age"`
	}

	var dgBuyers []Buyer

	if err := json.Unmarshal([]byte(buyers), &awsBuyers); err != nil { log.Fatal(err) }
	
	processBuyers := func() {
		for _, v := range awsBuyers {
			dgBuyers = append(
				dgBuyers, 
				Buyer{
					Uid: `_:`+ v.Id, 
					Id: v.Id, 
					Name: v.Name, 
					Age: v.Age, 
					DType: []string{"Buyer"}, 
				},
			)
		}
	}

	/*
	* process products
	*/
	var prodList []Product
	pLine := strings.Split(string(prodData),"\n")

	processProducts := func(){
		for _, inl := range pLine {
			inl2 := strings.Split(inl,`'`)

			if len(inl2) == 3 {
				prodList = append(
					prodList, 
					Product{
						Uid: `_:`+inl2[0], 
						Id: inl2[0], 
						Name: inl2[1], 
						Price: inl2[2], 
						DType: []string{"Product"}, 
					},
				)
			} else if len(inl2) == 4 {
				prodList = append(
					prodList, 
					Product{
						Uid: `_:`+inl2[0], 
						Id: inl2[0], 
						Name: inl2[1]+inl2[2], 
						Price: inl2[3], 
						DType: []string{"Product"}, 
					},
				)
			}
		}
	}

	go processBuyers()
	go processProducts()

	/*
	* process transactions
	*/
	var pTrans []Transaction
	tLine := strings.Split(string(transData),"#")

	for i, inl := range tLine {
		if i == 0 {continue}

		newL := strings.Split(inl,"\x00")
		var dgbuyer Buyer
		var dgprodList []Product
		transProdIds := strings.Split(strings.Replace(strings.Replace(newL[4],"(","",1),")","",1), ",")

		// get buyer
		for _,v := range dgBuyers{
			if v.Id == newL[1] {dgbuyer = v}
		}

		// get products
		for _,v := range transProdIds {
			for _,v1 := range prodList {
				if v1.Id == v {
					dgprodList = append(dgprodList, v1)
				}
			}
		}

		pTrans = append(pTrans, Transaction{
			Uid: `_:`+newL[0],
			Id: newL[0], 
			Buyer: dgbuyer, 
			Ip: newL[2], 
			Device: newL[3],
			Products: dgprodList,
			DType: []string{"Transaction"} })
	}

	jsonData,err := json.Marshal(pTrans)
	if err != nil { log.Fatal(err) }

	mu := &api.Mutation{ CommitNow: true, SetJson: jsonData}

	_, err = dgraphClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil { 
		log.Fatal(err)
		http.Error(w,http.StatusText(http.StatusInternalServerError),500) 
	}
}

func searchBuyers(w http.ResponseWriter, r *http.Request){
	dgraphClient := newClient()

	resp, err := dgraphClient.NewReadOnlyTxn().Query(context.Background(), QUERY_BUYERS)
	if err != nil { log.Fatal(err) }

	w.Write(resp.GetJson())
}

func getBuyer(w http.ResponseWriter, r *http.Request){
	buyerId := chi.URLParam(r, "buyerId")

	dgraphClient := newClient()

	resp, err := dgraphClient.NewReadOnlyTxn().Query(context.Background(), QUERY_BUYER_INFO)
	if err != nil { log.Fatal(err) }

	// store all buyers and their transactions
	var decode struct { 
		Q []struct { 
			Id string `json:"id"`
			Name string `json:"name"`
			Age string `json:"age"`
			Transactions []struct{ 
				Id string `json:"id"`
				Ip string `json:"ip"`
				Device string `json:"device"`
				Products []struct{
					Name string `json:"name"`
					Price float64 `json:"price"`
				}
			}
		}
		Qproducts []struct {
			Id string `json:"id"`
			Name string `json:"name"`
			Ntrans int `json:"ntrans"`
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		log.Fatal(err)
	}

	// store all transactions of buyerId
	type BuyerTrans []struct{
		Id string `json:"id"`
		Ip string `json:"ip"`
		Device string `json:"device"`
		Products []struct{
			Name string `json:"name"`
			Price float64 `json:"price"`
		}
	}

	var buyerTrans BuyerTrans

	for _,v := range decode.Q {
		if buyerId == v.Id {
			buyerTrans = append(buyerTrans, v.Transactions...)
		}
	}

	type Buyer struct {
		Id string `json:"id"`
		Name string `json:"name"`
		Age string `json:"age"`
	}

	// store all buyers with same ip as buyerId
	var hasSameIp []Buyer

	for _,v := range buyerTrans {
		for _,v1 := range decode.Q {
			for _,v3 := range v1.Transactions{
				if v.Ip == v3.Ip {
					hasSameIp = append(hasSameIp, Buyer{Id: v1.Id, Name: v1.Name, Age: v1.Age})
				}
			}
		}
	}

	/*
	* remove duplicate from 'hasSameIp'
	*/
	var hasSameIpWithNoRep []Buyer
	list := []string{}
	var boolVar bool

	for _,v := range hasSameIp{
		boolVar = false
		for _, v1 := range list {
			if v1 == v.Id {
				boolVar = true
			}
		}
		if boolVar { continue }
		list = append(list, v.Id)
		hasSameIpWithNoRep = append(hasSameIpWithNoRep, v)
	}

	// store transactions and buyers with same ip as buyerId
	type AllData struct {
		BuyerTransactions []BuyerTrans `json:"buyertransactions"`
		HasSameIp []Buyer `json:"hassameip"`
		Rproducts []struct { 
			Id string `json:"id"` 
			Name string `json:"name"`
		}
	}

	var allData AllData

	allData.BuyerTransactions = append(allData.BuyerTransactions, buyerTrans)
	allData.HasSameIp = append(allData.HasSameIp, hasSameIpWithNoRep...)

	// filter products linked to more than 400 transactions  
	for _,v := range decode.Qproducts {
		if v.Ntrans > 400 {
			allData.Rproducts = append(allData.Rproducts, 
				struct{ 
					Id string `json:"id"` 
					Name string `json:"name"`}{
						Id: v.Id, 
						Name: v.Name,
					})
		}
	}

	data, err := json.Marshal(&allData)

	if err != nil { log.Fatal(err) }

	w.Write(data)
}

func getData(url string, date string, item string, c chan itemData) { 
	querystr := fmt.Sprintf("%s=%s", url, date)
	resp, err := http.Get(querystr)		
	if err != nil { log.Fatal(err) }		
	defer resp.Body.Close()		
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { log.Fatal(err) }
	c <- itemData{string(data), item}
}

func newClient() *dgo.Dgraph {

	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil { 
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)
}

const SCHEMA = `
	<Id>: string @index(exact) .
	<age>: string .
	<buyer>: uid @reverse .
	<buyers>: [uid] .
	<device>: string .
	<id>: string .
	<ip>: string .
	<name>: string .
	<price>: float .
	<products>: [uid] @count @reverse .
	type <Buyer> {
		Id
		name
		age
	}
	type <Product> {
		Id
		name
		price
	}
	type <Transaction> {
		Id
		buyer
		ip
		device
		products
	}`

const QUERY_BUYERS = `{ q (func: type(Buyer)) @filter(gt(count(~buyer),0)) { id name age } }`

const AWS_ENDPOINT = `https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/`

const QUERY_BUYER_INFO = `{ 
	q (func: type(Buyer)) {
		id 
		name 
		age 
		transactions: ~buyer { 
			id 
			ip
			device 
			products { name price } 
		  } 
	  }
	qProducts (func: type(Product)){
	  id 
	  name
	  ntrans: count(~products)
	}
  }`

type Buyer struct {
	Uid string `json:"uid,omitempty"`
	Id string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Age int `json:"age,omitempty"`
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
	Buyer Buyer `json:"buyer,omitempty"`
	Ip string `json:"ip,omitempty"`
	Device string `json:"device,omitempty"`
	Products []Product `json:"products,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type itemData struct {
	data string
	item string
}