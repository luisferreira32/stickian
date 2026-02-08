package main

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func chainMiddleware(f http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	if len(middlewares) == 0 {
		return f
	}
	return middlewares[0](chainMiddleware(f, middlewares[1:]...))
}

// panicMiddleware prevents a panic from crashing the server in a failed request.
func panicMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					log.Printf("panic: %+v\nstack:\n%v\n", r, string(debug.Stack()))
				}
			}()

			f(w, r)
		}
	}
}

// loggingMiddleware adds logging for each request with method and path
//
// The middleware should only be chained if running in a development environment.
func loggingMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s\n", r.Method, r.URL.Path)
			f(w, r)
		}
	}
}

var (
	// file extensions that we will compress during build
	supportedCompressedExts = map[string]bool{
		".js":  true,
		".css": true,
	}
)

// compressionMiddleware serves pre-compressed files when available
//
// This reduces server load and network bandwidth for clients that support
// brotli compression, which is especially beneficial for large files.
func compressionMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fullPath := filepath.Join("dist", r.URL.Path)
			fileExt := filepath.Ext(fullPath)
			acceptedEncodings := r.Header.Get("Accept-Encoding")

			if strings.Contains(acceptedEncodings, "br") && supportedCompressedExts[fileExt] {
				w.Header().Set("Content-Encoding", "br")
				w.Header().Set("Content-Type", mime.TypeByExtension(fileExt))
				http.ServeFile(w, r, fullPath+".br")
				return
			}

			// serve uncompressed file
			f(w, r)
		})
	}
}
