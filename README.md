# receipt

db.counter.insert({"_id" : "receiptid" , "value": 0 })

curl -i -X POST \
  http://172.17.0.11:8000/api/v1/receipts \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: e7616b74-f54e-48e8-80e8-8f8851fd0fc5' \
  -H 'cache-control: no-cache' \
  -d '{
    "Amount": 12345,
    "PaymentMode": "1212121",
    "quoteNumber": 11,
    "paymentRefrence": "1212121"
}'