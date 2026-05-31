package serve

import (
	"context"
	"os"

	"github.com/JairoRiver/time_keeper/internal/api"
	"github.com/JairoRiver/time_keeper/internal/api/handler"
	"github.com/JairoRiver/time_keeper/internal/controller"
	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
	logtoClient "github.com/JairoRiver/time_keeper/pkg/identity/logto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewServerCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start time keeper server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			logger := zerolog.New(os.Stderr)
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

			config, err := util.LoadConfig(configFile)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot load config")
			}
			logger.Info().Msg("config loaded")

			ctx := context.Background()

			connPool, err := pgxpool.New(ctx, config.Database.DbSource)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot connect to db")
			}
			logger.Info().Msg("database pool created")

			idp, err := logtoClient.New(ctx, config)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot connect to identity provider")
			}
			logger.Info().Msg("identity provider ready")

			querier := db.New(connPool)
			control := controller.New(querier)
			h := handler.New(control, idp)
			server := api.New(h, &logger)

			logger.Info().Msgf("starting server at %s", config.Server.Address)
			if err := server.Start(config.Server.Address); err != nil {
				log.Fatal().Err(err).Msg("server stopped")
			}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", util.DefauldConfigPath, "config file")
	return cmd
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewServerCommand()
	parent.AddCommand(cmd)
}
