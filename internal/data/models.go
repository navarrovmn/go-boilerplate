package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models creates a wrapper that will have lots of models
type Models struct {
	Users UserModel
}

// NewModels for ease of us which returns Model struct containing the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
