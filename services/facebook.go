package services

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/Askadias/go-chat-cluster/models"
)

// Error in the format returned by the facebook.com
type FBError struct {
  Error conf.ApiError `json:"error"`
}

type FBPermissionsResponse struct {
  data []FBPermission
}

type FBPermission struct {
  permission string
  status     string
}

// Facebook Service performs necessary operations related to the Facebook.com social network such as:
//    - retrieving client token by service credentials
//    - exchanging authorization code to access token
//    - retrieving user profile
//    - retrieving user friends list
//    - retrieving user by id
type Facebook struct {
  options           conf.FacebookConf
  facebookClient    *http.Client
  clientAccessToken string
}

func NewFacebook(options conf.FacebookConf, httpClient *http.Client) (f *Facebook) {
  f = &Facebook{
    options: options,
    facebookClient: httpClient,
  }
  if err := f.setupClientToken(); err != nil {
    panic(err)
  }
  return f
}

// Retrieves client access token using service credentials.
func (f *Facebook) setupClientToken() error {
  resp, err := f.facebookClient.Get(f.options.BaseURL + "/oauth/access_token?grant_type=client_credentials" +
    "&client_id=" + f.options.ClientID +
    "&client_secret=" + f.options.ClientSecret)
  if err != nil {
    log.Println(err)
    return conf.NewApiError(err)
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)
  if resp.StatusCode >= 400 {
    return parseError(resp.StatusCode, body)
  }
  token := &models.Token{}

  if err := json.Unmarshal(body, token); err != nil {
    log.Println(err)
    return conf.NewApiError(err)
  }
  f.clientAccessToken = token.AccessToken
  return nil
}

// Exchanges user authorization code to its access token.
func (f *Facebook) ExchangeCodeToToken(accessCode string) (string, *conf.ApiError) {
  resp, err := f.facebookClient.Get(f.options.BaseURL + "/oauth/access_token" +
    "?client_id=" + f.options.ClientID +
    "&redirect_uri=" + f.options.RedirectURL +
    "&client_secret=" + f.options.ClientSecret +
    "&code=" + accessCode)
  if err != nil {
    log.Println(err)
    return "", conf.ErrAccountNotLoggedIn
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)
  if resp.StatusCode >= 400 {
    return "", parseError(resp.StatusCode, body)
  }
  token := &models.Token{}
  if err := json.Unmarshal(body, token); err != nil {
    log.Println(err)
    return "", conf.NewApiError(err)
  }
  return token.AccessToken, nil
}

// Retrieves user profile using access token.
func (f *Facebook) GetProfile(accessToken string) (*models.User, *conf.ApiError) {
  req, _ := http.NewRequest("GET", f.options.BaseURL+"/me", nil)
  req.Header.Set("Authorization", "Bearer "+accessToken)
  resp, err := f.facebookClient.Do(req)
  if err != nil {
    log.Println(err)
    return nil, conf.ErrNoProfile
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return nil, parseError(resp.StatusCode, body)
  }
  user := &models.User{}
  if err := json.Unmarshal(body, user); err != nil {
    log.Println(err)
    return nil, conf.NewApiError(err)
  }
  user.AvatarURL = f.options.BaseURL + "/" + user.ID + "/picture"
  return user, nil
}

// Retrieves user by its ID.
func (f *Facebook) GetUser(profileID string) (*models.User, *conf.ApiError) {
  resp, err := f.facebookClient.Get(f.options.BaseURL + "/" + profileID + "?access_token=" + f.clientAccessToken)
  if err != nil {
    log.Println(err)
    return nil, conf.ErrNoProfile
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return nil, parseError(resp.StatusCode, body)
  }
  user := &models.User{}
  if err := json.Unmarshal(body, user); err != nil {
    log.Println(err)
    return nil, conf.NewApiError(err)
  }
  user.AvatarURL = f.options.BaseURL + "/" + user.ID + "/picture"
  return user, nil
}

// Retrieves user friends list.
func (f *Facebook) GetFriends(profileID string) ([]models.User, *conf.ApiError) {
  resp, err := f.facebookClient.Get(f.options.BaseURL + "/" + profileID + "/friends?access_token=" + f.clientAccessToken)
  if err != nil {
    log.Println(err)
    return nil, conf.ErrNoProfile
  }

  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return nil, parseError(resp.StatusCode, body)
  }
  var friends = &models.UserList{}
  if err := json.Unmarshal(body, &friends); err != nil {
    log.Println(err)
    return nil, conf.NewApiError(err)
  }
  for i := range friends.Data {
    friends.Data[i].AvatarURL = f.options.BaseURL + "/" + friends.Data[i].ID + "/picture"
  }
  return friends.Data, nil
}

func (f *Facebook) HasFriendsPermissions(profileID string) *conf.ApiError {
  resp, err := f.facebookClient.Get(f.options.BaseURL + "/" + profileID + "/permissions?access_token=" + f.clientAccessToken)
  if err != nil {
    log.Println(err)
    return conf.ErrNoProfile
  }
  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  if resp.StatusCode >= 400 {
    return parseError(resp.StatusCode, body)
  }
  var permissions = &FBPermissionsResponse{}
  if err := json.Unmarshal(body, &permissions); err != nil {
    log.Println(err)
    return conf.NewApiError(err)
  }
  for _, p := range permissions.data {
    if p.permission == "user_friends" && p.status == "granted" {
      return nil
    }
  }
  return conf.ErrFriendsNoPermissions
}

// Parse facebook specific error.
func parseError(statusCode int, body []byte) *conf.ApiError {
  fbError := &FBError{}
  log.Println("Facebook returned error response", string(body))
  if err := json.Unmarshal(body, fbError); err != nil {
    log.Println(err)
    return conf.NewApiError(err)
  }
  fbError.Error.HttpCode = statusCode
  return &fbError.Error
}
