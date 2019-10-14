package update

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type options struct {}

func (opts *options) AddFlags(flags *pflag.FlagSet) {

}

func (opts *options) Run() error {
	return nil
}

func (opts *options) Complete() error {
	return nil
}

func (opts *options) Validate() error {
	return nil
}

func NewUpdateCommand() *cobra.Command {
	updateOptions := &options{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			if err := updateOptions.Validate(); err != nil {
				log.Fatal(err)
			}
			if err := updateOptions.Complete(); err != nil {
				log.Fatal(err)
			}
			if err := updateOptions.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	updateOptions.AddFlags(cmd.Flags())

	return cmd
}