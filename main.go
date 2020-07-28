package main

import "net/http"




var stmt = `
create table users (
	id integer primary key,
	username text unique,
	mobile text unique,
	is_provider bool default false,
	password text,
	verification_number text
);


create table services (
	id integer primary key,
	name text,
);

create table orders (
	id integer primary key,
	user_id integer,
	provider_id integer,
	status bool
);

create table issues (
	id integer primary key,
	is_resolved bool,
	order_id integer
);
`

var (
	u User
	o Order
	i Issue
	s Service
)

var db, _ = getDB("test.db")

func init(){

	u.db = db
	o.db = db
	i.db = db
	s.db = db
}

func main(){

	
	mux := http.NewServeMux()
	mux.Handle("/", Auth(http.HandlerFunc(login)))
	mux.Handle("/login", http.HandlerFunc(u.login))
	mux.Handle("/register", http.HandlerFunc(u.registerHandler))
	mux.Handle("/services", http.HandlerFunc(s.getHandler))
	mux.Handle("/new_order", http.HandlerFunc(o.saveHandler))
	mux.Handle("/orders", http.HandlerFunc(o.getOrdersHandler))
	mux.Handle("/orders/status", nil)
	mux.Handle("/issues", http.HandlerFunc(i.getIssuesHandler))
	mux.Handle("/issues/new", http.HandlerFunc(i.createIssueHandler))

	http.ListenAndServe(":8080", mux)
}

