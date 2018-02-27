package commands

import (
	"flag"

	"github.com/jcomo/banana"
)

var buildCommand *Command

var (
	buildVerbose bool
	buildClean   bool
)

func buildArgs(fs *flag.FlagSet) {
	fs.BoolVar(&buildVerbose, "verbose", false, "print debug output")
	fs.BoolVar(&buildClean, "clean", false, "cleans before building")
}

func buildRun() error {
	e, err := banana.NewEngine()
	if err != nil {
		return err
	}

	if buildClean {
		err = e.Clean()
		if err != nil {
			return err
		}
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
