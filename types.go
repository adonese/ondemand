package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type genericMap = map[string]string
type result = map[string]interface{}

var successfulCreated = make(map[string]interface{})

const (
	NA = iota
	Pending
	Accepted
	Rejected
)

type Pagination struct {
	Count  int         `json:"count"`
	Result interface{} `json:"result"`
}

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
	w.Header().Add("content-type", "application/json; charset=utf-8")
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
	w.Header().Add("content-type", "application/json; charset=utf-8")
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

//CustomerProvider type for joining results in orders output
type CustomerProvider struct {
	CustomerName   *string `json:"customer_name" db:"customer_name"`
	ProviderName   *string `json:"provider_name" db:"provider_name"`
	CustomerMobile *string `json:"customer_mobile" db:"customer_mobile"`
	ProviderMobile *string `json:"provider_mobile" db:"provider_mobile"`
}

//Order
type Order struct {
	ID          int          `json:"id" db:"id"`
	UserID      int          `json:"user_id" db:"user_id"`
	ProviderID  int          `json:"provider_id" db:"provider_id"`
	Status      bool         `json:"status" db:"status"`
	CreatedAt   sql.NullTime `db:"created_at" json:"created_at"`
	OrderUUID   string       `json:"uuid" db:"uuid"`
	db          *sqlx.DB
	IsPending   bool   `json:"is_pending" db:"is_pending"`
	Description string `json:"description" db:"description"`
	Category    int    `json:"category" db:"category"`
	Provider    *User  `json:"provider,omitempty"`
	UserProfile *User  `json:"user,omitempty"` //*nice*
	CustomerProvider
}

func (c *Order) verify() bool {
	if c.Description == "" || c.Category == 0 || c.UserID == 0 {
		return false
	}
	return true
}

func (c *Order) token() string {
	c.OrderUUID = uuid.New().String()
	return c.OrderUUID
}

func (c *Order) get(id int) ([]Order, error) {
	var services []Order

	c.db.Exec(stmt)
	if err := c.db.Select(&services, "select * from orders"); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *Order) getProviders(id int) ([]Order, error) {
	var services []Order

	if err := u.db.Select(&services, "select * from orders where provider_id = ?", id); err != nil {
		return nil, err
	}

	return services, nil
}

func (c *Order) getProvidersX(id int) ([]Order, error) {
	var services []Order
	var user User

	c.db.Exec(stmt)

	if err := u.db.Select(&services, "select * from orders where provider_id = ?", id); err != nil {
		return nil, err
	}

	for idx, v := range services {
		if err := u.db.Get(&user, "select * from users where id = ?", v.UserID); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		log.Printf("user is: %v", user)
		services[idx].Provider = &user
	}

	return services, nil
}

func (c *Order) getUsers(id int) ([]Order, error) {
	var services []Order

	c.db.Exec(stmt)
	if err := c.db.Select(&services, `select o.*, customers.fullname as customer_name, customers.mobile as customer_mobile, providers.fullname as provider_name, providers.mobile as provider_mobile from orders o
	inner join users customers on customers.id = o.user_id
	inner join users as providers on providers.id = o.provider_id
	where customers.id = ? order by is_pending desc`, id); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *Order) updateUUID() ([]Order, error) {
	var services []Order

	c.db.Exec(stmt)
	if _, err := c.db.NamedExec("Update orders set status = :status, is_pending = :is_pending where uuid = :id", map[string]interface{}{"status": c.Status, "id": c.OrderUUID, "is_pending": c.IsPending}); err != nil {
		log.Printf("Error in updateUUID: %v", err)
		return nil, err
	}
	return services, nil
}

func (c *Order) setProvider() ([]Order, error) {
	var services []Order

	c.db.Exec(stmt)
	if _, err := c.db.NamedExec("Update orders set provider_id = :provider where uuid = :id", map[string]interface{}{"provider": c.ProviderID, "id": c.OrderUUID}); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *Order) save() error {
	c.db.Exec(stmt)

	if _, err := c.db.NamedExec("insert into orders(user_id, created_at, provider_id, status, uuid, description, category) values(:user_id, :created_at, :provider_id, :status, :uuid, :description, :category)", c); err != nil {
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
	if c.UserID != 0 {
		c.save()
		w.WriteHeader(http.StatusCreated)
	}
}

func (c *Order) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	/*
		{"count": 12, "result": [{order_id, provider_id, order,}]}
	*/
	w.Header().Add("content-type", "application/json; charset=utf-8")
	var orders []Order
	var err error
	id := r.URL.Query().Get("id")
	userID := r.URL.Query().Get("user_id")

	if id != "" {
		orders, err = c.getProvidersX(toInt(id))
		if err != nil {
			vErr := errorHandler{Code: "not_found", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
	} else if userID != "" {
		orders, err = c.getUsers(toInt(userID))
		if err != nil {
			vErr := errorHandler{Code: "not_found", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
	}
	res := Pagination{Count: len(orders), Result: orders}
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(res))
}

func (c *Order) requestHandler(w http.ResponseWriter, r *http.Request) {
	/*
		todo marshall and then return id (for tracking and further inquiries)
	*/

	w.Header().Add("content-type", "application/json; charset=utf-8")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, c)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}

	if r.Method == "PUT" {
		res, err := c.updateUUID()
		if err != nil {
			res := errorHandler{Code: "bad_request", Message: "Error in request"}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshal(&res))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(marshal(res))
		return
	}
	// user_id, uuid
	// user_id, provider_id, uuid
	// user_id, provider_id, uuid
	if ok := c.verify(); !ok {
		res := errorHandler{Code: "bad_request", Message: "Fields are missing"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}
	t := c.token()
	c.save()
	maps := make(map[string]genericMap)
	tt := genericMap{
		"uuid": t,
		"time": time.Now().String(),
	}
	maps["result"] = tt
	res, _ := json.Marshal(maps)
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}

func (c *Order) setProviderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		todo marshall and then return id (for tracking and further inquiries)
	*/

	w.Header().Add("content-type", "application/json; charset=utf-8")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, c)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}

	res, err := c.setProvider()
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(marshal(res))
	return
}

func (c *Order) updateOrder(w http.ResponseWriter, r *http.Request) {
	/*
		todo marshall and then return id (for tracking and further inquiries)
	*/
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, c)
	if err != nil {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}

	if c.OrderUUID == "" || c.ProviderID == 0 {
		res := errorHandler{Code: "bad_request", Message: "Error in request"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(&res))
		return
	}
	c.setProvider()
	w.WriteHeader(http.StatusOK)

	res := result{
		"result": c,
	}
	w.Write(marshal(res))
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
	ID                 int     `db:"id" json:"id"`
	Username           string  `db:"username" json:"username"`
	Fullname           *string `db:"fullname" json:"fullname"`
	Mobile             string  `db:"mobile" json:"mobile"`
	db                 *sqlx.DB
	CreatedAt          *time.Time `db:"created_at" json:"created_at"`
	Password           string     `db:"password" json:"password"`
	VerificationNumber *string    `db:"verification_number" json:"verification_number"`
	IsProvider         bool       `db:"is_provider" json:"is_provider"`
	Services           []int      `json:"services"`
	IsActive           *bool      `json:"is_active" db:"is_active"`
}

func (u *User) generatePassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	u.Password = string(hash)
	return err
}

func (u *User) cleanInput() {
	u.Username = strings.TrimSpace(u.Username)

}

func (u *User) valid() bool {
	if u.Password == "" || u.Username == "" || u.Mobile == "" {
		return false
	}
	return true
}

func (u *User) verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

var getNames = make(map[string]bool)

func (u *User) getTags() (string, []interface{}, error) {
	var ss sq.UpdateBuilder
	stmt := sq.Update("users")
	ss = stmt
	if u.Username != "" {
		ss = stmt.Set("username", u.Username)
	}
	if u.Password != "" {
		ss = ss.Set("password", u.Password)
	}
	if u.Fullname != nil {
		ss = ss.Set("fullname", u.Fullname)
	}
	if u.Mobile != "" {
		ss = ss.Set("mobile", u.Mobile)
	}
	if u.IsActive != nil {
		ss = ss.Set("is_active", u.IsActive)
	}

	ss = ss.Where("id = ?", u.ID)

	return ss.ToSql()
}

func (u *User) updateUser() error {

	q, args, err := u.getTags()
	if err != nil {
		return err
	}

	if _, err := u.db.Exec(q, args...); err != nil {
		return err
	}
	return nil
}

func (u *User) saveUser() error {

	u.db.Exec(stmt)

	if n, err := u.db.NamedExec("insert into users(username, mobile, password, fullname, is_provider) values(:username, :mobile, :password, :fullname, :is_provider)", u); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	} else {
		id, _ := n.LastInsertId()
		u.ID = int(id)
	}
	return nil
}

func (u *User) saveUserTX() error {

	u.db.Exec(stmt)

	tx := u.db.MustBegin()
	rr, err := tx.NamedExec("insert into users(username, mobile, password, fullname, is_provider) values(:username, :mobile, :password, :fullname, :is_provider)", u)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, v := range u.Services {
		id, _ := rr.LastInsertId()

		if _, err := tx.NamedExec("insert into userservices(user_id, service_id) values(:user, :provider)", map[string]interface{}{"user": id, "provider": v}); err != nil {
			tx.Rollback()
			return err
		}

	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (u *User) getUser(username string) error {

	//TODO update all queries to use Get for single result and select from multiple results
	if err := u.db.Get(u, "select * from users where username = $1", username); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	}
	return nil
}

func (u *User) getPassword(id int) (string, error) {

	//TODO update all queries to use Get for single result and select from multiple results
	if err := u.db.Get(u, "select * from users where id = $1", id); err != nil {
		log.Printf("Error in DB: %v", err)
		return "", err
	}
	return u.Password, nil
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

type Provider struct {
	Score int `json:"score" db:"score"`
	db    *sqlx.DB
	User
}

func (p *Provider) getProviders() ([]Provider, error) {
	var users []Provider

	if err := p.db.Select(&users, "select * from users where is_provider = 1"); err != nil {
		log.Printf("Error in DB: %v", err)
		return nil, err
	}
	return users, nil
}

func (p *Provider) getProvidersWithScoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	data, err := p.getProviders()
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	mData := make(map[string]interface{})
	mData["result"] = data
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(mData))
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
	u.cleanInput()
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

func (u *User) updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	if err := unmarshal(b, u); err != nil {
		vErr := errorHandler{Code: "marshalling_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(vErr))
		return
	}

	u.cleanInput()
	if r.Method == "PUT" {
		if u.ID == 0 {
			vErr := errorHandler{Code: "empty_user_id", Message: "Empty user id"}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		if err := u.updateUser(); err != nil {
			vErr := errorHandler{Code: "update_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func (u *User) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	if err := unmarshal(b, u); err != nil {
		vErr := errorHandler{Code: "marshalling_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(vErr))
		return
	}

	u.cleanInput()
	if r.Method == "PUT" {
		if u.ID == 0 {
			vErr := errorHandler{Code: "empty_user_id", Message: "Empty user id"}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		if err := u.updateUser(); err != nil {
			vErr := errorHandler{Code: "update_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// this is for POST only requests
	if !u.valid() {
		vErr := errorHandler{Code: "bad_request", Message: "empty request fields"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	u.generatePassword(u.Password)

	if err := u.saveUser(); err != nil {
		vErr := errorHandler{Code: "db_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(u))

}

func (u *User) saveProviders(user int, provider int) error {
	if _, err := u.db.NamedExec("insert into userservices(user_id, service_id) values(:user, :provider", map[string]interface{}{"user": user, "provider": provider}); err != nil {
		return err
	}
	return nil
}

type Pushes struct {
	ID          int    `json:"id" db:"id"`
	UserID      int    `json:"user_id" db:"user_id"`
	OneSignalID string `json:"onesignal_id" db:"onesignal_id"`
	db          *sqlx.DB
}

func (p *Pushes) check() error {
	if p.UserID == 0 || p.OneSignalID == "" {
		return errors.New("empty_fields")
	}
	return nil
}

func (p *Pushes) save() error {
	if _, err := p.db.NamedExec("insert into pushes(user_id, onesignal_id) values(:user_id, :signal_id)", map[string]interface{}{"user_id": p.UserID, "signal_id": p.OneSignalID}); err != nil {
		return err
	}
	return nil
}

func (p *Pushes) getSignalID(id int) error {

	if err := p.db.Get(p, "select * from pushes where user_id = ?", id); err != nil {
		return err
	}
	return nil
}

func (p *Pushes) saveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		verr := errorHandler{Code: "missing_fields", Message: "Some fields are missing"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))

		return
	}
	defer r.Body.Close()
	err = json.Unmarshal(d, p)

	if err != nil {
		verr := errorHandler{Code: "marshalling_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))

		return
	}

	if err := p.check(); err != nil {
		verr := errorHandler{Code: "empty_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))

		return
	}

	if err := p.save(); err != nil {
		verr := errorHandler{Code: "missing_fields", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	res := successfulCreated
	res["result"] = true
	w.Write(marshal(res))

}

func (p *Pushes) getIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	var id string

	if id = r.URL.Query().Get("id"); id == "" {
		verr := errorHandler{Code: "missing_fields", Message: "id is missing in url query"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if err := p.getSignalID(toInt(id)); err != nil {
		verr := errorHandler{Code: "db_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))

		return
	}
	w.Write(marshal(p))
}

type Suggestion struct {
	ID         int    `json:"id" db:"id"`
	Suggestion string `json:"suggestion" db:"suggestion"`
	db         *sqlx.DB
}

func (s *Suggestion) save() error {
	if _, err := s.db.Exec("insert into suggestions(suggestion) value(?)", s.Suggestion); err != nil {
		return err
	}
	return nil
}

func (s *Suggestion) check() bool {
	if s.Suggestion != "" {
		return true
	}
	return false
}

func (s *Suggestion) saveHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("content-type", "application/json; charset=utf-8")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		verr := errorHandler{Code: "empty_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}
	marshal(b)
	if ok := s.check(); !ok {
		verr := errorHandler{Code: "empty_complain", Message: "Empty complain text"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}
	if err := s.save(); err != nil {
		verr := errorHandler{Code: "db_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}
	w.WriteHeader(http.StatusOK)
}

var upgrader = websocket.Upgrader{} // use default options
var upgrader2 = websocket.Upgrader{}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		data <- message

		select {
		case <-accept:
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, []byte("it worked"))
			if err != nil {
				log.Println("write:", err)

			}
		case <-time.After(1 * time.Minute):
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, []byte("timeout"))
			if err != nil {
				log.Println("write:", err)

			}
			c.Close()
			return
		}
		// close(accept)

	}
}
