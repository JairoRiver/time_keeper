package db

import (
	"context"
	"os"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var testQueries Queries

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../../../test.env")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config file")
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSOURCE)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	testQueries = *New(connPool)
	os.Exit(m.Run())
}
