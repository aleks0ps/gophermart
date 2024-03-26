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

test:
	curl --cookie-jar cookie.txt -X POST -H "Content-Type: application/json" -d '{"login": "alexey", "password": "123"}' $(ADDR)/api/user/register
	curl --cookie-jar cookie.txt -X POST -H "Content-Type: application/json" -d '{"login": "alexey", "password": "123"}' $(ADDR)/api/user/login

test1:
	curl -b cookie.txt -X POST -H "Content-Type: text/plain" -d '123456789007' $(ADDR)/api/user/orders

test2:
	curl -b cookie.txt $(ADDR)/api/user/orders

