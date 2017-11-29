package data

import (
	"crypto/md5"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type WpUser struct {
	ID           int `gorm:"primary_key;column:ID"`
	UserLogin    string
	UserPass     string
	UserNicename string
	UserEmail    string
	DisplayName  string
}

type Session struct {
	ID        int `gorm:"primary_key"`
	Uuid      string
	Email     string
	UserId    int
	CreatedAt time.Time
}

// Create a new session for an existing user
func (user *WpUser) CreateSession() (db *gorm.DB, session Session, err error) {
	session = Session{
		Uuid:   createUUID(),
		Email:  user.UserEmail,
		UserId: user.ID,
	}
	err = db.Create(&session).Error
	return
}

// Get the session for an existing user
func (user *WpUser) Session() (db *gorm.DB, session Session, err error) {
	err = db.Where("user_id = ?", user.ID).First(&session).Error
	return
}

// Check if session is valid in the database
func (session *Session) Check(db *gorm.DB) (valid bool, err error) {
	err = db.Where("uuid = ?", session.Uuid).First(&session).Error
	if err != nil {
		valid = false
		return
	}
	if session.ID != 0 {
		valid = true
	}
	return
}

// Delete session from database
func (session *Session) DeleteByUUID(db *gorm.DB) (err error) {
	err = db.Where("uuid = ?", session.Uuid).Delete(Session{}).Error
	return
}

// Get the user from the session
func (session *Session) User(db *gorm.DB) (user WpUser, err error) {
	err = db.First(&user, session.UserId).Error
	return
}

// Delete all sessions from database
func SessionDeleteAll(db *gorm.DB) (err error) {
	return db.Delete(Session{}).Error
}

// Get a single user given the email
func UserByLogin(db *gorm.DB, email string) (user WpUser, err error) {
	err = db.Where("user_login = ? OR user_email = ?", email, email).First(&user).Error
	return
}

// Get a single user given the UUID
func UserByUUID(db *gorm.DB, uuid string) (user WpUser, err error) {
	err = db.Where("uuid = ?", uuid).First(&user).Error
	return
}

func PasswordHashCheck(pw, storedHash string) bool {
	hx := cryptPrivate(pw, storedHash)
	return hx == storedHash
}

func encode64(inp []byte, count int) string {
	const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var outp string
	cur := 0
	for cur < count {
		value := uint(inp[cur])
		cur++
		outp += string(itoa64[value&0x3f])
		if cur < count {
			value |= (uint(inp[cur]) << 8)
		}
		outp += string(itoa64[(value>>6)&0x3f])

		if cur >= count {
			break
		}
		cur++
		if cur < count {
			value |= (uint(inp[cur]) << 16)
		}
		outp += string(itoa64[(value>>12)&0x3f])
		if cur >= count {
			break
		}
		cur++
		outp += string(itoa64[(value>>18)&0x3f])
	}
	return outp
}

func cryptPrivate(pw, setting string) string {
	const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var outp = "*0"
	var countLog2 uint
	countLog2 = uint(strings.Index(itoa64, string(setting[3])))
	if countLog2 < 7 || countLog2 > 30 {
		return outp
	}
	count := 1 << countLog2
	salt := setting[4:12]
	if len(salt) != 8 {
		return outp
	}
	hasher := md5.New()
	hasher.Write([]byte(salt + pw))
	hx := hasher.Sum(nil)
	for count != 0 {
		hasher := md5.New()
		hasher.Write([]byte(string(hx) + pw))
		hx = hasher.Sum(nil)
		count--
	}
	return setting[:12] + encode64(hx, 16)
}
