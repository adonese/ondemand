package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var stmt = `
create table users (
	id integer primary key,
	username text unique,
	fullname text not null,
	mobile text unique,
	is_provider integer not null default 0,
	password text,
	verification_number text,
	created_at DATE DEFAULT (datetime('now','localtime')),
	score integer default 0,
	description text
);

create table services (
	id integer primary key,
	name text
);

create table images (
	id integer primary key,
	uuid text not null,
	path text not null
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

create table views(
	user_id integer not null,
	count integer not null,
	FOREIGN key (user_id) REFERENCES users(id),
	unique(user_id)
	);

create table suggestions (
		id integer not null PRIMARY key,
		suggestion text not null,
		created_at DATE DEFAULT (datetime('now','localtime'))
	);
`

var (
	userField User
	o         Order
	i         Issue
	s         Service
	p         Provider
	Sugg      Suggestion
	pus       Pushes
	image     Image
)

var db, _ = getDB("test.db")
var accept = make(chan struct{ id int }, 256)

func init() {

	userField.db = db
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

	r := mux.NewRouter()

	r.Handle("/login", http.HandlerFunc(userField.login))
	r.Handle("/register", http.HandlerFunc(userField.registerHandler))
	r.Handle("/otp", http.HandlerFunc(userField.otpHander))
	r.Handle("/otp/check", http.HandlerFunc(userField.otpCheckHandler))
	r.Handle("/user/update", http.HandlerFunc(userField.updateHandler))
	r.Handle("/services", http.HandlerFunc(s.getHandler))
	r.Handle("/services/problems", http.HandlerFunc(s.serviceDetailsHandler))
	r.Handle("/new_order", http.HandlerFunc(o.saveHandler))
	r.Handle("/orders/new", http.HandlerFunc(o.saveHandler))
	r.Handle("/orders", http.HandlerFunc(o.getOrdersHandler))
	r.Handle("/orders/id", http.HandlerFunc(o.byUUID))
	r.Handle("/orders/request", http.HandlerFunc(o.requestHandler))
	r.Handle("/orders/provider", http.HandlerFunc(o.setProviderHandler))
	r.Handle("/orders/accept", http.HandlerFunc(o.updateOrder))
	r.Handle("/view", http.HandlerFunc(userField.incrHandler))

	r.Handle("/providers", http.HandlerFunc(p.getProvidersWithScoreHandler))

	// r.Handle("/orders/status")

	r.Handle("/issues", http.HandlerFunc(i.getIssuesHandler))
	r.Handle("/issues/new", http.HandlerFunc(i.createIssueHandler))
	r.Handle("/push/save", http.HandlerFunc(pus.saveHandler))
	r.Handle("/push/get", http.HandlerFunc(pus.getIDHandler))

	r.Handle("/image/save", http.HandlerFunc(image.storeHandler))
	r.Handle("/image/get", http.HandlerFunc(image.getFileHandler))

	r.Handle("/suggestion", http.HandlerFunc(Sugg.saveHandler))
	r.Handle("/ws2", http.HandlerFunc(p.ws))
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	r.Handle("/admin/providers", http.HandlerFunc(userField.getProvidersHandler))
	r.Handle("/admin/providers/{id}", http.HandlerFunc(userField.getByIDHandler))
	r.Handle("/admin/orders", http.HandlerFunc(o.adminOrdersHandler))
	r.Handle("/admin/orders/{id}", http.HandlerFunc(o.byID))
	r.Handle("/admin/login", http.HandlerFunc(userField.loginAdmin))
	r.Handle("/password_reset", http.HandlerFunc(userField.PasswordReset))
	r.Handle("/success", http.HandlerFunc(userField.success))

	r.Handle("/fail", http.HandlerFunc(userField.fail))
	r.Handle("/otp/change_password", http.HandlerFunc(userField.otpCheckHandler))
	r.Handle("/_otp", http.HandlerFunc(userField.otpPage))
	// r.Handle("/admin/stats", http.HandlerFunc(o.stats))
	r.Handle("/terms/", http.HandlerFunc(userField.terms))
	r.Handle("/terms", http.HandlerFunc(userField.terms))

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)

	//TODO handle position in orders/request

	corsHandler := cors.New(cors.Options{ExposedHeaders: []string{"X-Total-Count"}, AllowedMethods: []string{"GET", "POST", "PUT"}}).Handler(r)
	log.Fatal(http.ListenAndServe(":6662", corsHandler))
}
