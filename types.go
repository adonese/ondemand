package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const (
	NA = iota
	Pending
	Accepted
	Rejected
)

type errorHandler struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e errorHandler) toJson() []byte {
	d, _ := json.Marshal(&e)
	return d
}

func unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

func marshal(o interface{}) []byte {
	d, _ := json.Marshal(&o)
	return d
}

type Service struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	db   *sqlx.DB
}

func (c *Service) all() ([]Service, error) {
	var services []Service

	c.db.Exec(stmt)
	if err := c.db.Select(&services, "select * from services"); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *Service) getHandler(w http.ResponseWriter, r *http.Request) {
	/*
			{[
		        'تكييف',
		        'اعمال جبسية و اسقف',
		        'مكافحة الحشرات و القوارض',
		        'كهرباء',
		        'ارضيات وباركية',
		        'تنسيق الأشجار',
		        'تست',
		    ]}
	*/
	svcs := []string{"تكييف",
		"اعمال جبسية و اسقف",
		"مكافحة الحشرات و القوارض",
		"كهرباء",
		"ارضيات وباركية",
		"تنسيق الأشجار",
		"تست"}

	maps := make(map[string][]string)
	maps["result"] = svcs
	res, _ := json.Marshal(maps)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	w.Write(res)
	return
	service, err := c.all()
	if err != nil {
		vErr := errorHandler{Code: "user_not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	w.Write(marshal(service))
}

func (c *Service) serviceDetailsHandler(w http.ResponseWriter, r *http.Request) {
	/*
			{[
		   [
		        "عطل مكيف",
				"مشكلة في السباكة",
				"مشكلة في الكهرباء",
				"صيانة عامة",
				"آخرى"
		    ]
		    ]}
	*/
	svcs := []string{"عطل مكيف",
		"مشكلة في السباكة",
		"مشكلة في الكهرباء",
		"صيانة عامة",
		"آخرى"}

	maps := make(map[string][]string)
	maps["result"] = svcs
	res, _ := json.Marshal(maps)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	w.Write(res)
	return
	service, err := c.all()
	if err != nil {
		vErr := errorHandler{Code: "user_not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	w.Write(marshal(service))
}

func (c *Service) save() error {

	c.db.Exec(stmt)
	tx := c.db.MustBegin()
	tx.NamedExec("insert into services(name) values(:name)", c)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in cart.save: TX: %v", err)
		return err
	}
	return nil

}

func (c *Service) populateTest() error {

	c.db.Exec(stmt)
	tx := c.db.MustBegin()
	tx.Exec("insert into services(name) values(:name)", c)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in cart.save: TX: %v", err)
		return err
	}
	return nil

}

type Order struct {
	ID         int  `json:"id" db:"id"`
	UserID     int  `json:"user_id" db:"user_id"`
	ProviderID int  `json:"provider_id" db:"provider_id"`
	Status     bool `json:"status" db:"status"`
	CreatedAt  sql.NullTime
	db         *sqlx.DB
}

func (c *Order) get(id int) ([]Order, error) {
	var services []Order

	c.db.Exec(stmt)
	if err := c.db.Select(&services, "select * from orders"); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *Order) save() error {
	c.db.Exec(stmt)
	tx := c.db.MustBegin()
	tx.NamedExec("insert into orders(user_id, created_at, provider_id, status) values(:user_id, :created_at, :provider_id, :status)", c)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in cart.save: TX: %v", err)
		return err
	}
	return nil

}

func (c *Order) saveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	unmarshal(b, c)
	if c.UserID != 0 || c.ProviderID != 0 {
		c.save()
		w.WriteHeader(http.StatusCreated)
	}
}

func (c *Order) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	orders, err := c.get(toInt(r.URL.Query().Get("id")))
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(orders))
}

type Getter interface {
	get(int) (interface{}, error)
}

func getHandler(g Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders, err := g.get(toInt(r.URL.Query().Get("id")))
		if err != nil {
			vErr := errorHandler{Code: "not_found", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(orders))
	}
}

type Issue struct {
	ID         int  `json:"id" db:"id"`
	OrderID    int  `json:"order_id" db:"order_id"`
	IsResolved bool `json:"is_resolved" db:"is_resolved"`
	getHandler func(g Getter) http.HandlerFunc
	db         *sqlx.DB
}

func (c *Issue) save() error {

	c.db.Exec(stmt)
	tx := c.db.MustBegin()
	tx.NamedExec("insert into issues(order_id, created_at, is_resolved) values(:order_id, :created_at, :is_resolved)", c)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in cart.save: TX: %v", err)
		return err
	}
	return nil

}

func (c *Issue) get() ([]Issue, error) {
	var issues []Issue

	c.db.Exec(stmt)
	if err := c.db.Select(&issues, "select * from issues"); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Issue) getIssuesHandler(w http.ResponseWriter, r *http.Request) {
	i, err := c.get()
	if err != nil {
		vErr := errorHandler{Code: "issue_not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.Write(marshal(&i))
}

func (c *Issue) createIssueHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	unmarshal(b, c)
	if err := c.save(); err != nil {
		vErr := errorHandler{Code: "db_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// User struct in the system
type User struct {
	ID                 int    `db:"id" json:"id"`
	Username           string `db:"username" json:"username"`
	Mobile             string `db:"mobile" json:"mobile"`
	db                 *sqlx.DB
	CreatedAt          sql.NullTime `db:"created_at" json:"created_at"`
	Password           string       `db:"password" json:"password"`
	VerificationNumber string       `db:"verification_number" json:"verification_number"`
	IsProvider         bool         `db:"is_provider" json:"is_provider"`
}

func (u *User) generatePassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	u.Password = string(hash)
	return err
}

func (u *User) verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (u *User) saveUser() error {

	u.db.Exec(stmt)
	tx := u.db.MustBegin()
	tx.NamedExec("insert into users(username, mobile, password) values(:username, :mobile, :password)", u)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	}
	return nil
}

func (u *User) getUser(username string) error {

	tx := u.db.MustBegin()
	tx.Get(u, "select * from users where username = $1", username)
	if err := tx.Commit(); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	}
	return nil
}

func (u *User) getProviders() ([]User, error) {
	var users []User
	tx := u.db.MustBegin()

	tx.Get(&users, "select * from users where is_provider = 1")
	if err := tx.Commit(); err != nil {
		log.Printf("Error in DB: %v", err)
		return users, err
	}
	return users, nil
}

func (u *User) getProvidersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := u.getProviders()
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(users))
}

func (u *User) getUserHandler(w http.ResponseWriter, r *http.Request) {

	err := u.getUser(r.URL.Query().Get("id"))
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(u))
}

func (u *User) login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	unmarshal(b, u)
	pass := u.Password
	log.Printf("User model is: %#v", u)
	if err := u.getUser(u.Username); err != nil {
		vErr := errorHandler{Code: "user_not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	log.Printf("Passwords are: %v, %v", u.Password, pass)
	if err := u.verifyPassword(u.Password, pass); err != nil {
		vErr := errorHandler{Code: "wrong_password", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.Write(marshal(u))

}

func (u *User) registerHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	unmarshal(b, u)
	u.generatePassword(u.Password)
	err = u.saveUser()
	if err != nil {
		vErr := errorHandler{Code: "db_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.Write(marshal(u))
	return
}
