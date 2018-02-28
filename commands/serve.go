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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func logAccess(next http.Handler) http.Handler {
	white := color.New(color.FgWhite).Add(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)

		chalk := green
		if lrw.statusCode >= 500 {
			chalk = red
		} else if lrw.statusCode >= 400 {
			chalk = yellow
		} else if lrw.statusCode >= 300 {
			chalk = cyan
		}

		code := strconv.Itoa(lrw.statusCode)
		log.Printf("%s %s %s\n", chalk(code), white(r.Method), r.RequestURI)
	})
}

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

	addr := fmt.Sprintf(":%d", servePort)
	handler := http.FileServer(http.Dir(e.OutDir))

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

	return http.ListenAndServe(addr, logAccess(handler))
}

func init() {
	serveCommand = &Command{
		Name:        "serve",
		Description: "Builds and serves the site",
		Args:        serveArgs,
		Run:         serveRun,
	}
}
