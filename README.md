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
  http://172.17.0.11:8000/api/v1/receipts \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: e7616b74-f54e-48e8-80e8-8f8851fd0fc5' \
  -H 'cache-control: no-cache' \
  -d '{ \             
    "Amount": 12345, \ 
    "PaymentMode": "1212121", \
    "quoteNumber": 11, \
    "paymentRefrence": "1212121" \
} \'
```
