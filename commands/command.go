package commands

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	Name        string
	Description string
	Args        func(*flag.FlagSet)
	Run         func() error
}

var (
	description string
	commands    map[string]*Command
)

func addCommand(cmd *Command) {
	commands[cmd.Name] = cmd
}

func loadCommands() {
	description = `A static site generator`
	commands = make(map[string]*Command)
	addCommand(buildCommand)
	addCommand(cleanCommand)
	addCommand(helpCommand)
	addCommand(serveCommand)
}

func printHelp() {
	fmt.Printf("usage: %s COMMAND\n", os.Args[0])
	fmt.Printf("\n%s\n", description)
	fmt.Printf("\navailable commands\n")
	for _, cmd := range commands {
		fmt.Printf("   %-12s %s\n", cmd.Name, cmd.Description)
	}
}

func createUsage(cmd *Command, fs *flag.FlagSet) func() {
	return func() {
		prog := os.Args[0]
		fmt.Printf("usage: %s %s [OPTIONS]\n", prog, cmd.Name)
		fmt.Printf("\n%s\n", cmd.Description)
		fmt.Printf("\navailable options\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Printf("   --%-20s %s\n", f.Name, f.Usage)
		})
	}
}

func Run() {
	loadCommands()
	if len(os.Args) <= 1 {
		printHelp()
		os.Exit(2)
	}

	name := os.Args[1]
	cmd, ok := commands[name]
	if !ok {
		printHelp()
		os.Exit(2)
	}

	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = createUsage(cmd, fs)
	cmd.Args(fs)

	err := fs.Parse(os.Args[2:])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
