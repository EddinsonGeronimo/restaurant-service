package main

import (
	_ "context"
	_ "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/dgraph-io/dgo"
	_ "github.com/dgraph-io/dgo/protos/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "google.golang.org/grpc"
)

type Buyer struct {
	ID string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Age string `json:"age,omitempty"`
}

type Product struct {
	ID string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Price string `json:"price,omitempty"`
}

type Transaction struct {
	ID string `json:"id,omitempty"`
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
		_, err = ioutil.ReadAll(respBuyers.Body)
		if err != nil { log.Fatal(err) }
		/* 
		* get products from external endpoint
		*/
		urlProducts := fmt.Sprintf("https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/products?date=%s", currentTime)
		respProducts, err := http.Get(urlProducts)		
		if err != nil { log.Fatal(err) }		
		defer respProducts.Body.Close()		
		bodyBuyers, err := ioutil.ReadAll(respProducts.Body)
		if err != nil { log.Fatal(err) }
		fmt.Println(productsData(string(bodyBuyers)))
		
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
		* open connection to local database
		*
		conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())		
		if err != nil { log.Fatal(err) }		
		defer conn.Close()		
		_ = dgo.NewDgraphClient(api.NewDgraphClient(conn))
		*
		* 
		*/
	})

	http.ListenAndServe(":3000", r)
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