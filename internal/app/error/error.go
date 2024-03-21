package error

import "errors"

var LoginAlreadyTaken = errors.New("login already taken")
var InvalidLoginOrPassword = errors.New("invalid login or password")
