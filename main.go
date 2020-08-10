package main

import "net/http"




var stmt = `
create table users (
	id integer primary key,
	username text unique,
	fullname text not null,
	mobile text unique,
	is_provider integer not null default 0,
	password text,
	verification_number text,
	created_at DATE DEFAULT (datetime('now','localtime'))
);

create table services (
	id integer primary key,
	name text
);

create table orders (
	id integer primary key,
	user_id integer,
	provider_id integer,
	status integer not null default 0,
	uuid text,
	created_at DATE DEFAULT (datetime('now','localtime'))
);

create table issues (
	id integer primary key,
	is_resolved integer not null default 0,
	order_id integer,
	created_at DATE DEFAULT (datetime('now','localtime'))
);
`

var (
	u User
	o Order
	i Issue
	s Service
	p Provider
)

var db, _ = getDB("test.db")

func init(){

	u.db = db
	o.db = db
	i.db = db
	s.db = db
	p.db = db
}

func main(){

	
	mux := http.NewServeMux()
	mux.Handle("/", Auth(http.HandlerFunc(login)))
	mux.Handle("/login", http.HandlerFunc(u.login))
	mux.Handle("/register", http.HandlerFunc(u.registerHandler))
	mux.Handle("/services", http.HandlerFunc(s.getHandler))
	mux.Handle("/services/problems", http.HandlerFunc(s.serviceDetailsHandler))
	mux.Handle("/new_order", http.HandlerFunc(o.saveHandler))
	mux.Handle("/orders/new", http.HandlerFunc(o.saveHandler))
	mux.Handle("/orders", http.HandlerFunc(o.getOrdersHandler))
	mux.Handle("/orders/request", http.HandlerFunc(o.requestHandler))
	mux.Handle("/orders/accept", http.HandlerFunc(o.updateOrder))
	mux.Handle("/providers", http.HandlerFunc(p.getProvidersWithScoreHandler))
	// mux.Handle("/orders/status")
	mux.Handle("/issues", http.HandlerFunc(i.getIssuesHandler))
	mux.Handle("/issues/new", http.HandlerFunc(i.createIssueHandler))

	http.ListenAndServe(":6662", mux)
}

