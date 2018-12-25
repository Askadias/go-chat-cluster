package services

import (
  "github.com/Askadias/go-chat-cluster/conf"
  "github.com/Askadias/go-chat-cluster/models"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/onsi/gomega/ghttp"
  "net/http"
  "time"
)

var _ = Describe("Facebook API Client", func() {
  var (
    facebook *Facebook
    server   *ghttp.Server
  )

  BeforeEach(func() {
    server = ghttp.NewServer()
    facebook = &Facebook{
      options: conf.FacebookConf{
        ClientID:     "123",
        ClientSecret: "234",
        BaseURL:      server.URL(),
        Timeout:      1 * time.Second,
        RedirectURL:  "https://localhost/authorize",
      },
      facebookClient: &http.Client{},
    }
  })

  AfterEach(func() {
    server.Close()
  })

  Describe("retrieving facebook client token", func() {
    var statusCode int
    var tokenResponse interface{}
    BeforeEach(func() {
      statusCode = http.StatusOK
      tokenResponse = models.Token{
        AccessToken: "777",
        TokenType:   "bearer",
        ExpiresIn:   111,
      }
      server.AppendHandlers(ghttp.CombineHandlers(
        ghttp.VerifyRequest("GET",
          "/oauth/access_token",
          "grant_type=client_credentials&client_id=123&client_secret=234"),
        ghttp.RespondWithJSONEncodedPtr(&statusCode, &tokenResponse),
      ))
    })

    Context("when response is successful", func() {
      It("should get client token using personal credentials", func() {
        facebook.setupClientToken()
        Ω(server.ReceivedRequests()).Should(HaveLen(1))
        Ω(facebook.clientAccessToken).Should(Equal(tokenResponse.(models.Token).AccessToken))
      })
    })

    Context("when the response fails to authenticate", func() {
      BeforeEach(func() {
        statusCode = http.StatusUnauthorized
        tokenResponse = FBError{Error: conf.ApiError{Message: "invalid credentials"}}
      })
      It("should get client token using personal credentials", func() {
        err := facebook.setupClientToken()
        Ω(server.ReceivedRequests()).Should(HaveLen(1))
        Ω(err.Error()).Should(Equal("invalid credentials"))
      })
    })
  })
})
