export const environment = {
  production: true,
  oauth: {
    oAuthRedirectUriBase: 'https://hisc.herokuapp.com/login',
    redirectUri: 'https://hisc.herokuapp.com/authorized',
    facebook: {
      clientId: '180253089366075',
      authUri: 'https://www.facebook.com/dialog/oauth',
      scope: 'public_profile,user_friends'
    }
  },
  socket: {
    maxSubscriptionRetries: 100
  },
  chat: {
    closeMessagesRangeSec: 60
  }
};
