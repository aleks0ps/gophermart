APP := gophermart
MODULE := github.com/aleks0ps/gophermart

.PHONY: all
all: build

go.mod:
	go mod init $(MODULE)

.PHONY: build
build:
	go build -o ./cmd/$(APP)/$(APP) ./cmd/$(APP)


ADDR := localhost:8080

test:
	curl -X POST -H "Content-Type: application/json" -d '{"login": "alexey", "password": "123"}' $(ADDR)/api/user/register
