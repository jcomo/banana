package commands

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jcomo/banana"
)

var serveCommand *Command

var (
	servePort  int
	serveClean bool
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)
		fmt.Printf("%d %s %s\n", lrw.statusCode, r.Method, r.RequestURI)
	})
}

func serveArgs(fs *flag.FlagSet) {
	fs.IntVar(&servePort, "port", 4000, "The port to serve on")
	fs.BoolVar(&serveClean, "clean", false, "Clean before building")
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
	handler := http.FileServer(http.Dir("_build"))
	fmt.Printf("Serving your site on %s!\n", addr)
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
