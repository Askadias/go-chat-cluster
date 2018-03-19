package services

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  "conf"
  "models"
  "time"
)

var Facebook = NewFacebookService(conf.FBClientID, conf.FBClientSecret, conf.FBRedirectURL, conf.FBTimeoutMS)

type FacebookService struct {
  clientId          string
  clientSecret      string
  redirectURL       string
  facebookClient    *http.Client
  clientAccessToken string
}

type FBError struct {
  Error conf.ApiError `json:"error"`
}

func NewFacebookService(clientId string, clientSecret string, redirectURL string, timeoutMS time.Duration) (f *FacebookService) {
  f = &FacebookService{
    clientId:     clientId,
    clientSecret: clientSecret,
    redirectURL:  redirectURL,
    facebookClient: &http.Client{
      Timeout: time.Millisecond * timeoutMS,
    },
  }
  err := f.setupClientToken()
  if err != nil {
    panic(err)
  }
  return
}

func (f *FacebookService) setupClientToken() error {
  resp, err := f.facebookClient.Get(conf.FBBaseURL + "/oauth/access_token?grant_type=client_credentials" +
    "&client_id=" + f.clientId +
    "&client_secret=" + f.clientSecret)
  if err != nil {
    log.Fatal(err)
    return conf.NewApiError(err)
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)
  token := &models.Token{}

  err = json.Unmarshal(body, token)
  if err != nil {
    log.Fatal(err)
    return conf.NewApiError(err)
  }
  f.clientAccessToken = token.AccessToken
  return nil
}

func parseError(statusCode int, body []byte) *conf.ApiError {
  fbError := &FBError{}
  err := json.Unmarshal(body, fbError)
  if err != nil {
    log.Fatal(err)
    return conf.NewApiError(err)
  }
  fbError.Error.HttpCode = statusCode
  return &fbError.Error
}

func (f *FacebookService) GetAccessToken(accessCode string) (string, *conf.ApiError) {
  resp, err := f.facebookClient.Get(conf.FBBaseURL + "/oauth/access_token" +
    "?client_id=" + f.clientId +
    "&redirect_uri=" + f.redirectURL +
    "&client_secret=" + f.clientSecret +
    "&code=" + accessCode)
  if err != nil {
    log.Fatal(err)
    return "", conf.ErrAccountNotLoggedIn
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)
  if resp.StatusCode >= 400 {
    return "", parseError(resp.StatusCode, body)
  }
  token := &models.Token{}
  err = json.Unmarshal(body, token)
  if err != nil {
    log.Fatal(err)
    return "", conf.NewApiError(err)
  }
  return token.AccessToken, nil
}

func (f *FacebookService) GetProfile(accessToken string) (*models.User, *conf.ApiError) {
  req, _ := http.NewRequest("GET", conf.FBBaseURL+"/me", nil)
  req.Header.Set("Authorization", "Bearer "+accessToken)
  resp, err := f.facebookClient.Do(req)
  if err != nil {
    log.Fatal(err)
    return nil, conf.ErrNoProfile
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return nil, parseError(resp.StatusCode, body)
  }
  user := &models.User{}
  err = json.Unmarshal(body, user)
  if err != nil {
    log.Fatal(err)
    return nil, conf.NewApiError(err)
  }
  user.AvatarURL = conf.FBBaseURL + "/" + user.Id + "/picture"
  return user, nil
}

func (f *FacebookService) GetFriends(profileID string) ([]models.User, *conf.ApiError) {
  resp, err := f.facebookClient.Get(conf.FBBaseURL + "/" + profileID + "/friends?access_token=" + f.clientAccessToken)
  if err != nil {
    log.Fatal(err)
    return nil, conf.ErrNoProfile
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return nil, parseError(resp.StatusCode, body)
  }
  var friends = &models.UserList{}
  err = json.Unmarshal(body, &friends)
  if err != nil {
    log.Fatal(err)
    return nil, conf.NewApiError(err)
  }
  for i := range friends.Data {
    friends.Data[i].AvatarURL = conf.FBBaseURL + "/" + friends.Data[i].Id + "/picture"
  }
  return friends.Data, nil
}

func (f *FacebookService) GetAuthorizeURL() string {
  return conf.FBAuthorizeURL +
    "?client_id=" + f.clientId +
    "&redirect_uri=" + f.redirectURL +
    "&scope=" + conf.FBScope
}
