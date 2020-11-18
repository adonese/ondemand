package main

import (
	"encoding/base32"
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func getDB(filename string) (*sqlx.DB, error) {
	if path := os.Getenv("DB_PATH"); path != "" {
		filename = path
	}

	db, err := sqlx.Connect("sqlite3", "test.db")
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

func toString(i int) string {
	return strconv.Itoa(i)
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

func generateOTP(hash string) (string, error) {

	secret := base32.StdEncoding.EncodeToString([]byte(hash + "12345678"))
	passcode, err := totp.GenerateCodeCustom(secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    360,
		Skew:      1,
		Digits:    4,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		// panic(err)
		log.Printf("error in totp: %v", err)
		return "", err
	}
	log.Printf("passcode: %s", passcode)
	return passcode, nil
}

func validateOTP(key string, hash string) bool {

	secret := base32.StdEncoding.EncodeToString([]byte(hash + "12345678"))
	if ok, err := totp.ValidateCustom(key, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    360,
		Skew:      1,
		Digits:    4,
		Algorithm: otp.AlgorithmSHA1,
	}); !ok {
		log.Printf("err is: %v", err)
		return false
	} else {
		return true
	}

}

//haverSine returns approximate distance between a pair of (lat1, lon1), (lat2, lon2)
func haverSine(lat1, lat2, lon1, lon2 float64) float64 {

	var R = 6371.0                            // Radius of the earth in km
	var dLat = deg2rad(math.Abs(lat2 - lat1)) // deg2rad below
	var dLon = deg2rad(math.Abs(lon2 - lon1))
	var a = math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	var d = R * c // Distance in km
	return d
}

func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}
