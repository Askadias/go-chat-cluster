export const environment = {
  production: true,
  oauth: {
    oAuthRedirectUriBase: 'https://go-chat-cluster.heroku.com/login',
    redirectUri: 'https://go-chat-cluster.heroku.com/authorized',
    facebook: {
      clientId: '1132078350149238',
      authUri: 'https://www.facebook.com/dialog/oauth',
      scope: 'public_profile,user_friends'
    }
  }
};
