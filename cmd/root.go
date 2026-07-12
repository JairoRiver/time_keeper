package cmd

import (
	"log"

	"github.com/JairoRiver/time_keeper/cmd/migrate"
	"github.com/JairoRiver/time_keeper/cmd/serve"
	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "timeKeeper",
		Version: version,
	}

	serve.RegisterCommands(cmd, version)
	migrate.RegisterCommands(cmd)

	return cmd
}

func Execute(version string) {
	cmd := NewRootCmd(version)

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
