package banana

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fatih/color"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func withAccessLog(next http.Handler) http.Handler {
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
