package commands

import (
	"flag"

	"github.com/jcomo/banana"
)

var buildCommand *Command

var (
	buildVerbose bool
)

func buildArgs(f *flag.FlagSet) {
	f.BoolVar(&buildVerbose, "verbose", false, "print debug output")
}

func buildRun() error {
	e, err := banana.NewEngine()
	if err != nil {
		return err
	}

	return e.Build()
}

func init() {
	buildCommand = &Command{
		Name:        "build",
		Description: "Build the site",
		Args:        buildArgs,
		Run:         buildRun,
	}
}
