package controllers

import (
  "net/http"
  "github.com/dgrijalva/jwt-go"
  "services"
  "github.com/codegangsta/martini-contrib/render"
  "conf"
)

func GetFriends(req *http.Request, render render.Render, facebook *services.Facebook) {
  tkn := req.Context().Value("user").(*jwt.Token)
  if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
    profileID := claims["jti"]

    if friends, err := facebook.GetFriends(profileID.(string)); err != nil {
      render.JSON(err.HttpCode, err)
    } else {
      render.JSON(http.StatusOK, friends)
    }
  } else {
    render.JSON(http.StatusUnauthorized, conf.ErrInvalidToken)
  }
}
