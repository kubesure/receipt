package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var mongoreceiptsvc = os.Getenv("mongoreceiptsvc")

//Payment as a API input
type Payment struct {
	Amount           int    `json:"amount"`
	PaymentMode      string `json:"paymentMode"`
	QuoteNumber      int64  `json:"quoteNumber"`
	paymentReference string `json:"refrenceNumber"`
}

//Receipt is response of API
type Receipt struct {
	ReceiptNumber int `json:"receiptNumber"`
}

func main() {
	log.Println("server receipt starting...")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/receipts", receipt)
	srv := http.Server{Addr: ":8080", Handler: mux}
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Print("shutting down receipt server...")
			srv.Shutdown(ctx)
			<-ctx.Done()
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %s", err)
	}

	//log.Fatal(srv.ListenAndServe(":8000", mux))
}

func receipt(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	p, merr := marshallProposal(string(body))
	r, serr := save(p)
	if merr != nil {
		log.Println(merr)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if serr != nil {
		log.Println(serr)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		data, _ := json.Marshal(r)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s", data)
	}
}

func marshallProposal(data string) (*Payment, error) {
	var p Payment
	err := json.Unmarshal([]byte(data), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func save(p *Payment) (*Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongoreceiptsvc+":27017"))
	errping := client.Ping(ctx, nil)

	if errping != nil {
		return nil, errping
	}

	collection := client.Database("receipts").Collection("receipt")
	id, errSeq := nextcounter(client)
	if errSeq != nil {
		return nil, errSeq
	}

	_, errcol := collection.InsertOne(context.Background(), bson.D{
		{"receiptNumber", id}, {"quoteNumber", p.QuoteNumber}, {"amount", p.Amount},
		{"paymentMode", p.PaymentMode}, {"paymentReference", p.paymentReference},
		{"createdDate", time.Now().String()},
	})

	if errcol != nil {
		return nil, errcol
	}

	//oid, _ := res.InsertedID.(primitive.ObjectID)
	//r := Receipt{ReceiptNumber: oid.Hex()}
	r := Receipt{ReceiptNumber: id}
	return &r, nil
}

func nextcounter(c *mongo.Client) (int, error) {
	collection := c.Database("receipts").Collection("counter")
	filter := bson.M{"_id": "receiptid"}
	update := bson.M{"$inc": bson.M{"value": 1}}
	aft := options.After
	opt := options.FindOneAndUpdateOptions{Upsert: new(bool), ReturnDocument: &aft}
	result := collection.FindOneAndUpdate(context.Background(), filter, update, &opt)
	type record struct {
		Receiptid string `bson:"receiptid"`
		Value     int    `bson:"value"`
	}
	var data record
	err := result.Decode(&data)
	if err != nil {
		return 0, err
	}
	return data.Value, nil
}
