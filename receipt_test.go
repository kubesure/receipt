package main

import (
	"log"
	"testing"
)

func TestSaveReceipt(t *testing.T) {
	r, err := save(&Payment{Amount: 1234, PaymentMode: "internet"})
	log.Println(r.ReceiptNumber)
	log.Println(err)
}
