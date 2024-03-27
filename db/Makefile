URI := postgres://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable 

.PHONY: all
all: init

.PHONY: init
init:
	psql $(URI) -f db.sql -w gophermart
