package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegister(t *testing.T) {
	ctx := context.TODO()
	user := storage.User{Login: "SomeUser", Password: "password"}
	userJSON, err := json.Marshal(&user)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	defaultDatabaseURI := "postgres://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable"
	defaultAccrualAddress := "localhost:8080"
	// lookup envs
	dbURI, exists := os.LookupEnv("DABASE_URI")
	if !exists {
		dbURI = defaultDatabaseURI
	}
	db, err := storage.NewPGStorage(ctx, dbURI, sugar)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	accrualURL, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if !exists {
		accrualURL = defaultAccrualAddress
	}
	svc := Service{Logger: sugar, DB: db, AccrualURL: accrualURL}
	handler := http.HandlerFunc(svc.Register)
	srv := httptest.NewServer(handler)
	r := resty.New().R()
	r.Method = http.MethodPost
	r.URL = srv.URL
	r.SetHeader("Content-Type", myhttp.STypeJSON)
	r.SetBody(userJSON)
	resp, err := r.Send()
	assert.NoError(t, err, "Error making HTTP request")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Status codes does not match")
	defer srv.Close()
}
