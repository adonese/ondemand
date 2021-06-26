package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	im "image"

	sq "github.com/Masterminds/squirrel"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type genericMap = map[string]string
type result = map[string]interface{}

type idName struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

var successfulCreated = make(map[string]interface{})

const (
	SMS_GATEWAY = "http://www.oursms.net/api/sendsms.php"
)
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

type OrdersUsers struct {
	Order
	User
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

type En struct {
	IsEn bool
	Err  errorHandler
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
	ID          int        `json:"id,omitempty" db:"id"`
	UserID      int        `json:"user_id,omitempty" db:"user_id"`
	ProviderID  int        `json:"provider_id,omitempty" db:"provider_id"`
	Status      bool       `json:"status,omitempty" db:"status"`
	CreatedAt   *time.Time `db:"created_at,omitempty" json:"created_at"`
	OrderUUID   string     `json:"uuid,omitempty" db:"uuid"`
	db          *sqlx.DB
	IsPending   bool   `json:"is_pending,omitempty" db:"is_pending"`
	Description string `json:"description,omitempty" db:"description"`
	Category    int    `json:"category,omitempty" db:"category"`
	Provider    *User  `json:"provider,omitempty"`
	UserProfile *User  `json:"user,omitempty"` //*nice*
	CustomerProvider
	CustomerName string `json:"customer_name,omitempty" db:"customer_name"`
	ProviderName string `json:"provider_name,omitempty" db:"provider_name"`
}

type AdminOrder struct {
	ID           int        `json:"id,omitempty" db:"id"`
	Status       bool       `json:"status,omitempty" db:"status"`
	CreatedAt    *time.Time `db:"created_at,omitempty" json:"created_at"`
	OrderUUID    string     `json:"uuid,omitempty" db:"uuid"`
	db           *sqlx.DB
	IsPending    bool   `json:"is_pending,omitempty" db:"is_pending"`
	Description  string `json:"description,omitempty" db:"description"`
	Category     int    `json:"category,omitempty" db:"category"`
	CustomerName string `json:"customer_name,omitempty" db:"customer_name"`
	ProviderName string `json:"provider_name,omitempty" db:"provider_name"`
	ProviderCity string `json:"provider_city,omitempty" db:"provider_city"`
	CustomerCity string `json:"customer_city,omitempty" db:"customer_city"`
}

//Image stored in fs
type Image struct {
	ID   int    `json:"id" db:"id"`
	UUID string `json:"uuid" db:"uuid"`
	Data string `json:"data"`
}

func (i *Image) init(uuid string) bool {
	i.UUID = uuid
	return true
}

func (i *Image) store() (string, error) {

	var ext string

	if strings.HasPrefix(i.Data, "data:image/png") {
		ext = "png"
	} else {
		ext = "jpg"
	}

	if strings.HasPrefix(i.Data, "data:image/") {
		// temp = i.Data
		index := strings.Index(i.Data, ",")
		i.Data = i.Data[index+1:]
	}

	r, err := base64.StdEncoding.DecodeString(i.Data)
	if err != nil {
		panic(err)
	}

	var img im.Image
	f, err := os.OpenFile("data/"+i.UUID+"."+ext, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}

	defer f.Close() // This fucked us so badly!

	switch ext {
	case "png":
		img, err = png.Decode(bytes.NewReader(r))
		if err != nil {
			return "", err
		}
		png.Encode(f, img)
		return "data/" + i.UUID + "." + ext, nil
	case "jpg":
		img, err = jpeg.Decode(bytes.NewReader(r))
		if err != nil {
			return "", err
		}
		if err != nil {
			return "", err
		}
		jpeg.Encode(f, img, nil)
		return "data/" + i.UUID + "." + ext, nil
	}

	return "data/" + i.UUID + "." + ext, nil

}

func (c *Order) all(sort string) []AdminOrder {
	var orders []AdminOrder
	if err := c.db.Select(&orders, `select orders.id, orders.category, orders.created_at, orders.description, orders.uuid, u.fullname as customer_name, uu.fullname as provider_name, uu.city as provider_city, u.city as customer_city from orders
	join users u on u.id = orders.user_id
	join users uu on uu.id = orders.provider_id`); err != nil {
		log.Printf("error in orders: %v", err)
		return nil
	}
	return orders
}

func (c *Order) stats() []Order {
	var orders []Order
	c.db.Select(&orders, "select * from orders")
	return orders
}

func (c *Order) statsHandler(w http.ResponseWriter, r *http.Request) {

	var users []User
	if r.URL.Query().Get("type") == "providers" {
		// return providers with most used providers
		c.db.Select(&users, "select * from orders ")
		return
	}

	var orders []Order
	c.db.Select(&orders, "select * from orders")
	return
}

func (i *Image) getBytes(path string) ([]byte, error) {
	if d, err := ioutil.ReadFile("data/" + path); err != nil {
		return nil, err
	} else {

		return d, nil

	}
}

func (i *Image) getString(uuid string) (string, error) {

	if d, err := ioutil.ReadFile("data/" + uuid + ".png"); err != nil {
		return "", err
	} else {
		data := base64.RawStdEncoding.EncodeToString(d)
		return data, nil

	}
}

func (i *Image) getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
}

func (i *Image) getFileHandler(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("path")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f, err := i.getBytes(name)
	if err != nil {
		log.Printf("The error is: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s := strconv.Itoa(len(f))
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", s)
	io.Copy(w, bytes.NewReader(f))
}

func (i *Image) storeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		verr := errorHandler{Code: "empty_request", Message: err.Error()}
		w.Write(marshal(verr))
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, i); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		verr := errorHandler{Code: "marshaling_error", Message: err.Error()}
		w.Write(marshal(verr))
		return
	}
	// i.init()

	if _, err = i.store(); err != nil {
		log.Printf("unable to store file: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		verr := errorHandler{Code: "server_error", Message: err.Error()}
		w.Write(marshal(verr))
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (c *Order) verify() bool {
	if c.Category == 0 || c.UserID == 0 {
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

	if err := c.db.Select(&services, "select * from orders where provider_id = ?", id); err != nil {
		return nil, err
	}

	return services, nil
}

func (c *Order) getProvidersX(id int) ([]Order, error) {
	var services []Order
	var user User

	c.db.Exec(stmt)

	if err := c.db.Select(&services, "select * from orders where provider_id = ?", id); err != nil {
		return nil, err
	}

	for idx, v := range services {
		if err := c.db.Get(&user, "select * from users where id = ?", v.UserID); err != nil {
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

	if _, err := c.db.NamedExec("insert into orders(user_id, provider_id, status, uuid, description, category) values(:user_id, :provider_id, :status, :uuid, :description, :category)", c); err != nil {
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
	w.Header().Add("X-Total-Count", toString(len(orders)))
	w.WriteHeader(http.StatusOK)

	w.Write(marshal(res))
}

func (c *Order) adminOrdersHandler(w http.ResponseWriter, r *http.Request) {

	// sort and filtering
	// filter=%7B%7D&order=ASC&page=1&perPage=10&sort=uuid

	sort := r.URL.Query().Get("sort")

	w.Header().Add("content-type", "application/json; charset=utf-8")

	res := c.all(sort)
	w.Header().Add("X-Total-Count", toString(len(res)))
	w.WriteHeader(http.StatusOK)

	w.Write(marshal(res))
}

func (c *Order) byUUID(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("content-type", "application/json; charset=utf-8")

	id := r.URL.Query().Get("uuid")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var res OrdersUsers

	if err := c.db.Get(&res, `select u.fullname, u.mobile, o.*  from users u
	join orders o on o.user_id = u.id where o.uuid = ?`, id); err != nil {
		verr := errorHandler{Code: "db_err", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(marshal(res))
}

func (c *Order) byID(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("content-type", "application/json; charset=utf-8")

	vars := mux.Vars(r)

	id := toInt(vars["id"])

	var res OrdersUsers

	if err := c.db.Get(&res, `select u.fullname, u.mobile, o.*  from users u
	join orders o on o.user_id = u.id where o.id = ?`, id); err != nil {
		verr := errorHandler{Code: "db_err", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(marshal(res))
}

func (c *Order) requestHandler(w http.ResponseWriter, r *http.Request) {

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
		vErr := errorHandler{Code: "db_err", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type UserService struct {
	User
	db *sqlx.DB
}

// User struct in the system
type User struct {
	ID                 int        `db:"id" json:"id"`
	Username           string     `db:"username" json:"username"`
	Fullname           *string    `db:"fullname" json:"fullname"`
	Mobile             string     `db:"mobile" json:"mobile"`
	CreatedAt          *time.Time `db:"created_at" json:"created_at"`
	Password           string     `db:"password" json:"password"`
	VerificationNumber *string    `db:"verification_number" json:"verification_number"`
	IsProvider         bool       `db:"is_provider" json:"is_provider"`
	Services           []int      `json:"services"`
	IsActive           *bool      `json:"is_active" db:"is_active"`
	Score              int        `json:"score" db:"score"`
	Description        *string    `json:"description" db:"description"`
	Channel            *int       `json:"channel"`
	Image              *string    `json:"image"`
	ImagePath          *string    `json:"path" db:"path"`
	ServiceName        []idName   `json:"service_names"`
	IsAdmin            bool       `json:"is_admin" db:"is_admin"`
	City               string     `json:"city" db:"city"`
	Whatsapp           *string    `json:"whatsapp" db:"whatsapp"`
	Latitude           *float64   `json:"latitude" db:"latitude"`
	Longitude          *float64   `json:"longitude" db:"longitude"`
	MobileChecked      *bool      `json:"mobile_checked" db:"mobile_checked"`
	DeviceID           *string    `json:"device_id" db:"device_id"`
	IsDisabled         *bool      `json:"is_disabled" db:"is_disabled"`
	db                 *sqlx.DB
}

func fixNumbers(text string) string {
	//ToEnglishDigits Converts all Persian digits in the string to English digits.
	//٠١٢٣٤٥٦٧٨٩
	var checker = map[string]string{
		"٠": "0",
		"١": "1",
		"٢": "2",
		"٣": "3",
		"٤": "4",
		"٥": "5",
		"٦": "6",
		"٧": "7",
		"٨": "8",
		"٩": "9",
	}
	re := regexp.MustCompile("[٠-٩]+")
	out := re.ReplaceAllFunc([]byte(text), func(s []byte) []byte {
		out := ""
		ss := string(s)
		for _, ch := range ss {
			o := checker[string(ch)]
			out = out + o
		}
		return []byte(out)
	})
	return string(out)

}

func (u *User) generatePassword(password string) error {
	if len(u.Password) > 12 {
		return errors.New("wrong_password_length")
	}
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

func (u *User) getAdminTags() (string, []interface{}, error) {
	var ss sq.UpdateBuilder
	stmt := sq.Update("users")
	ss = stmt

	if u.Password != "" { // memory BUG

		log.Printf("the password in getTags is: %v", u.Password)
		u.generatePassword(u.Password)
		ss = ss.Set("password", u.Password)
	}
	// test for nullable here
	if u.Fullname != nil {
		if *u.Fullname != "" {
			ss = ss.Set("fullname", u.Fullname)
		}

	}
	if u.Mobile != "" {
		ss = ss.Set("mobile", u.Mobile)
		ss = ss.Set("username", u.Mobile)
	}
	if u.IsActive != nil {
		ss = ss.Set("is_active", u.IsActive)
	}
	if u.Description != nil {
		if *u.Description != "" {
			ss = ss.Set("description", u.Description)
		}

	}
	if u.Whatsapp != nil {
		ss = ss.Set("whatsapp", u.Whatsapp)
	}
	if u.ImagePath != nil {
		log.Printf("the getTags image path is: %v", *u.ImagePath)
		ss = ss.Set("path", u.ImagePath)
	}
	if u.City != "" {

		ss = ss.Set("city", u.City)
	}
	if u.Score != 0 {
		ss = ss.Set("score", u.Score)
	}
	if u.Latitude != nil {
		ss = ss.Set("latitude", u.Latitude)
	}
	if u.Longitude != nil {
		ss = ss.Set("longitude", u.Longitude)
	}
	if u.MobileChecked != nil {
		ss = ss.Set("mobile_checked", u.MobileChecked)
	}
	if u.IsDisabled != nil {
		ss = ss.Set("is_disabled", u.IsDisabled)
	}

	ss = ss.Set("is_provider", u.IsProvider)

	ss = ss.Where("id = ?", u.ID)

	return ss.ToSql()
}

func (u *User) getTags() (string, []interface{}, error) {
	var ss sq.UpdateBuilder
	stmt := sq.Update("users")
	ss = stmt

	if u.Password != "" { // memory BUG

		log.Printf("the password in getTags is: %v", u.Password)
		u.generatePassword(u.Password)
		ss = ss.Set("password", u.Password)
	}
	// test for nullable here
	if u.Fullname != nil {
		if *u.Fullname != "" {
			ss = ss.Set("fullname", u.Fullname)
		}

	}
	if u.Mobile != "" {
		ss = ss.Set("mobile", u.Mobile)
		ss = ss.Set("username", u.Mobile)
	}
	if u.IsActive != nil {
		ss = ss.Set("is_active", u.IsActive)
	}
	if u.Description != nil {
		if *u.Description != "" {
			ss = ss.Set("description", u.Description)
		}

	}
	if u.ImagePath != nil {
		log.Printf("the getTags image path is: %v", *u.ImagePath)
		ss = ss.Set("path", u.ImagePath)
	}
	if u.City != "" {

		ss = ss.Set("city", u.City)
	}
	if u.Score != 0 {
		ss = ss.Set("score", u.Score)
	}
	if u.Latitude != nil {
		if *u.Latitude != 0 {
			ss = ss.Set("latitude", u.Latitude)
		}

	}
	if u.Longitude != nil {
		if *u.Longitude != 0 {
			ss = ss.Set("longitude", u.Longitude)
		}

	}

	ss = ss.Where("id = ?", u.ID)

	return ss.ToSql()
}

func (u *User) saveImage() error {
	var err error
	if u.Image != nil {
		log.Print("we should not be here")
		img := &Image{}
		imID := uuid.New().String()
		img.init(imID)
		img.Data = *u.Image
		var path string
		if path, err = img.store(); err != nil {
			log.Printf("error in saving data: %v", err)
			return err
		} else {
			log.Printf("the image path is: %v", path)
			u.ImagePath = &path
			return nil
		}

	} else {
		return errors.New("Image not found")
	}
	return nil

}

func (u *User) updateUserAdmin() error {

	log.Print(u.saveImage())
	q, args, err := u.getAdminTags()
	if err != nil {
		log.Printf("errors are: %v", err)
		return err
	}
	log.Printf("the sql query is: %s", q)
	// log.Printf("the new value is: %#v", args[1].(string))

	// Store image HERE!

	log.Printf("the image path in db is: %v", u.ImagePath)
	if _, err := u.db.Exec(q, args...); err != nil {

		log.Printf("Errors are: %v", err)
		return err
	}
	return nil
}

func (u *User) updateUser() error {

	log.Print(u.saveImage())
	q, args, err := u.getTags()
	if err != nil {
		log.Printf("errors are: %v", err)
		return err
	}
	log.Printf("the sql query is: %s", q)
	// log.Printf("the new value is: %#v", args[1].(string))

	// Store image HERE!

	log.Printf("the image path in db is: %v", u.ImagePath)
	if _, err := u.db.Exec(q, args...); err != nil {

		log.Printf("Errors are: %v", err)
		return err
	}
	return nil
}

func (u *User) saveUser() error {

	u.db.Exec(stmt)

	if n, err := u.db.NamedExec("insert into users(username, mobile, password, fullname, is_provider, path, city) values(:username, :mobile, :password, :fullname, :is_provider, :path, :city)", u); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	} else {
		id, _ := n.LastInsertId()
		u.ID = int(id)
	}
	return nil
}

func (u *User) saveProvider() error {

	u.db.Exec(stmt)

	if n, err := u.db.NamedExec("insert into users(username, mobile, password, fullname, is_provider, path, description, is_active, city, whatsapp, latitude, longitude) values(:username, :mobile, :password, :fullname, :is_provider, :path, :description, :is_active, :city, :whatsapp, :latitude, :longitude)", u); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	} else {
		id, _ := n.LastInsertId()
		u.ID = int(id)
	}
	return nil
}

func (u *User) savePush() error {
	if _, err := u.db.Exec("insert into pushes(user_id, onesignal_id) values(?, ?)", u.ID, u.DeviceID); err != nil {
		log.Printf("error in push: %v", err)
		return err
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
	if err := u.db.Get(u, "select * from users where username = ?", username); err != nil {
		log.Printf("Error in DB: %v", err)
		return err
	}
	return nil
}

func (u *User) getServices(username string) ([]int, error) {

	//TODO update all queries to use Get for single result and select from multiple results
	//todo add this
	// source="post_id" reference="posts"
	// "posts": [{"service_id": 1}]
	var dest []int
	if err := u.db.Select(&dest, `select us.service_id from users u
	join userservices us on us.user_id = u.id where u.username = ?`, username); err != nil {
		return nil, err
	}
	return dest, nil
}

func (u *User) fetchServices(username string) ([]idName, error) {
	var dest []idName
	if err := u.db.Select(&dest, `select us.service_id id, s.name from users u
	join userservices us on us.user_id = u.id
	join services s on s.id = us.service_id
	where u.username = ?`, username); err != nil {
		return nil, err
	}
	return dest, nil
}

func (u *User) changePassword(mobile string, rawPassword string) bool {

	u.generatePassword(rawPassword)
	log.Printf("the new password is: %v - previous password is: %v", u.Password, rawPassword)
	log.Printf("the mobile is: %v", u.Mobile)
	if res, err := u.db.Exec("update users set password = ? where mobile = ?", u.Password, mobile); err != nil {
		log.Printf("Error in password creation: %v", err)
		log.Printf("Error in password creation: %v", err)
		id, _ := res.LastInsertId()
		rows, _ := res.RowsAffected()
		log.Printf("Error in password creation: ids affected: %v - rows affected: %v", id, rows)
		return false
	} else {

		rows, _ := res.RowsAffected()
		if rows < 1 {
			return false
		}
		return true
	}

}

func handleMobile(m string) string {

	if strings.HasPrefix(m, "966") {

		d := m[3:]
		if strings.HasPrefix(d, "0") {
			return "966" + d[1:]
		}
		return m
	}

	if strings.HasPrefix(m, "00966") {
		d := m[5:]
		if strings.HasPrefix(d, "0") {
			return "966" + d[1:]
		}
		return m[2:]
	} else if strings.HasPrefix(m, "0") {
		return "966" + m[1:]
	} else {
		return "966" + m
	}

}

func (u *User) getPassword(id int) (string, error) {

	//TODO update all queries to use Get for single result and select from multiple results
	if err := u.db.Get(u, "select * from users where id = $1", id); err != nil {
		log.Printf("Error in DB: %v", err)
		return "", err
	}
	return u.Password, nil
}

func (u *User) sendSms(otp string) error {

	/*
		http://www.oursms.net/api/sendsms.php?username=SEARCHFORME&password=a@2092002&message=3344&numbers=00966556789882&sender=SEARCHFORMY&unicode=E&return=json*/
	if u.Mobile == "" {
		return errors.New("mobile_not_provided")
	}

	mm := handleMobile(u.Mobile)

	v := url.Values{}
	v.Add("username", "SEARCHFORME")
	v.Add("password", "a@2092002")
	v.Add("sender", "SEARCHFORMY")
	v.Add("numbers", mm)
	v.Add("message", otp)
	v.Add("return", "json")
	v.Add("unicode", "E")

	uri := SMS_GATEWAY + "?" + v.Encode()

	log.Print(uri)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Get(uri)
	res.Close = true

	if err != nil {
		log.Printf("The error is: %v", err)
	}
	log.Printf("The response body is: %v", res)
	return nil
}

func (u *User) otpPassword(w http.ResponseWriter, r *http.Request) {
	// verify mobile and otp
	mobile := r.URL.Query().Get("mobile")
	otp := r.URL.Query().Get("otp")
	password := r.URL.Query().Get("password")

	if ok := validateOTP(otp, mobile); !ok {
		http.Error(w, "wrong_otp", http.StatusBadRequest)
		return
	}

	if ok := u.changePassword(mobile, password); !ok {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}

}

func (u *User) otpHander(w http.ResponseWriter, r *http.Request) {
	var mobile string
	user := &User{db: u.db}
	var noCacheHeaders = map[string]string{
		"Expires":         "-1",
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}

	// Set our NoCache headers
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}

	if mobile = r.URL.Query().Get("mobile"); mobile == "" {
		verr := errorHandler{Code: "mobile_not_found", Message: "Mobile not found"}

		if strings.Contains(r.Referer(), "_otp") {
			http.Redirect(w, r, "/fail", http.StatusPermanentRedirect)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if otp, err := generateOTP(mobile); err != nil {
		verr := errorHandler{Code: "otp_error", Message: "OTP error"}
		if strings.Contains(r.Referer(), "_otp") {
			http.Redirect(w, r, "/fail", http.StatusPermanentRedirect)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(marshal(verr))
		return
	} else {
		// ACTUALLY sending an otp
		user.Mobile = mobile
		err := user.sendSms(otp)
		if err != nil {
			if strings.Contains(r.Referer(), "_otp") {
				http.Redirect(w, r, "/fail", http.StatusPermanentRedirect)
				return
			}
			verr := errorHandler{Code: "otp_error", Message: err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(marshal(verr))
			return
		}

		log.Printf("Otp res: %v", err)
		log.Printf("the referrer == :%v", r.Referer())
		log.Printf("The lang is: %v", r.URL.Query().Get("lang"))
		if strings.Contains(r.Referer(), "_otp") {
			if strings.Contains(r.Referer(), "lang=en") {
				http.Redirect(w, r, "/success?lang=en", http.StatusPermanentRedirect)
				return
			}
			http.Redirect(w, r, "/success", http.StatusPermanentRedirect)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(map[string]interface{}{"result": otp}))
		return
	}
}

func (u *User) verifyOTPhandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("content-type", "application/json")
	var mobile string
	var otp string

	var verr errorHandler
	if mobile = r.URL.Query().Get("mobile"); mobile == "" {

		verr = errorHandler{Code: "mobile_not_provided", Message: "Mobile not provided"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if otp = r.URL.Query().Get("otp"); otp == "" {
		verr = errorHandler{Code: "otp_not_found", Message: otpErrEn}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if ok := validateOTP(otp, mobile); !ok {
		verr = errorHandler{Code: "otp_error", Message: "OTP Error"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}
	//OTP is verified. NOW set the db

	if _, err := u.db.Exec("update users set mobile_checked = ? where mobile = ?", 1, mobile); err != nil {
		verr = errorHandler{Code: "db_err", Message: "Couldn't able to amend mobile checked"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u *User) otpCheckHandler(w http.ResponseWriter, r *http.Request) {
	var mobile string
	var otp string
	var password string
	var json int
	var isEn bool

	var tmpl *template.Template
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		isEn = true
		log.Print("i'm forever here")
		tmpl = template.Must(template.ParseFiles("password/fail_en.html"))
	} else {
		log.Print("wy the fuck i'm here")
		tmpl = template.Must(template.ParseFiles("password/fail.html"))
	}

	user := &User{db: u.db}

	json = toInt(r.URL.Query().Get("json"))
	log.Printf("the json is: %v", json)
	var verr errorHandler

	if mobile = r.URL.Query().Get("mobile"); mobile == "" {
		if isEn {
			verr = errorHandler{Code: "mobile_not_provided", Message: "Mobile not provided"}
		} else {
			verr = errorHandler{Code: "mobile_not_provided", Message: "لم يتم ادخال رقم الهاتف"}

		}

		if json != 1 {
			tmpl.Execute(w, En{isEn, verr})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if otp = r.URL.Query().Get("otp"); otp == "" {
		if isEn {
			verr = errorHandler{Code: "otp_not_found", Message: otpErrEn}
		} else {
			verr = errorHandler{Code: "otp_not_found", Message: "خطأ في رمز الOTP"}
		}

		if json != 1 {
			tmpl.Execute(w, En{isEn, verr})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(marshal(verr))
		return
	}

	if password = r.URL.Query().Get("password"); password == "" {
		if isEn {
			verr = errorHandler{Code: "wrong_password", Message: wrongPasswordEn}
		} else {
			verr = errorHandler{Code: "wrong_password", Message: "لم يتم ادخال الرمز السري الجديد"}
		}
		if json != 1 {
			tmpl.Execute(w, En{isEn, verr})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(marshal(verr))
		return
	}

	log.Printf("OTP is: %v, mobile is: %v", otp, mobile)

	if ok := validateOTP(otp, mobile); !ok {
		if isEn {
			verr = errorHandler{Code: "otp_error", Message: otpErrEn}

		} else {
			verr = errorHandler{Code: "otp_error", Message: "خطأ في الOTP"}

		}
		if json != 1 {
			tmpl.Execute(w, En{isEn, verr})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(marshal(verr))
		return
	}

	if ok := user.changePassword(mobile, password); !ok {
		if isEn {
			verr = errorHandler{Code: "otp_error", Message: "Error in changing password. Try again."}

		} else {
			verr = errorHandler{Code: "otp_error", Message: "خطأ في تعديل كلمة المرور. الرجاء المحاولة مرة آخرى"}

		}

		if json != 1 {
			tmpl.Execute(w, En{isEn, verr})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}

	if isEn {
		verr = errorHandler{Code: "successfull", Message: closePromptEn}

	} else {
		verr = errorHandler{Code: "successfull", Message: "تم تعديل كلمة المرور بنجاح. الرجاء اغلاق هذه النافذة وادخال كلمة المرور الجديدة في ابحث لي."}
	}

	if json != 1 {
		tmpl.Execute(w, En{isEn, verr})
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(marshal(verr))

}

func (u *User) getProviders() ([]User, error) {
	var users []User

	// now we ought to fix this one
	if err := u.db.Select(&users, "select * from users where is_provider = 1"); err != nil {

		log.Printf("Error in DB: %v", err)
		return nil, err
	}

	return users, nil
}

func (u *User) getProvidersWithViews(page int, query string) ([]UserViews, int, error) {
	var users []UserViews

	// now we ought to fix this one
	if query == "" {
		if err := u.db.Select(&users, `select u.*, v.count from users u
		left join views v on v.user_id = u.id where u.is_provider = 1 order by u.id desc`); err != nil {
	
			log.Printf("Error in DB: %v", err)
			return []UserViews{}, 0, err
		}
	}else{
		like := "%" + query + "%"
		if err := u.db.Select(&users, `select u.*, v.count from users u
		left join views v on v.user_id = u.id where u.is_provider = 1 and u.mobile like ? or u.fullname like ? order by u.id desc`, like, like); err != nil {
	
			log.Printf("Error in DB: %v", err)
			return []UserViews{}, 0, err
		}else{
			log.Printf("-- why am i here!", users)
			return users, len(users), nil
		}
	}

	log.Print("why am i here😤╰(*°▽°*)╯")

	var count int
	if err := u.db.Get(&count, "select count(*) from users where is_provider = 1"); err != nil {

		log.Printf("Error in DB: %v", err)
		return []UserViews{}, count, err
	}
	// NOW here is the thing, starts at 10, and limit to the rest
	if len(users) <= 10 {
		return users, count, nil
	} else {
		return users[page*10 : page*10+11], count, nil
	}

	// this is so fucked up
	for _, v := range users {
		if v.Whatsapp != nil {
			*v.Whatsapp = fixNumbers(*v.Whatsapp) // BUG(adonese)
		}

	}

	return users, 0, nil
}

func (u *User) getUsers(page int) ([]User, int, error) {
	var users []User

	log.Printf("The page is: %v", page)

	// now we ought to fix this one
	if err := u.db.Select(&users, "select * from users where is_provider = 0 order by id desc"); err != nil {

		log.Printf("Error in DB: %v", err)
		return nil, 0, err
	}
	var count int
	if err := u.db.Get(&count, "select count(*) from users where is_provider = 0"); err != nil {

		log.Printf("Error in DB: %v", err)
		return nil, count, err
	}
	// NOW here is the thing, starts at 10, and limit to the rest
	if len(users) <= 10 {
		return users, count, nil
	} else {
		return users[page*10 : page*10+11], count, nil
	}

}

func (u *User) getProvidersByID(id int) (User, error) {
	var user User

	// now we ought to fix this one
	if err := u.db.Get(&user, "select * from users where id = ?", id); err != nil {

		log.Printf("Error in DB: %v", err)
		return user, err
	}
	return user, nil
}

//getProvidersHandler
// http://localhost:3000/#/providers?filter=%7B%7D&order=ASC&page=1&perPage=10&sort=id
func (u *User) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	if r.Method == "GET" {
		// compute for start and end and get results accordingly.

		e := r.URL.Query().Get("_end")
		end, _ := strconv.Atoi(e)

		if end == 0 {
			end = 10
		}
		page := end / 10

		users, count, err := u.getUsers(page)
		if err != nil {
			vErr := errorHandler{Code: "not_found", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		w.Header().Add("X-Total-Count", toString(count))
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(users))
	}

}

type UserViews struct {
	User
	Count *int `json:"count" db:"count"`
}

//getProvidersHandler
// http://localhost:3000/#/providers?filter=%7B%7D&order=ASC&page=1&perPage=10&sort=id
func (u *User) getProvidersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	if r.Method == "GET" {

		e := r.URL.Query().Get("_end")
		end, _ := strconv.Atoi(e)

		if end == 0 {
			end = 10
		}
		page := end / 10
		
		q := r.URL.Query().Get("q")
		log.Printf("The query is: %v", q)

		users, count, err := u.getProvidersWithViews(page, q)
		if err != nil {
			vErr := errorHandler{Code: "not_found", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		w.Header().Add("X-Total-Count", toString(count))
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(users))
	}
	// POSTing a new user
	if r.Method == "POST" {
		req, err := ioutil.ReadAll(r.Body)

		if err != nil {
			log.Printf("the error is: %v", err)
			vErr := errorHandler{Code: "empty_body", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		defer r.Body.Close()
		if err := json.Unmarshal(req, u); err != nil {
			log.Printf("the error is: %v", err)
			vErr := errorHandler{Code: "marshalling_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}

		u.generatePassword(u.Password)
		if err := u.saveUser(); err != nil {
			log.Printf("the error is: %v", err)
			vErr := errorHandler{Code: "db_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(marshal(u))
		return
	}

}

func (u *User) incrView(id int) error {
	if _, err := u.db.Exec(`INSERT OR REPLACE INTO views
	VALUES (?,
	  COALESCE(
		(SELECT count FROM views
		   WHERE user_id=? ),
		0) + 1);`, id, id); err != nil {
		return err
	}
	return nil
}

func (u *User) incrHandler(w http.ResponseWriter, r *http.Request) {
	if id := r.URL.Query().Get("id"); id == "" {
		verr := errorHandler{Code: "user_id_not_provided", Message: "ID not provided"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	} else {
		u.incrView(toInt(id))
		w.WriteHeader(http.StatusOK)
		return
	}

}

func (u *User) getByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")
	vars := mux.Vars(r)

	id := toInt(vars["id"])
	if r.Method == "PUT" {
		req, err := ioutil.ReadAll(r.Body)
		if err != nil {
			verr := errorHandler{Code: "marshalling_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshal(verr))
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(req, u)
		if err != nil {
			verr := errorHandler{Code: "marshalling_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshal(verr))
			return
		}
		log.Printf("The marshal is: %v", string(req))
		if err := u.updateUserAdmin(); err != nil {
			verr := errorHandler{Code: "db_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshal(verr))
			return
		}

		return
	} else if r.Method == "DELETE" {
		if _, err := u.db.Exec("delete from users where id = ?", id); err != nil {
			verr := errorHandler{Code: "db_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshal(verr))
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var services []int

	u.db.Select(&services, "select service_id from userservices where user_id = ?", id)

	users, err := u.getProvidersByID(id)
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	users.Services = services
	w.Header().Add("X-Total-Count", toString(1))
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(users))
}

type Provider struct {
	Score2    int     `json:"score2" db:"score2"`
	Haversine float64 `json:"distance"`
	Count     *int    `json:"count" db:"count"`
	db        *sqlx.DB
	User
}

func (p *Provider) getProviders(id int) ([]Provider, error) {
	var users []Provider

	// here is the real shit
	// ok check is_active = 1
	if err := p.db.Select(&users, `select u.*, v.count from users u
	left join views v on v.user_id = u.id
	join userservices us on us.user_id = u.id where us.service_id = ? and is_active = 1 and is_disabled = 0 and u.latitude not null and u.longitude not null
	order by score desc`, id); err != nil {
		log.Printf("Error in DB: %v", err)
		return nil, err
	}

	for _, v := range users {
		if v.Whatsapp != nil {
			*v.Whatsapp = fixNumbers(*v.Whatsapp)
		}

	}

	return users, nil
}

func (p *Provider) byUUID(id string) (OrdersUsers, error) {
	var res OrdersUsers
	log.Printf("the uuid is: %v", id)

	if err := p.db.Get(&res, `select u.fullname, u.mobile, o.* from users u
	join orders o on o.user_id = u.id where o.uuid = ?`, id); err != nil {

		return OrdersUsers{}, err
	}
	return res, nil

}

func (p *Provider) byID(id int) (User, error) {
	var user User

	// skip is_provider check here
	if err := p.db.Get(&user, "select * from users where id = ?", id); err != nil {
		log.Printf("Error in DB: %v", err)
		return user, err
	}
	return user, nil
}

func (p *Provider) getProvidersWithScoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	var id string
	var latitude, longitude float64

	var lat, long string

	if id = r.URL.Query().Get("id"); id == "" {
		verr := errorHandler{Code: "not_found", Message: "ID not found"}
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(verr))
		return
	}

	// for latitude
	if lat = r.URL.Query().Get("latitude"); lat == "" {
		verr := errorHandler{Code: "not_found", Message: "Latutide not found"}
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(verr))

		return
	}

	// for longitude
	if long = r.URL.Query().Get("longitude"); long == "" {
		verr := errorHandler{Code: "not_found", Message: "Longitude not found"}
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(verr))

		return
	}
	latitude, _ = strconv.ParseFloat(lat, 64)
	longitude, _ = strconv.ParseFloat(long, 64)

	data, err := p.getProviders(toInt(id))
	if err != nil {
		vErr := errorHandler{Code: "not_found", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	for _, vv := range data {
		if vv.Whatsapp != nil {
			if strings.HasPrefix(*vv.Whatsapp, "966") || strings.HasPrefix(*vv.Whatsapp, "00966") || strings.HasPrefix(*vv.Whatsapp, "+966") {
				continue
			} else {
				*vv.Whatsapp = "+966" + *vv.Whatsapp
			}
		}
	}

	var prov []Provider
	for _, v := range data {
		if v.Latitude != nil && v.Longitude != nil { // this is an uncessary check. We did it in db.
			v.Haversine = haverSine(*v.Latitude, latitude, *v.Longitude, longitude)
			prov = append(prov, v)
		}
	}

	sort.SliceStable(prov, func(i, j int) bool {
		return prov[i].Haversine < prov[j].Haversine
	})

	// just go in here and any value that is more than 60 make it LESS.

	for _, v := range prov {
		if v.Haversine > 60 {
			v.Haversine = 100
		}
	}

	mData := make(map[string]interface{})
	mData["result"] = prov
	w.Header().Add("X-Total-Count", toString(len(prov)))
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

func (u *User) isAuthorized() bool {
	if u.Channel == nil {
		if !u.IsProvider {
			return true
		} else {
			return false
		}
	}

	if u.IsProvider && *u.Channel == 0 {
		return true
	}
	if !u.IsProvider && *u.Channel == 1 {
		return true
	}
	return false
}

func (u *User) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	user := &User{db: u.db}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	unmarshal(b, user)
	user.cleanInput()
	pass := user.Password
	log.Printf("User model is: %#v", u)

	if err := user.getUser(user.Username); err != nil {
		vErr := errorHandler{Code: "user_not_found", Message: "الحساب غير مسجّل"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	log.Printf("Passwords are: %v, %v", user.Password, pass)
	// if ok := user.isAuthorized(); !ok {
	// 	vErr := errorHandler{Code: "access_denied", Message: "Not authorized"}
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write(vErr.toJson())
	// 	return
	// }
	if err := user.verifyPassword(user.Password, pass); err != nil {
		vErr := errorHandler{Code: "wrong_password", Message: "كلمة المرور غير صحيحة"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	if user.MobileChecked == nil {
		vErr := errorHandler{Code: "otp_not_confirmed", Message: "OTP not confirmed"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	log.Printf("Description is: %v", user.Description)
	user.Image = nil
	w.Write(marshal(user))

}

//PasswordReset API for payment
func (u *User) PasswordReset(w http.ResponseWriter, r *http.Request) {
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		tmpl := template.Must(template.ParseFiles("password/layout_en.html"))
		tmpl.Execute(w, En{IsEn: true})
	} else {
		tmpl := template.Must(template.ParseFiles("password/layout.html"))
		tmpl.Execute(w, En{})
	}

}

//PasswordReset API for payment
func (u *User) success(w http.ResponseWriter, r *http.Request) {
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		tmpl := template.Must(template.ParseFiles("password/success_en.html"))
		tmpl.Execute(w, En{IsEn: true})
	} else {
		tmpl := template.Must(template.ParseFiles("password/success.html"))
		tmpl.Execute(w, En{IsEn: false})
	}

}

func (u *User) terms(w http.ResponseWriter, r *http.Request) {
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		tmpl := template.Must(template.ParseFiles("password/terms_en.html"))
		tmpl.Execute(w, En{IsEn: true})
	} else {
		tmpl := template.Must(template.ParseFiles("password/terms.html"))
		tmpl.Execute(w, En{IsEn: false})
	}

}

//PasswordReset API for payment
func (u *User) fail(w http.ResponseWriter, r *http.Request) {
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		tmpl := template.Must(template.ParseFiles("password/fail_en.html"))
		tmpl.Execute(w, En{IsEn: true})
	} else {
		tmpl := template.Must(template.ParseFiles("password/fail.html"))
		tmpl.Execute(w, En{IsEn: false})
	}

}

//PasswordReset API for payment
func (u *User) otpPage(w http.ResponseWriter, r *http.Request) {
	if lang := r.URL.Query().Get("lang"); lang == "en" {
		tmpl := template.Must(template.ParseFiles("password/otp_en.html"))
		tmpl.Execute(w, En{IsEn: true})
	} else {
		tmpl := template.Must(template.ParseFiles("password/otp.html"))
		tmpl.Execute(w, En{IsEn: true})
	}

}

func (u *User) loginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	//TODO add is_admin to users table
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
	// if ok := u.isAuthorized(); !ok {
	// 	vErr := errorHandler{Code: "access_denied", Message: "Not authorized"}
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write(vErr.toJson())
	// 	return
	// }

	if err := u.verifyPassword(u.Password, pass); err != nil {
		vErr := errorHandler{Code: "wrong_password", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	// add services here
	u.Services, _ = u.getServices(u.Username)
	u.ServiceName, _ = u.fetchServices(u.Username)

	w.Write(marshal(u))

}

func (u *User) updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	user := &User{db: u.db}

	user.Password = "" //workaround previous bugs
	var id string
	id = r.URL.Query().Get("id")
	if id == "" {
		id = "0"
	}

	user.ID = toInt(id)

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	if err := unmarshal(b, &user); err != nil {
		vErr := errorHandler{Code: "marshalling_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(vErr))
		return
	}

	log.Printf("the data is: %#v", user.ID)
	log.Printf("the data is: %#v", user.Password)

	user.cleanInput()
	if r.Method == "PUT" {
		if user.ID == 0 {
			vErr := errorHandler{Code: "empty_user_id", Message: "Empty user id"}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		if err := user.updateUser(); err != nil {
			vErr := errorHandler{Code: "update_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}

		log.Printf("user services are: %v", user.Services)
		log.Printf("the ID is: %v", id)

		if len(user.Services) > 0 {
			log.Printf("In services")
			if _, err := user.db.Exec("delete from userservices where user_id = ?", toInt(id)); err != nil {
				log.Printf("error in removing previous services: %v", err)
			}

			for _, service := range user.Services {
				u.saveProviders(toInt(id), service)
			}

		}

		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)

}

func (us *User) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json; charset=utf-8")

	user := &User{db: us.db}

	var id string
	id = r.URL.Query().Get("id")
	if id == "" {
		id = "0"
	}

	user.ID = toInt(id)

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		vErr := errorHandler{Code: "bad_request", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}

	if err := unmarshal(b, user); err != nil {
		vErr := errorHandler{Code: "marshalling_error", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(vErr))
		return
	}

	log.Printf("the data is: %#v", user)
	user.cleanInput()
	if r.Method == "PUT" {
		if user.ID == 0 {
			vErr := errorHandler{Code: "empty_user_id", Message: "Empty user id"}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		if err := user.updateUser(); err != nil {
			vErr := errorHandler{Code: "update_error", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// this is for POST only requests
	if !user.valid() {
		vErr := errorHandler{Code: "bad_request", Message: "empty request fields"}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(vErr.toJson())
		return
	}
	user.generatePassword(user.Password)

	if user.Image != nil {
		log.Print("we should not be here")
		img := &Image{}
		imID := uuid.New().String()
		img.init(imID)
		img.Data = *user.Image
		var path string
		if path, err = img.store(); err != nil {
			log.Printf("error in saving data: %v", err)
		} else {
			user.ImagePath = &path
		}

	}

	// this code is not clean; should be fixed
	if user.IsProvider {
		if err := user.saveProvider(); err != nil {
			vErr := errorHandler{Code: "db_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
		// save the push
		user.savePush()
	} else {
		if err := user.saveUser(); err != nil {
			vErr := errorHandler{Code: "db_err", Message: err.Error()}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(vErr.toJson())
			return
		}
	}

	if user.Services != nil {
		for _, v := range user.Services {
			user.saveProviders(user.ID, v)
		}

	}
	user.Image = nil // we don't want to pollute the user with their image
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(user))

}

func (u *User) saveProviders(user int, service int) error {
	if _, err := u.db.Exec("insert into userservices(user_id, service_id) values(?, ?)", user, service); err != nil {
		log.Printf("Error in saving providers services: %v", err)
		return err
	}
	return nil
}

func (u *User) deleteServices(user int) error {
	if _, err := u.db.Exec("delete from userservices where user_id = ?", user); err != nil {
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
		verr := errorHandler{Code: "db_err", Message: err.Error()}
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
	return s.Suggestion != ""

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
		verr := errorHandler{Code: "db_err", Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(marshal(verr))
		return
	}
	w.WriteHeader(http.StatusOK)
}

var upgrader = websocket.Upgrader{} // use default options
var upgrader2 = websocket.Upgrader{}

func (p *Provider) ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		// close connection
		c.Close()
	}

	// get user info
	//

	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// get user info here
		user, err := p.byUUID(id)
		if err != nil {
			log.Printf("error in ws get by uuid: %v", err)
			c.Close()
			return
		}
		xbroadcast <- []byte(marshal(user))

		select {
		case data := <-accept:
			log.Printf("recv_accept: %d", data.id)
			provider, err := p.byID(data.id)
			if err != nil {
				log.Printf("error in retrieving user info: Err: %v", err)
			}
			log.Printf("The providers list is: %v", provider)
			err = c.WriteJSON(provider)
			if err != nil {
				log.Println("write:", err)

			}
		case <-time.After(10 * time.Second):
			log.Printf("recv_timeout: %s", message)
			verr := errorHandler{Code: "timeout", Message: "No providers found. Try again."}
			err = c.WriteJSON(verr)
			if err != nil {
				log.Println("write:", err)
			}
			c.Close()
			return
		}
		// close(accept)

	}
}

// getCity demonstrates the different geocoding services
func getCity(lat, lng float64) (string, error) {
	v := openstreetmap.GeocoderWithURL("https://nominatim.openstreetmap.org/reverse?accept-language=en&format=json&")
	d, err := v.ReverseGeocode(lat, lng)
	if err != nil {
		return "", err
	}
	return d.City, nil

}

func sendPushes(w http.ResponseWriter, r *http.Request) {
	// onesignal app key: cec9979b-de9c-47f6-ad3d-5ee314614a9f
	// api key: YTk4NDU4Y2YtYjZiNS00ZTY1LWFlMGYtZGRlYTUzZWQ5Zjc4
	/*

		// we use this request for signal stuff
			curl --include --request POST --header "Content-Type: application/json; charset=utf-8" -H "Authorization: Basic YTk4NDU4Y2YtYjZiNS00ZTY1LWFlMGYtZGRlYTUzZWQ5Zjc4" -d '{ "app_id": "cec9979b-de9c-47f6-ad3d-5ee314614a9f", "include_player_ids": ["9c512b96-8d7b-4942-9452-ba515016b8a8"], "channel_for_external_user_ids": "push", "data": {"foo": "bar"}, "contents": {"en": "it works shdeeeeed!"} }' https://onesignal.com/api/v1/notifications
	*/
}

var (
	wrongPassword   = "لم يتم ادخال الرمز السري الجديد"
	wrongPasswordEn = "New password not supplied"
	otpErr          = "خطأ في تعديل كلمة المرور. الرجاء المحاولة مرة آخرى"
	otpErrEn        = "Error in OTP. Please try again later"
	clostPrompt     = "تم تعديل كلمة المرور بنجاح. الرجاء اغلاق هذه النافذة وادخال كلمة المرور الجديدة في ابحث لي."
	closePromptEn   = "Password updated successfully. You may close this window and enter the new password in Search Me App"
)
