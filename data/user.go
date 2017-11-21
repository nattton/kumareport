package data

import (
	"crypto/md5"
	"log"
	"strings"
	"time"
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
func (user *WpUser) CreateSession() (session Session, err error) {
	session = Session{
		Uuid:   createUUID(),
		Email:  user.UserEmail,
		UserId: user.ID,
	}
	err = Db.Create(&session).Error
	return
}

// Get the session for an existing user
func (user *WpUser) Session() (session Session, err error) {
	err = Db.Where("user_id = ?", user.ID).First(&session).Error
	return
}

// Check if session is valid in the database
func (session *Session) Check() (valid bool, err error) {
	err = Db.Where("uuid = ?", session.Uuid).First(&session).Error
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
func (session *Session) DeleteByUUID() (err error) {
	err = Db.Where("uuid = ?", session.Uuid).Delete(Session{}).Error
	return
}

// Get the user from the session
func (session *Session) User() (user WpUser, err error) {
	err = Db.First(&user, session.UserId).Error
	return
}

// Delete all sessions from database
func SessionDeleteAll() (err error) {
	return Db.Delete(Session{}).Error
}

// Get a single user given the email
func UserByLogin(email string) (user WpUser, err error) {
	err = Db.Where("user_login = ? OR user_email = ?", email, email).First(&user).Error
	return
}

// Get a single user given the UUID
func UserByUUID(uuid string) (user WpUser, err error) {
	err = Db.Where("uuid = ?", uuid).First(&user).Error
	return
}

func PasswordHashCheck(pw, storedHash string) bool {
	hx := cryptPrivate(pw, storedHash)
	log.Println(hx)
	return hx == storedHash
}

func encode64(inp []byte, count int) string {
	const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var outp string
	cur := 0
	for cur < count {
		value := uint(inp[cur])
		cur += 1
		outp += string(itoa64[value&0x3f])
		if cur < count {
			value |= (uint(inp[cur]) << 8)
		}
		outp += string(itoa64[(value>>6)&0x3f])

		if cur >= count {
			break
		}
		cur += 1
		if cur < count {
			value |= (uint(inp[cur]) << 16)
		}
		outp += string(itoa64[(value>>12)&0x3f])
		if cur >= count {
			break
		}
		cur += 1
		outp += string(itoa64[(value>>18)&0x3f])
	}
	return outp
}

func cryptPrivate(pw, setting string) string {
	const itoa64 = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var outp = "*0"
	var count_log2 uint
	count_log2 = uint(strings.Index(itoa64, string(setting[3])))
	if count_log2 < 7 || count_log2 > 30 {
		return outp
	}
	count := 1 << count_log2
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
		count -= 1
	}
	return setting[:12] + encode64(hx, 16)
}
