package conf

import "net/http"

/* ApiError describes serializable error ready to be returned as http response*/
type ApiError struct {
  Code     int    `json:"code"`
  HttpCode int    `json:"-"`
  Message  string `json:"message"`
}

func (e *ApiError) Error() string {
  return e.Message
}

func NewApiError(err error) *ApiError {
  return &ApiError{0, http.StatusInternalServerError, err.Error()}
}

var (
  ErrAccountNotLoggedIn   = &ApiError{1, http.StatusUnauthorized, "Not logged in"}
  ErrInvalidToken         = &ApiError{2, http.StatusUnauthorized, "Access Toke is invalid"}
  ErrNoProfile            = &ApiError{3, http.StatusBadRequest, "Failed to get user profile"}
  ErrFriendsNoPermissions = &ApiError{4, http.StatusForbidden, "No permissions to fetch friends"}
  ErrNotAnOwner           = &ApiError{5, http.StatusForbidden, "Sorry, only owner can modify a chat room"}
  ErrNotAMember           = &ApiError{6, http.StatusForbidden, "Sorry, you are not a member of this chat room"}
  ErrInvalidId            = &ApiError{7, http.StatusBadRequest, "Invalid id specified"}
  ErrNotFound             = &ApiError{8, http.StatusNotFound, "Nothing to update"}
  ErrAlreadyExists        = &ApiError{9, http.StatusBadRequest, "Already exists"}
  ErrTooManyMembers       = &ApiError{10, http.StatusBadRequest, "Too many members"}
  ErrTooManyChatsOpened   = &ApiError{11, http.StatusBadRequest, "Too many rooms opened"}
)
