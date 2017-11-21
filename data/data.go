package data

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var Db *gorm.DB

func init() {
	databaseDSN := "mybookzy:ZasSaMi&ZasSaMi@tcp(52.187.124.136:3306)/bookzywpdb?charset=utf8&parseTime=true"
	Db, err := gorm.Open("mysql", databaseDSN)
	if err != nil {
		log.Panic(err)
	}
	Db.LogMode(gin.Mode() == "debug")
}

func createUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		log.Fatalln("Cannot generate UUID", err)
	}

	// 0x40 is reserved variant from RFC 4122
	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}
