package handler

import (
	"errors"
)

// errEmailAlreadyTaken represents creating user with duplicate email error.
var errEmailAlreadyTaken = errors.New("email already taken")
