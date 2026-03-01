package main

import (
	"context"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

var (
	// noAuthEndpoints is an allowlist for endpoints that do not require authentication
	noAuthEndpoints = map[string]struct{}{
		"POST /api/login":  {},
		"POST /api/signup": {},
	}
)

// authMiddleware validates the JWT in the Authorization header and adds the user ID to the context
//
// The middleware should be chained for all endpoints per default, and the noAuthEndpoint variable
// should be used to specify any endpoints that should skip authentication (e.g. login, signup). This ensures a
// default secure behavior while allowing flexibility for public endpoints.
//
// Endpoints are still responsible for implementing the authorization part, i.e., checking if a certain
// user is allowed to perform a certain action, by using the user ID in the context.
func authMiddleware(secretKey string) func(http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := noAuthEndpoints[r.Method+" "+r.URL.Path]; ok {
				f(w, r)
				return
			}

			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "missing authorization token", http.StatusUnauthorized)
				return
			}
			tokenString, ok := strings.CutPrefix(tokenString, "Bearer ")
			if !ok {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			sub, err := token.Claims.GetSubject()
			if err != nil {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			// add user ID from token claims to request context for future handlers to use in authorization
			r = r.WithContext(context.WithValue(r.Context(), "sub", sub))
			f(w, r)
		})
	}
}
