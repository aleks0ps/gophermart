package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aleks0ps/gophermart/internal/app/gerror"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

// impletemnts http.HandlerFunc interface
func (s *Service) Register(w http.ResponseWriter, r *http.Request) {
	stype := r.Header.Get("Content-Type")
	// Ignore non-JSON data
	if myhttp.GetContentTypeCode(stype) != myhttp.CTypeJSON {
		myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusBadRequest, nil)
		return
	}
	// JSON
	var user storage.User
	var buf bytes.Buffer
	// Read data from request body
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		myhttp.WriteError(&w, http.StatusBadRequest, err)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &user); err != nil {
		myhttp.WriteError(&w, http.StatusBadRequest, err)
		return
	}
	if err := s.DB.Register(r.Context(), &user); err != nil {
		if errors.Is(err, gerror.LoginAlreadyTaken) {
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusConflict, nil)
			return
		}
		myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
		return
	}
	myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusOK, nil)
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login " + r.Method)
}
