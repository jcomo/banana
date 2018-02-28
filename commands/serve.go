package commands

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/jcomo/banana"
	"github.com/jcomo/banana/browser"
)

var serveCommand *Command

var (
	serveHost  string
	servePort  int
	serveClean bool
	serveWatch bool
)

func serveArgs(fs *flag.FlagSet) {
	fs.StringVar(&serveHost, "host", "localhost", "The network interface to listen on")
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

	addr := fmt.Sprintf("%s:%d", serveHost, servePort)
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

	go func() {
		time.Sleep(300 * time.Millisecond)
		browser.Open("http://" + addr)
	}()

	return e.Serve(addr)
}

func init() {
	serveCommand = &Command{
		Name:        "serve",
		Description: "Builds and serves the site",
		Args:        serveArgs,
		Run:         serveRun,
	}
}
