package main

import (
	"encoding/base32"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func getDB(filename string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		log.Printf("Error in db: %v", err)
		return nil, err
	}
	return db, nil
}

func dbFields(values interface{}) ([]string, error) {

	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fields := []string{}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i).Tag.Get("db")
			if field != "" {

				fields = append(fields, field)
			}
		}
		return fields, nil
	}
	return nil, errors.New("no data")
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})

}

func toInt(s string) int {
	d, err := strconv.Atoi(s)
	if err != nil {
		return -9999
	}
	return d
}

// var SMSKey = make([]byte, 10)

// var SMSKey string
// var smsErr error

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func secret(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return b
}

func generateOTP() (string, error) {

	secret := base32.StdEncoding.EncodeToString([]byte("12345678"))
	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    4,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		panic(err)
		log.Printf("error in totp: %v", err)
		return "", err
	}
	log.Printf("passcode: %s", passcode)
	return passcode, nil
}

func validateOTP(key string) bool {

	secret := base32.StdEncoding.EncodeToString([]byte("12345678"))
	if ok, err := totp.ValidateCustom(key, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    4,
		Algorithm: otp.AlgorithmSHA1,
	}); err != nil {
		log.Printf("err is: %v", err)
		return ok
	} else {
		return ok
	}

}
