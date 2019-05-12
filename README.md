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
  172.17.0.2:8000/api/v1/receipts \
  -H 'Content-Type: application/json' \
  -d '{ \             
    "Amount": 12345, \ 
    "PaymentMode": "1212121", \
    "quoteNumber": 11, \
    "paymentRefrence": "1212121" \
} \'
```
