package commands

import (
	"flag"

	"github.com/jcomo/banana"
)

var cleanCommand *Command

func cleanArgs(fs *flag.FlagSet) {
}

func cleanRun() error {
	e, err := banana.NewEngine()
	if err != nil {
		return err
	}

	return e.Clean()
}

func init() {
	cleanCommand = &Command{
		Name:        "clean",
		Description: "Cleans the build directory",
		Args:        cleanArgs,
		Run:         cleanRun,
	}
}
