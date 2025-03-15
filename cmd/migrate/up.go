package migrate

import (
	"database/sql"

	"github.com/JairoRiver/time_keeper/internal/repository/db/migrations"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewMigrateUpCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database up",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			log.Print("migrate up")

			config, err := util.LoadConfig(configFile)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot load config")
			}

			d, err := iofs.New(migrations.MigrationsFS, ".")
			if err != nil {
				log.Fatal().Err(err).Msg("cannot load migration files")
			}

			db, err := sql.Open("pgx", config.Database.DbSource)
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to connect to database")
			}
			defer db.Close()

			driver, err := postgres.WithInstance(db, &postgres.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to create mimgrate postgres intance")
			}

			migration, err := migrate.NewWithInstance("iofs", d, config.Database.DbName, driver)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot create new migrate instance")
			}

			if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatal().Err(err).Msg("failed to run migrate up")
			}

			log.Info().Msg("db migrated successfully")

		},
	}

	cmd.Flags().StringVar(&configFile, "config", util.DefauldConfigPath, "config file")

	return cmd
}
