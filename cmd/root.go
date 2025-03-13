package cmd

import (
	"log"

	"github.com/JairoRiver/time_keeper/cmd/migrate"
	"github.com/JairoRiver/time_keeper/cmd/serve"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "timeKeeper",
	}

	serve.RegisterCommands(cmd)
	migrate.RegisterCommands(cmd)

	return cmd
}

func Execute() {
	cmd := NewRootCmd()

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
