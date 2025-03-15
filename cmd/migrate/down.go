package migrate

import (
	"database/sql"
	"strconv"

	"github.com/JairoRiver/time_keeper/internal/repository/db/migrations"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var steps int

func NewMigrateDownCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "down",
		Short: "migrate the database down - given the number of steps",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			var err error
			if steps, err = strconv.Atoi(args[0]); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Print("migrate down")

			config, err := util.LoadConfig(configFile)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot load config")
			}

			d, err := iofs.New(migrations.MigrationsFS, ".")
			if err != nil {
				log.Fatal().Err(err).Msg("cannot load migration files")
			}

			db, err := sql.Open("pgx", config.DBSOURCE)
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to connect to database")
			}
			defer db.Close()

			driver, err := postgres.WithInstance(db, &postgres.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to create mimgrate postgres intance")
			}

			migration, err := migrate.NewWithInstance("iofs", d, config.DbName, driver)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot create new migrate instance")
			}

			if err = migration.Steps(-1 * steps); err != nil && err != migrate.ErrNoChange {
				log.Fatal().Err(err).Msg("failed to run migrate down")
			}

			log.Info().Msg("db migrated down successfully")
		},
	}

	cmd.Flags().StringVar(&configFile, "config", util.DefauldConfigPath, "config file")

	return cmd
}
