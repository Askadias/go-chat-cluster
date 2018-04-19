package auth

import (
  "context"
  "errors"
  "fmt"
  "github.com/dgrijalva/jwt-go"
  "log"
  "net/http"
  "strings"
  "github.com/go-martini/martini"
)

// A function called whenever an error is encountered
type errorHandler func(w http.ResponseWriter, err string)

// TokenExtractor is a function that takes a request as input and returns
// either a token or an error.  An error should only be returned if an attempt
// to specify a token was found, but the information was somehow incorrectly
// formed.  In the case where a token is simply not present, this should not
// be treated as an error.  An empty string should be returned in that case.
type TokenExtractor func(r *http.Request) (string, error)

// Options is a struct for specifying configuration options for the middleware.
type Options struct {
  // The function that will return the Key to validate the JWT.
  // It can be either a shared secret or a public key.
  // Default value: nil
  ValidationKeyGetter jwt.Keyfunc
  // The name of the property in the request where the user information
  // from the JWT will be stored.
  // Default value: "user"
  UserProperty string
  // The function that will be called when there's an error validating the token
  // Default value:
  ErrorHandler errorHandler
  // A boolean indicating if the credentials are required or not
  // Default value: false
  CredentialsOptional bool
  // A function that extracts the token from the request
  // Default: FromAuthHeader (i.e., from Authorization header as bearer token)
  Extractor TokenExtractor
  // Debug flag turns on debugging output
  // Default: false
  Debug bool
  // When set, all requests with the OPTIONS method will use authentication
  // Default: false
  EnableAuthOnOptions bool
  // When set, the middleware verifies that tokens are signed with the specific signing algorithm
  // If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
  // Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
  // Default: nil
  SigningMethod jwt.SigningMethod
}

type JWTMiddleware struct {
  Options Options
}

func OnError(w http.ResponseWriter, err string) {
  http.Error(w, err, http.StatusUnauthorized)
}

// New constructs a new Secure instance with supplied options.
func NewJwtMiddleware(options ...Options) *JWTMiddleware {

  var opts Options
  if len(options) == 0 {
    opts = Options{}
  } else {
    opts = options[0]
  }

  if opts.UserProperty == "" {
    opts.UserProperty = "user"
  }

  if opts.ErrorHandler == nil {
    opts.ErrorHandler = OnError
  }

  if opts.Extractor == nil {
    opts.Extractor = FromAuthHeader
  }

  return &JWTMiddleware{
    Options: opts,
  }
}

func (m *JWTMiddleware) logf(format string, args ...interface{}) {
  if m.Options.Debug {
    log.Printf(format, args...)
  }
}

// FromAuthHeader is a "TokenExtractor" that takes a give request and extracts
// the JWT token from the Authorization header.
func FromAuthHeader(r *http.Request) (string, error) {
  authHeader := r.Header.Get("Authorization")
  if authHeader == "" {
    return "", nil // No error, just no token
  }
  authHeaderParts := strings.Split(authHeader, " ")
  if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
    return "", errors.New("authorization header format must be Bearer {token}")
  }
  return authHeaderParts[1], nil
}

// FromJWTCookie is a "TokenExtractor" that takes a give request and extracts
// the JWT token from the JWT cookie.
func FromJWTCookie(r *http.Request) (string, error) {
  if authCookie, err := r.Cookie("JWT"); err != nil {
    return "", err
  } else {
    return authCookie.Value, nil
  }
}

func (m *JWTMiddleware) CheckJWT(w http.ResponseWriter, c martini.Context, r *http.Request) {
  if !m.Options.EnableAuthOnOptions {
    if r.Method == "OPTIONS" {
      c.Next()
      return
    }
  }

  // Use the specified token extractor to extract a token from the request
  token, err := m.Options.Extractor(r)

  // If debugging is turned on, log the outcome
  if err != nil {
    m.logf("Error extracting JWT: %v", err)
  } else {
    m.logf("Token extracted: %s", token)
  }

  // If an error occurs, call the error handler and return an error
  if err != nil {
    m.Options.ErrorHandler(w, err.Error())
    OnError(w, fmt.Sprintf("error extracting token: %v", err))
  }

  // If the token is empty...
  if token == "" {
    // Check if it was required
    if m.Options.CredentialsOptional {
      m.logf("no credentials found (CredentialsOptional=true)")
      // No error, just no token (and that is ok given that CredentialsOptional is true)
      c.Next()
      return
    }

    // If we get here, the required token is missing
    errorMsg := "required authorization token not found"
    m.Options.ErrorHandler(w, errorMsg)
    m.logf("no credentials found (CredentialsOptional=false)")
    return
  }

  // Now parse the token
  parsedToken, err := jwt.Parse(token, m.Options.ValidationKeyGetter)

  // Check if there was an error in parsing...
  if err != nil {
    m.logf("Error parsing token: %v", err)
    m.Options.ErrorHandler(w, err.Error())
    return
  }

  if m.Options.SigningMethod != nil && m.Options.SigningMethod.Alg() != parsedToken.Header["alg"] {
    message := fmt.Sprintf("expected %s signing method but token specified %s",
      m.Options.SigningMethod.Alg(),
      parsedToken.Header["alg"])
    m.logf("error validating token algorithm: %s", message)
    m.Options.ErrorHandler(w, errors.New(message).Error())
    return
  }

  // Check if the parsed token is valid...
  if !parsedToken.Valid {
    m.logf("token is invalid")
    m.Options.ErrorHandler(w, "token isn't valid")
    return
  }

  m.logf("JWT: %v", parsedToken)

  // If we get here, everything worked and we can set the
  // user property in context.
  newRequest := r.WithContext(context.WithValue(r.Context(), m.Options.UserProperty, parsedToken))
  // Update the current request with the new context information.
  *r = *newRequest
  c.Next()
  return
}
