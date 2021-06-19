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
	Buyer DgBuyer `json:"buyer,omitempty"`
	Ip string `json:"ip,omitempty"`
	Device string `json:"device,omitempty"`
	Products []Product `json:"products,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//loadSchema()

	// endpoint: load data to dgraph
	r.Get("/sync",func(w http.ResponseWriter, r *http.Request) {
		
		currentTime := chi.URLParam(r, "date")

		if len(currentTime) == 0 {
			currentTime = time.Now().Format("2006-01-02")
		}

		buyers := getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/buyers?date", currentTime)
		prodData := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/products?date", currentTime))
		transData  := string(getData("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions?date", currentTime))

		/*
		* connection to dgraph
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

		w.Write([]byte("done"))
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
	
	http.ListenAndServe(":3000", r)
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