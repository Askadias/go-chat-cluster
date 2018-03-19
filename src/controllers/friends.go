package controllers

import (
  "net/http"
  "github.com/dgrijalva/jwt-go"
  "services"
  "github.com/codegangsta/martini-contrib/render"
  "conf"
)

func GetFriends(req *http.Request, render render.Render) {
  tkn := req.Context().Value("user").(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"]
    friends, err := services.Facebook.GetFriends(profileID.(string))
    if err != nil {
      render.JSON(err.HttpCode, err)
      return
    }
    render.JSON(http.StatusOK, friends)
  } else {
    render.JSON(http.StatusUnauthorized, conf.ErrInvalidToken)
  }
}
