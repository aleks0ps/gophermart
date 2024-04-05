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

func (s *Service) accrualDiscount(ctx context.Context, inputCh <-chan *storage.Order, resultCh chan<- *storage.Order) {
	for order := range inputCh {
		client := &http.Client{}
		// Calculate discount
		req, err := http.NewRequestWithContext(ctx, "GET", s.AccrualHTTP()+"/api/orders/"+order.Order, nil)
		if err != nil {
			s.Logger.Errorln(err.Error())
			resultCh <- order
			return
		}
		res, err := client.Do(req)
		if err != nil {
			s.Logger.Errorln(err.Error())
			resultCh <- order
			return
		}
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			s.Logger.Errorln(err.Error())
			resultCh <- order
			return
		}
		res.Body.Close()
		if len(buf) > 0 {
			// change original object
			if err := json.Unmarshal(buf, &order); err != nil {
				s.Logger.Errorln(err.Error())
				resultCh <- order
				return
			}
		}
		// return
		resultCh <- order
	}
}

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
	// Validate order number
	err = goluhn.Validate(buf.String())
	if err != nil {
		// 422
		myhttp.WriteError(&w, http.StatusUnprocessableEntity, err)
		return
	}
	user := storage.User{Login: login}
	order := storage.Order{Order: buf.String(), UploadedAt: time.Now().Format(time.RFC3339)}
	// Store order to database
	if err := s.DB.LoadOrder(r.Context(), &user, &order); err != nil {
		if errors.Is(err, myerror.ErrOrderLoaded) {
			// 200
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusOK, nil)
		} else if errors.Is(err, myerror.ErrOrderInUse) {
			// 409
			myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusConflict, nil)
		} else {
			// 500
			myhttp.WriteError(&w, http.StatusInternalServerError, err)
		}
		return
	}
	inputCh := make(chan *storage.Order, 1)
	resultCh := make(chan *storage.Order, 1)
	go s.accrualDiscount(r.Context(), inputCh, resultCh)
	inputCh <- &order
	close(inputCh)
	<-resultCh
	// Update balance
	err = s.DB.BalanceIncrease(r.Context(), &user, &order)
	if err != nil {
		s.Logger.Errorln(err.Error())
		myhttp.WriteError(&w, http.StatusInternalServerError, err)
		return
	}
	// 202
	myhttp.WriteResponse(&w, myhttp.CTypeNone, http.StatusAccepted, nil)
}
