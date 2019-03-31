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
	"time"
)

var mongoreceiptsvc = os.Getenv("mongoreceiptsvc")

//Payment as a API input
type Payment struct {
	Amount      int    `json:"amount"`
	PaymentMode string `json:"paymentmode"`
}

//Receipt is response of API
type Receipt struct {
	ReceiptNumber int `json:"receiptnumber"`
}

func main() {
	log.Println("server receipt starting...")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/receipts", receipt)
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func receipt(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	/*TODO error handling best prac to be implemted*/
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
		log.Println("error in unmarshalling payment", err)
		return nil, err
	}
	log.Println("return p", p)
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
	_, errcol := collection.InsertOne(context.Background(), bson.M{
		"receiptNumber": id, "amount": p.Amount, "paymentmode": p.PaymentMode})

	if errcol != nil {
		log.Println("errcol")
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
