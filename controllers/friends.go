package controllers

import (
  "net/http"
  "github.com/dgrijalva/jwt-go"
  "github.com/Askadias/go-chat-cluster/services"
  "github.com/codegangsta/martini-contrib/render"
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/go-martini/martini"
  "github.com/Askadias/go-chat-cluster/models"
)

// Returns list of authenticated user friends
func GetFriends(req *http.Request, render render.Render, friends services.Friends) {
  tkn := req.Context().Value("user").(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"]

    if friends, err := friends.GetFriends(profileID.(string)); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, friends)
    }
  } else {
    render.JSON(http.StatusUnauthorized, conf.ErrInvalidToken)
  }
}

// Returns error response if application doesn't have permission to fetch user friends
func HasFriendsPermissions(req *http.Request, res http.ResponseWriter, render render.Render, friends services.Friends) {
  tkn := req.Context().Value("user").(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"]

    if err := friends.HasFriendsPermissions(profileID.(string)); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      res.WriteHeader(http.StatusOK)
    }
  } else {
    render.JSON(http.StatusUnauthorized, conf.ErrInvalidToken)
  }
}

// Returns user by ID
func GetUser(params martini.Params, render render.Render, account services.Account) {
  userID := params["id"]
  if user, err := account.GetUser(userID); err != nil {
    render.JSON(err.HttpCode, err)
  } else {
    render.JSON(http.StatusOK, user)
  }
}

// Batch request to fetch users by IDs
func GetUsers(req *http.Request, render render.Render, account services.Account) {
  qs := req.URL.Query()
  userIDs := qs["userId"]
  var users []models.User
  for _, userID := range userIDs {
    if user, err := account.GetUser(userID); err != nil {
      render.JSON(err.HttpCode, err)
      return
    } else {
      users = append(users, *user)
    }
  }
  render.JSON(http.StatusOK, users)
}
