package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoreceiptsvc = os.Getenv("mongoreceiptsvc")

//Error Code Enum
const (
	SystemErr = iota
	InputJSONInvalid
	AgeRangeInvalid
	RiskDetailsInvalid
	InvalidRestMethod
	InvalidContentType
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
}

//Payment as a API input
type Payment struct {
	Amount           int    `json:"amount"`
	PaymentMode      string `json:"paymentMode"`
	QuoteNumber      int64  `json:"quoteNumber"`
	PaymentReference string `json:"paymentRefrence"`
}

//Receipt is response of API
type Receipt struct {
	ReceiptNumber int `json:"receiptNumber"`
}

type errorresponse struct {
	Code    int    `json:"errorCode"`
	Message string `json:"errorMessage"`
}

func main() {
	log.Debug("server receipt starting...")
	mux := http.NewServeMux()
	//mux.HandleFunc("/", healthz)
	mux.HandleFunc("/isready", isReady)
	mux.HandleFunc("/api/v1/receipts", receipt)
	srv := http.Server{Addr: ":8000", Handler: mux}
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Debug("shutting down receipt server...")
			srv.Shutdown(ctx)
			<-ctx.Done()
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %s", err)
	}
}

//HTTP API called by k8s readiness probe.
func isReady(w http.ResponseWriter, req *http.Request) {
	client, errping := conn()
	if errping != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	coll := client.Database("receipts").Collection("receipt")
	if coll == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//not utilized
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	data := (time.Now()).String()
	log.Debug("health ok")
	w.Write([]byte(data))
}

//Create a receipt
func receipt(w http.ResponseWriter, req *http.Request) {

	p, err := validateReq(w, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(err)
		fmt.Fprintf(w, "%s", data)
	} else {
		r, serr := save(p)
		if serr != nil {
			log.Error(serr)
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			data, _ := json.Marshal(r)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func validateReq(w http.ResponseWriter, req *http.Request) (*Payment, *errorresponse) {
	if req.Method != http.MethodPost {
		log.Error("invalid method ", req.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, &errorresponse{Code: InvalidRestMethod, Message: fmt.Sprintf("Invalid method %s", req.Method)}
	}

	if req.Header.Get("Content-Type") != "application/json" {
		log.Error("invalid content type ", req.Header.Get("Content-Type"))
		msg := fmt.Sprintf("Invalid content-type %s require %s", req.Header.Get("Content-Type"), "application/json")
		return nil, &errorresponse{Code: InvalidContentType, Message: msg}
	}

	body, _ := ioutil.ReadAll(req.Body)
	p, merr := marshallProposal(string(body))

	if merr != nil {
		return nil, merr
	}
	return p, nil
}

func marshallProposal(data string) (*Payment, *errorresponse) {
	var p Payment
	err := json.Unmarshal([]byte(data), &p)
	if err != nil {
		return nil, &errorresponse{Code: InputJSONInvalid, Message: "Invalid Input"}
	}

	if p.Amount == 0 || len(p.PaymentMode) == 0 || len(p.PaymentReference) == 0 || p.QuoteNumber == 0 {
		return nil, &errorresponse{Code: InputJSONInvalid, Message: "Invalid Input"}
	}

	return &p, nil
}

//Save the payment for the policy and return a receipt.
func save(p *Payment) (*Receipt, error) {

	client, errping := conn()

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
		{"paymentMode", p.PaymentMode}, {"paymentReference", p.PaymentReference},
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

//Generate new receipt number.
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

func conn() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	uri := "mongodb://" + mongoreceiptsvc + "/?replicaSet=rs0"
	log.Debug(uri)
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	//client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongoreceiptsvc+":27017"))
	errping := client.Ping(ctx, nil)
	if errping != nil {
		return nil, errping
	}
	return client, nil
}
