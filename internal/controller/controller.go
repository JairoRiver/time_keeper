package controller

import (
	"errors"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
)

// Control defines a Entry Time service controller.
type Control struct {
	repo   db.Querier
	config *util.Config
}

// New creates a short link service controller.
func New(repo db.Querier, config *util.Config) *Control {
	return &Control{repo, config}
}

var ErrInvalidRoleValue = errors.New("error invalid role value")
var ErrInvalidIdType = errors.New("error id must be an UUID")
var ErrInvalidEmailType = errors.New("error email must be a string")
var ErrInvalidGetParamType = errors.New("error get param type are invalid")
var ErrEmptyId = errors.New("error id are empty")
var ErrEmptyEmail = errors.New("error email are empty")
