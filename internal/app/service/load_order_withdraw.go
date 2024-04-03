package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	mycookie "github.com/aleks0ps/gophermart/internal/app/cookie"
	myerror "github.com/aleks0ps/gophermart/internal/app/error"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

func (s *Service) LoadOrderWithdraw(w http.ResponseWriter, r *http.Request) {
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
	stype := r.Header.Get("Content-Type")
	// Plain text
	if myhttp.GetContentTypeCode(stype) != myhttp.CTypeJSON {
		// 400
		myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusBadRequest, nil)
		return
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		myhttp.WriteError(&w, http.StatusBadRequest, err)
		return
	}
	// Parse json
	user := storage.User{Login: login}
	order := storage.Order{UploadedAt: time.Now().Format(time.RFC3339)}
	if err := json.Unmarshal(buf.Bytes(), &order); err != nil {
		myhttp.WriteError(&w, http.StatusBadRequest, err)
		return
	}
	err = s.DB.CheckWithdrawn(r.Context(), &user, &order)
	if err != nil {
		s.Logger.Errorln(err.Error())
		if errors.Is(err, myerror.InsufficientBalance) {
			// 402
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusPaymentRequired, nil)
		} else {
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusInternalServerError, nil)
		}
		return
	}
	// Store order to database
	if err := s.DB.LoadOrder(r.Context(), &user, &order); err != nil {
		s.Logger.Errorln(err.Error())
		if errors.Is(err, myerror.OrderLoaded) {
			// 200
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusOK, nil)
		} else if errors.Is(err, myerror.OrderInUse) {
			// 409
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusConflict, nil)
		} else {
			// 500
			myhttp.WriteError(&w, http.StatusInternalServerError, err)
		}
		return
	}
	err = s.DB.BalanceDecrease(r.Context(), &user, &order)
	if err != nil {
		s.Logger.Errorln(err.Error())
	}
	// XXX 202
	myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusOK, nil)
}
