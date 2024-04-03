package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	mycookie "github.com/aleks0ps/gophermart/internal/app/cookie"
	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
	"github.com/aleks0ps/gophermart/internal/app/storage"
)

func (s *Service) GetOrders(w http.ResponseWriter, r *http.Request) {
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
	orders, err := s.DB.GetOrders(r.Context(), &user)
	if err != nil {
		myhttp.WriteError(&w, http.StatusInternalServerError, err)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, order := range orders {
			order.Status = "NEW"
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
			// for json output
			order.Number = strings.Clone(order.Order)
			if len(buf) > 0 {
				if err := json.Unmarshal(buf, order); err != nil {
					s.Logger.Errorln(err.Error())
					return
				}
			}
		}
	}()
	wg.Wait()
	for _, order := range orders {
		order.Order = ""
	}
	if len(orders) == 0 {
		myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusNoContent, nil)
		return
	}
	res, err := json.Marshal(orders)
	if err != nil {
		myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusInternalServerError, nil)
		return
	}
	myhttp.WriteResponse(&w, myhttp.CTypeJSON, http.StatusOK, res)
}
