package controllers

import (
  "net/http"
  "services"
  "conf"
  "log"
  "github.com/dgrijalva/jwt-go"
  "time"
  "github.com/codegangsta/martini-contrib/render"
  "models"
)

type UserClaims struct {
  AvatarURL string `json:"avatar"`
  jwt.StandardClaims
}

func LoginWithProvider(extAuth models.ExtAuthCredentials, render render.Render) {
  accessToken, err := services.Facebook.GetAccessToken(extAuth.Code)
  if err != nil {
    render.JSON(err.HttpCode, err)
    return
  }
  profile, err := services.Facebook.GetProfile(accessToken)
  if err != nil {
    render.JSON(err.HttpCode, err)
    return
  }
  log.Println("User logged in:", profile.Name, "id:", profile.Id)
  jwtSignKey := []byte(conf.JWTSecret)

  // Create the Claims
  claims := UserClaims{
    AvatarURL: profile.AvatarURL,
    StandardClaims: jwt.StandardClaims{
      ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
      Id:        profile.Id,
      Issuer:    profile.Name,
    },
  }

  tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  signedJWT, e := tkn.SignedString(jwtSignKey)
  if e != nil {
    render.JSON(http.StatusInternalServerError, conf.NewApiError(e))
    return
  }
  log.Println("JWT issued:", signedJWT, "error:", err)
  render.JSON(http.StatusOK, map[string]string{"token": signedJWT})
}
