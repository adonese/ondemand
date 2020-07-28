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


func main(){
	mux := http.NewServeMux()
	mux.Handle("/", Auth(http.HandlerFunc(login)))
	http.ListenAndServe(":8080", mux)
}

