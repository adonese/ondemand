package main




var stmt = `
create table users (
	id integer primary key,
	username text unique,
	mobile text unique,
	is_provider bool default false,
);

`


func main(){

}