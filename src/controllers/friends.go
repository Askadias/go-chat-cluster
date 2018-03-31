package controllers

import (
  "net/http"
  "github.com/dgrijalva/jwt-go"
  "services"
  "github.com/codegangsta/martini-contrib/render"
  "conf"
  "github.com/go-martini/martini"
  "models"
)

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

func GetUser(params martini.Params, render render.Render, account services.Account) {
  userID := params["id"]
  if user, err := account.GetUser(userID); err != nil {
    render.JSON(err.HttpCode, err)
  } else {
    render.JSON(http.StatusOK, user)
  }
}

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
