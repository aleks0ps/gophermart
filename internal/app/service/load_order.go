package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	mycookie "github.com/aleks0ps/gophermart/internal/app/cookie"
	myerror "github.com/aleks0ps/gophermart/internal/app/error"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"

	"github.com/ShiraazMoollatjie/goluhn"
)

func (s *Service) LoadOrder(w http.ResponseWriter, r *http.Request) {
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
	if myhttp.GetContentTypeCode(stype) != myhttp.CTypePlain {
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
	// XXX
	//s.Logger.Infoln(" LoadOrder: " + string(buf.Bytes()))
	// Validate order number
	err = goluhn.Validate(string(buf.Bytes()))
	if err != nil {
		// 422
		myhttp.WriteError(&w, http.StatusUnprocessableEntity, err)
		return
	}
	user := storage.User{Login: login}
	order := storage.Order{Order: string(buf.Bytes()), UploadedAt: time.Now().Format(time.RFC3339)}
	// Store order to database
	if err := s.DB.LoadOrder(r.Context(), &user, &order); err != nil {
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
	go func() {
		res, err := http.Get(s.AccrualHTTP() + "/api/orders/" + order.Order)
		if err != nil {
			s.Logger.Errorln(err.Error())
			return
		}
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			s.Logger.Errorln(err.Error())
			return
		}
		if len(buf) > 0 {
			// XXX
			s.Logger.Infoln(string(buf))
			if err := json.Unmarshal(buf, &order); err != nil {
				s.Logger.Errorln(err.Error())
				return
			}
		}
		// Update balance
		ctx := context.Background()
		err = s.DB.BalanceIncrease(ctx, &user, &order)
		if err != nil {
			s.Logger.Errorln(err.Error())
		}
	}()
	// 202
	myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusAccepted, nil)
}
