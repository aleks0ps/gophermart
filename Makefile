APP := gophermart
MODULE := github.com/aleks0ps/gophermart

.PHONY: all
all: build

go.mod:
	go mod init $(MODULE)

.PHONY: build
build:
	go build -o ./cmd/$(APP)/$(APP) ./cmd/$(APP)


ADDR := localhost:8088

login:
	curl --cookie-jar cookie.txt -X POST -H "Content-Type: application/json" -d '{"login": "alexey", "password": "123"}' $(ADDR)/api/user/register
	curl --cookie-jar cookie.txt -X POST -H "Content-Type: application/json" -d '{"login": "alexey", "password": "123"}' $(ADDR)/api/user/login

JSON := "Content-Type: application/json"
TEXT := "Content-Type: text/plain" 
ORDERS := 640516806186665 \
	  3554827372 \
	  83366120 \
	  83130581404 \
	  2142256141423 \
	  46744502157 \
	  5272232 

.ONESHELL:
test: login 
	curl -v -b cookie.txt http://localhost:8088/api/user/orders
	@#curl -v -b cookie.txt http://localhost:8088/api/user/balance
	@#curl -v -b cookie.txt -X POST -H $(TEXT) -d '83130581404' $(ADDR)/api/user/orders
	@#curl -v -b cookie.txt $(ADDR)/api/user/balance
	@#curl -v -b cookie.txt -X POST -H $(JSON) -d '{"order": "83130581404","sum": 411.78 }' $(ADDR)/api/user/balance/withdraw
	@#curl -v -b cookie.txt -X POST -H $(JSON) -d '{ "order": "7138742213177", "sum": 252.52 }' http://localhost:8088/api/user/balance/withdraw
