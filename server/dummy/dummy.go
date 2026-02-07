package dummy

import (
	"io"
	"net/http"
)

// Echo is an example handler that simply echoes back the request body.
func Echo(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	_, _ = w.Write(append(b, []byte("\n")...))
}

// Panic is an example handler that will panic when called.
func Panic(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	panic("this is a panic: " + string(b))
}

// Hello is an example handler that only answers to GET requests with "hello world".
func Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world\n"))
}

func City(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "server/static/city.html")
}
