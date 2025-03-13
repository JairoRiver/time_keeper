package serve

import (
	"context"
	"os"

	"github.com/JairoRiver/time_keeper/internal/api"
	"github.com/JairoRiver/time_keeper/internal/api/handler"
	"github.com/JairoRiver/time_keeper/internal/controller"
	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
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

			//ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
			//defer stop()
			ctx := context.Background()

			connPool, err := pgxpool.New(ctx, config.DBSOURCE)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot connect to db")
			}
			querier := db.New(connPool)
			control := controller.New(querier)
			handler := handler.New(control)
			server := api.New(handler, &logger)
			err = server.Start(config.ServerAddress)
			if err != nil {
				log.Fatal().Err(err).Msg("cannot start server:")
			}
			log.Info().Msgf("start HTTP gateway server at %s", config.ServerAddress)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", util.DefauldConfigPath, "config file")
	return cmd
}

func RegisterCommands(parent *cobra.Command) {
	cmd := NewServerCommand()
	parent.AddCommand(cmd)
}
