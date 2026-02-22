package user

import "net/http"

type UserService struct{}

func (h *UserService) Login(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// TODO: implementation of basic auth login
}
