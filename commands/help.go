package commands

import (
	"flag"
)

var helpCommand *Command

func helpArgs(f *flag.FlagSet) {
}

func helpRun() error {
	printHelp()
	return nil
}

func init() {
	helpCommand = &Command{
		Name:        "help",
		Description: "Prints this help message and exits",
		Args:        helpArgs,
		Run:         helpRun,
	}
}
