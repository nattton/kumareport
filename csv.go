package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jinzhu/gorm"
)

func WriteCSVFile(csvRow [][]string, fileName string) {
	csvfile, err := os.Create(GetFileNameNow(fileName) + ".csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.WriteAll(csvRow) // calls Flush internally
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}

func ReadFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return b
}

func ReadCSV(file string) [][]string {
	b := ReadFile(file)
	r := csv.NewReader(bytes.NewReader(b))
	r.Comma = ','
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	return records
}

func ImportStock(db *gorm.DB) {
	lines := ReadCSV("KumamonInventory.csv")
	for _, line := range lines[1:] {
		db.Exec("UPDATE products SET stock= ? WHERE sku=?", line[1], line[0])
	}
}

func ImportEMS(db *gorm.DB) {
	lines := ReadCSV("ems.csv")
	fmt.Print(lines)
	for _, line := range lines[1:] {
		log.Println(line[1], line[0])
		db.Exec("UPDATE attendees SET ems = ? WHERE order_id=?", line[1], line[0])
	}
}
