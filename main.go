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
	order_status integer not null default 0,
	is_pending integer not null default 1,
	created_at DATE DEFAULT (datetime('now','localtime'))
);

create table issues (
	id integer primary key,
	is_resolved integer not null default 0,
	order_id integer,
	created_at DATE DEFAULT (datetime('now','localtime'))
);

create table pushes (
	id integer not null PRIMARY key,
	user_id integer not null,
	onesignal_id text not null,
	FOREIGN key (user_id) references users(id)
);

create table userservices(
	user_id integer not null,
	service_id integer not null,
	FOREIGN key (user_id) REFERENCES users(id),
	FOREIGN key (service_id) REFERENCES services(id),
	unique(user_id, service_id)
	);


create table suggestions (
		id integer not null PRIMARY key,
		suggestion text not null,
		created_at DATE DEFAULT (datetime('now','localtime'))
	);
`

var (
	u    User
	o    Order
	i    Issue
	s    Service
	p    Provider
	Sugg Suggestion
	pus  Pushes
)

var db, _ = getDB("test.db")
var accept = make(chan struct{ id int })

func init() {

	u.db = db
	o.db = db
	i.db = db
	s.db = db
	p.db = db
	pus.db = db
	Sugg.db = db
}

func main() {

	hub := newHub()
	go hub.run()

	mux := http.NewServeMux()
	mux.Handle("/", Auth(http.HandlerFunc(login)))
	mux.Handle("/login", http.HandlerFunc(u.login))
	mux.Handle("/register", http.HandlerFunc(u.registerHandler))
	mux.Handle("/user/update", http.HandlerFunc(u.updateHandler))
	mux.Handle("/services", http.HandlerFunc(s.getHandler))
	mux.Handle("/services/problems", http.HandlerFunc(s.serviceDetailsHandler))
	mux.Handle("/new_order", http.HandlerFunc(o.saveHandler))
	mux.Handle("/orders/new", http.HandlerFunc(o.saveHandler))
	mux.Handle("/orders", http.HandlerFunc(o.getOrdersHandler))
	mux.Handle("/orders/id", http.HandlerFunc(o.byUUID))
	mux.Handle("/orders/request", http.HandlerFunc(o.requestHandler))
	mux.Handle("/orders/provider", http.HandlerFunc(o.setProviderHandler))
	mux.Handle("/orders/accept", http.HandlerFunc(o.updateOrder))
	mux.Handle("/providers", http.HandlerFunc(p.getProvidersWithScoreHandler))
	// mux.Handle("/orders/status")
	mux.Handle("/issues", http.HandlerFunc(i.getIssuesHandler))
	mux.Handle("/issues/new", http.HandlerFunc(i.createIssueHandler))
	mux.Handle("/push/save", http.HandlerFunc(pus.saveHandler))
	mux.Handle("/push/get", http.HandlerFunc(pus.getIDHandler))

	mux.Handle("/suggestion", http.HandlerFunc(Sugg.saveHandler))
	mux.Handle("/ws2", http.HandlerFunc(p.ws))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.ListenAndServe(":6662", mux)
}
