# receipt

#### biz design

Service creates a receipt for an payment. Binds the receipt to qoute and responds with a receipt number.

#### components

Mongodb v4, Golang 

#### Dev setup and test

1. create db receipt 
    ```
       use receipts
       db.counter.insert({"_id" : "receiptid" , "value": 0 })
       db.counter.find({}).pretty()
    ```
2. Run receipt
   ```
      go run receipt.go 
   ```

3. Run curl to create receipt

```
curl -i -X POST \
  http://localhost:8000/api/v1/receipts \
  -H 'Content-Type: application/json' \
  -d '{              
    "Amount": 12345,  
    "PaymentMode": "1212121", 
    "quoteNumber": 11, 
    "paymentRefrence": "1212121" 
}'
```

## Pod 

1. Create and configure mongo in k8s
 
    ```
    alias k=kubectl
    complete -F __start_kubectl k

   k exec mongo-quote-0 -it mongo

    rs.initiate({ _id: "rs0", members:[ 
        { _id: 0, host: "mongo-quote-0.mongoquotesvc:27017" },
        { _id: 1, host: "mongo-quote-1.mongoquotesvc:27017" },
        { _id: 2, host: "mongo-quote-2.mongoquotesvc:27017" },
    ] });

    rs.conf()
    ```

    create document store follow step 2 in Dev setup 

2. Create quote docker image
    
    ```
    docker build . -t quote:v1
    k apply config/quote.yaml
    k get po -o wide
    ```

3.  Apply Quote to k8s

    ```
        k apply -f config/quote.yaml
        k get po -o wide 
    ```    

    curl test. Follow step 4.
