package commands

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/fatih/color"
	"github.com/jcomo/banana"
)

var serveCommand *Command

var (
	servePort  int
	serveClean bool
	serveWatch bool
)

func serveArgs(fs *flag.FlagSet) {
	fs.IntVar(&servePort, "port", 4000, "The port to serve on")
	fs.BoolVar(&serveClean, "clean", false, "Clean before building")
	fs.BoolVar(&serveWatch, "watch", false, "Watch for file changes")
}

func serveRun() error {
	e, err := banana.NewEngine()
	if err != nil {
		return err
	}

	err = e.Build()
	if err != nil {
		return err
	}

	magenta := color.New(color.FgMagenta).SprintFunc()
	log.Printf("Serving your site on %s!\n", magenta(addr))

	if serveWatch {
		closer, err := e.Watch()
		if err != nil {
			return err
		}

		log.Println("Watching for file changes...")
		defer closer.Close()
	}

	return e.Serve()
}

func init() {
	serveCommand = &Command{
		Name:        "serve",
		Description: "Builds and serves the site",
		Args:        serveArgs,
		Run:         serveRun,
	}
}
