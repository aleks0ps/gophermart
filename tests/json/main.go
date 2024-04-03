package main

import (
	"encoding/json"
	"fmt"
)

type Order struct {
	Order    string      `json:"order"`
	Withdraw json.Number `json:"sum"`
}

func main() {
	var js string = `{ "order": "7138742213177", "sum": 252.52 }`
	var emptySlice []Order
	var order Order
	err := json.Unmarshal([]byte(js), &order)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fl, err := order.Withdraw.Float64()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Order: " + order.Order)
	fmt.Printf("Withdraw: %f\n", fl)
	var order2 Order
	order2.Order = "7138742213177"
	order2.Withdraw = "252.52"
	res, err := json.Marshal(&order)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Res: " + string(res))
	//
	res, err = json.Marshal(emptySlice)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("RES of empty slice " + string(res))

}
