package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

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

func ImportShirt(db *gorm.DB) {
	m := NewModel(db)
	m.RetieveTickets()
	m.RetieveSkuList()

	lines := ReadCSV("attendee-2017-11-15_15-52.csv")
	for _, record := range lines[1:] {
		fmt.Printf("Record %s %s\n", record[0], record[8])
		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		product := m.GetSkuList(record[8])
		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key='ticket_type_id' AND post_id = ?", product.ID, record[0])
		attendee := GetAttendee(db, m, id)
		db.Save(&attendee)
	}
}

func ImportChangeShirt(db *gorm.DB) {
	m := NewModel(db)
	m.RetieveTickets()
	m.RetieveSkuList()
	lines := ReadCSV("ChangeSize2.csv")
	for _, record := range lines[1:] {
		fmt.Printf("Record %s %s %s\n", record[0], record[1], record[2])
		postMetas := []WpPostmeta{}
		db.Where("meta_value LIKE ? AND meta_key = ?", "%"+record[2]+"%", kPhone).Find(&postMetas)
		resultCount := len(postMetas)
		if resultCount != 1 {
			log.Panicf("Result =%d | MObile No. %s", resultCount, record[2])
		}
		product := m.GetSkuList(record[1])

		db.Exec("UPDATE wp_postmeta SET meta_value= ? WHERE meta_key='ticket_type_id' AND post_id = ?", product.ID, postMetas[0].PostID)
		attendee := GetAttendee(db, m, postMetas[0].PostID)
		db.Save(&attendee)
	}
}

func ImportStock(db *gorm.DB) {
	lines := ReadCSV("KumamonInventory.csv")
	for _, line := range lines[1:] {
		db.Exec("UPDATE products SET stock= ? WHERE sku=?", line[1], line[0])
	}
}

func ImportStockLeft(db *gorm.DB) {
	lines := ReadCSV("KumamonInventoryLeft.csv")
	stockLeft := make(map[string]int)
	for _, line := range lines {
		left, _ := strconv.Atoi(line[1])
		stockLeft[line[0]] = left
	}

	skus := GetSkuAmount(db)
	for _, sku := range skus {
		stock := stockLeft[sku.ShirtSize] + sku.Amount
		db.Exec("UPDATE products SET stock= ? WHERE sku=?", stock, sku.ShirtSize)
	}
}
