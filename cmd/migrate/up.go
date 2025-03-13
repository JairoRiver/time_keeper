package migrate

import (
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

			migration, err := migrate.New(util.MigrationFilesPath, config.DBSOURCE)
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
