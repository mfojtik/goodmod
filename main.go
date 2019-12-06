package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mfojtik/goodmod/pkg/cmd/bump"
	"github.com/mfojtik/goodmod/pkg/cmd/replace"
	"github.com/mfojtik/goodmod/pkg/cmd/report"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := NewMainCommand()
	if err := command.Execute(); err != nil {
		if _, printErr := fmt.Fprintf(os.Stderr, "%v\n", err); printErr != nil {
			panic(printErr)
		}
		os.Exit(1)
	}
}

func NewMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goodmod",
		Short: "A pocket knife tool for manipulating go.mod files",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic(err)
			}
			os.Exit(255)
		},
	}

	cmd.AddCommand(replace.NewReplaceCommand())
	cmd.AddCommand(report.NewReportCommand())
	cmd.AddCommand(bump.NewBumpCommand())

	return cmd
}
