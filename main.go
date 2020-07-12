package main

import "net/http"




var stmt = `
create table users (
	id integer primary key,
	username text unique,
	mobile text unique,
	is_provider bool default false,
);

`


func main(){
	mux := http.NewServeMux()
	mux.Handle("/", Auth(http.HandlerFunc(login)))
	http.ListenAndServe(":8080", mux)
}

