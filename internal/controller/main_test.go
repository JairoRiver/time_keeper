package controller

import (
	"context"
	"os"
	"testing"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var testControl Control

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config file")
	}

	connPool, err := pgxpool.New(context.Background(), config.Database.DbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	repo := *db.New(connPool)
	testControl = *New(&repo)
	os.Exit(m.Run())
}
