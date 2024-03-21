package middleware

import (
	"log"
	"net/http"

	m "github.com/go-chi/chi/middleware"
)

type DummyIO int

func (e DummyIO) Write(p []byte) (int, error) {
	e += DummyIO(len(p))
	return len(p), nil
}

func DisableDefaultLogger() func(next http.Handler) http.Handler {
	var dummy DummyIO
	dummyLogFormatter := m.DefaultLogFormatter{Logger: log.New(dummy, "", log.LstdFlags), NoColor: true}
	dummyLogger := m.RequestLogger(&dummyLogFormatter)
	return dummyLogger
}
