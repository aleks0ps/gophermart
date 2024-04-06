package service

import (
	"encoding/json"
	"net/http"

	mycookie "github.com/aleks0ps/gophermart/internal/app/cookie"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

func (s *Service) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Validate user
	if err := mycookie.ValidateCookie(r); err != nil {
		s.Logger.Errorln(err.Error())
		myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusUnauthorized, nil)
		return
	}
	login, err := mycookie.GetCookie(r, "id")
	if err != nil {
		s.Logger.Errorln(err.Error())
		// 401
		myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusUnauthorized, nil)
		return
	}
	user := storage.User{Login: login}
	balance, err := s.DB.GetBalance(r.Context(), &user)
	if err != nil {
		myhttp.WriteError(&w, http.StatusInternalServerError, err)
		return
	}
	res, err := json.Marshal(balance)
	if err != nil {
		myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusInternalServerError, nil)
		return
	}
	myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusOK, res)
}
