package service

import (
	"encoding/json"
	"errors"
	"net/http"

	mycookie "github.com/aleks0ps/gophermart/internal/app/cookie"
	myerror "github.com/aleks0ps/gophermart/internal/app/error"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

func (s *Service) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
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
	orders, err := s.DB.GetWithdrawals(r.Context(), &user)
	if err != nil {
		if errors.Is(err, myerror.ErrNoWithdrawals) {
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusNoContent, nil)
			return
		}
		myhttp.WriteError(&w, http.StatusInternalServerError, err)
		return
	}
	res, err := json.Marshal(orders)
	if err != nil {
		myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusInternalServerError, nil)
		return
	}
	myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusOK, res)
}
