package anypay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const authKey = "bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1bmlxdWVfbmFtZSI6ImY3MTgyM2UzLWJhOGMtNGMwNi05ODkzLTA0MDgzMzJhNDA1NSIsInN1YiI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJlbWFpbCI6ImJvb2t6eUBhbnlwYXkuY28udGgiLCJyb2xlIjoiQU5ZUEFfV0VCIiwiaXNzIjoiYXV0aC5hbnlwYXkuY28udGgiLCJhdWQiOiI2ODViZDg0MmY3ZDQ0NmU2OWY4Yjc4ZTY0MzJjNzY3OCIsImV4cCI6MTUxNTUzNDYxNCwibmJmIjoxNTEwMzUwNjE0fQ.M7CB3T2PmzT2L9laT2m-8WiJpHCD1RcqEvv0SoepT6Q"

type JSONPaymentStatus struct {
	TotalRow int
	DataRow  []PaymentStatus
}

type PaymentStatus struct {
	RefID           string
	TransDesc       string
	TransId         string
	PaymentCode     string
	PaymentMessage  string
	Amount          float64
	PaymentType     string
	PaymentAmount   string
	PaymentDateTime string
	OrderNo         string
}

func GetPaymentStatus(refID string) (status PaymentStatus, err error) {
	body := callAnypay(fmt.Sprintf("/payment/status2?RefId=%s", refID))
	var j JSONPaymentStatus
	err = json.Unmarshal(body, &j)
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(j.DataRow) < 1 {
		err = errors.New("No Data")
	}
	status = j.DataRow[0]
	return
}

func callAnypay(apiPath string) []byte {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.anypay.co.th%s", apiPath), nil)
	req.Header.Set("Authorization", authKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return body
}
