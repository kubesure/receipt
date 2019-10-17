# receipt

#### biz design

Service creates a receipt for an payment. Binds the receipt to qoute and responds with a receipt number.

#### components

Mongodb v4

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

   k exec mongo-request-0 -it mongo

    rs.initiate({ _id: "rs0", members:[ 
        { _id: 0, host: "mongo-receipt-0.mongoreceiptsvc:27017" },
        { _id: 1, host: "mongo-receipt-1.mongoreceiptsvc:27017" },
        { _id: 2, host: "mongo-receipt-2.mongoreceiptsvc:27017" },
    ] });

    rs.conf()
    ```
