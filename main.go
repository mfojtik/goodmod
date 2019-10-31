package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mfojtik/goodmod/pkg/cmd/replace"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := NewMainCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func NewMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goodmod",
		Short: "A pocket knife tool for manipulating go.mod files",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(255)
		},
	}

	cmd.AddCommand(replace.NewReplaceCommand())

	return cmd
}
