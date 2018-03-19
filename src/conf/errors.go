package conf

import "net/http"

type ApiError struct {
  Code     int    `json:"code"`
  HttpCode int    `json:"-"`
  Message  string `json:"message"`
  Info     string `json:"info"`
}

func (e *ApiError) Error() string {
  return e.Message
}

func NewApiError(err error) *ApiError {
  return &ApiError{0, http.StatusInternalServerError, err.Error(), ""}
}

var ErrAccountNotLoggedIn = &ApiError{1, http.StatusUnauthorized, "Not logged in", "Please Sign In"}
var ErrInvalidToken = &ApiError{2, http.StatusUnauthorized, "Access Toke is invalid", ""}
var ErrNoProfile = &ApiError{3, http.StatusBadRequest, "Failed to get user profile", ""}
var ErrFriendsNoPermissions = &ApiError{4, http.StatusForbidden, "No permissions to fetch friends", "Chat doesn't have permissions to fetch friends"}
