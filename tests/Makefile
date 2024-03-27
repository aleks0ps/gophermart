VER := v0.10.6
TEST_URL := https://github.com/Yandex-Practicum/go-autotests/releases/download/$(VER)/gophermarttest
BIN := gophermarttest

.PHONY: all
all: install

.PHONY: install
install:
	if ! test -f $(BIN); then \
	  wget $(TEST_URL); \
	fi
	chmod +x $(BIN)
