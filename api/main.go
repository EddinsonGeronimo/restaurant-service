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
	"strconv"
	"os"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"github.com/go-chi/cors"
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
	Buyer DgBuyer `json:"buyer,omitempty"`
	Ip string `json:"ip,omitempty"`
	Device string `json:"device,omitempty"`
	Products []Product `json:"products,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Control-Allow-Origin"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  }))

	if len(os.Args) > 1 && os.Args[1] == "--load-schema" {
		loadSchema()	
	}

	// endpoint: load data to dgraph
	r.Get("/sync",func(w http.ResponseWriter, r *http.Request) {

		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		// drop all data before loading new data to dgraph
		if err := dgraphClient.Alter(context.Background(), &api.Operation{DropOp: api.Operation_DATA}); err != nil {
			log.Fatal(err)
		}
		
		currentTime := r.URL.Query().Get("date")

		if len(currentTime) == 0 {
			currentTime = time.Now().Format("2006-01-02")
		}

		buyers := getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date", currentTime)
		prodData := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/products?date", currentTime))
		transData  := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions?date", currentTime))

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
			var dgbuyer DgBuyer
			var dgprodList []Product
			transProdIds := strings.Split(strings.Replace(strings.Replace(newL[4],"(","",1),")","",1), ",")

			// get buyer
			for _,v := range outBuyers{
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
		if err != nil { log.Fatal(err) }

		w.Write([]byte(fmt.Sprintf(`{"date": "%s"}`, currentTime)))
	})

	// endpoint: return buyers who have transactions
	r.Get("/buyers",func(w http.ResponseWriter, r *http.Request){
		/*
		* connection to dgraph
		*/
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		const query = `{ q (func: type(Buyer)) { id name age transactions: ~buyer { id } } }`

		resp, err := dgraphClient.NewReadOnlyTxn().Query(context.Background(), query)
		if err != nil { log.Fatal(err) }

		var decode struct { 
			Q []struct { 
				Id string `json:"id"`
				Name string `json:"name"`
				Age string `json:"age"`
				Transactions []struct{ Id string `json:"id"`}
			} 
		}
		
		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			log.Fatal(err)
		}

		type Buyer struct{
			Id string `json:"id"`
			Name string `json:"name"`
			Age string `json:"age"`
		}

		var buyersWithTrans []Buyer

		for _,v := range decode.Q{
			if len(v.Transactions) > 0 {
				buyersWithTrans = append(buyersWithTrans, Buyer{Id: v.Id, Name: v.Name, Age: v.Age})
			}
		}
		
		data, err := json.Marshal(&buyersWithTrans)

		if err != nil { log.Fatal(err) }

		// if 'data' is empty than 'data' is null
		w.Write(data)
	})
	
	// endpoint: return transactions of buyerId and buyers with same ip , also recommended products 
	r.Get("/buyer",func(w http.ResponseWriter, r *http.Request){
		
		buyerId := r.URL.Query().Get("id")

		/*
		* connection to dgraph
		*/
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

		const query = `{ 
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

		resp, err := dgraphClient.NewReadOnlyTxn().Query(context.Background(), query)
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

		// store all buyers with same ips as buyerId
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

		for _,v := range hasSameIp{
			if contains(list, v.Id) {
				continue
			}
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

		//if 'data' is empty than 'data' is null
		w.Write(data)
	})

	http.ListenAndServe(":4000", r)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
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
	
	if err := dgraphClient.Alter(context.Background(), op); err != nil {
		log.Fatal(err)
	}
}