package accounts

import "errors"

var ErrAlreadyExists = errors.New("an account already exists")

type Account struct {
	Id       int    `db:"id"`
	Username string `db:"username"`
}

type CreateParams struct {
	Username string
	Password string
}

type CreateValues struct {
	Username     string
	PasswordHash string
}
