package error

import "errors"

var LoginAlreadyTaken = errors.New("login already taken")
var InvalidLoginOrPassword = errors.New("invalid login or password")
var OrderLoaded = errors.New("order already loaded by user")
var OrderInUse = errors.New("order alredy loaded by another user")
var InsufficientBalance = errors.New("insufficient balance")
var NoWithdrawals = errors.New("no withdrawals")
var NoOrders = errors.New("no orders")
