export const environment = {
  production: true,
  oauth: {
    oAuthRedirectUriBase: 'https://hisc.herokuapp.com/login',
    redirectUri: 'https://hisc.herokuapp.com/authorized',
    facebook: {
      clientId: '1132078350149238',
      authUri: 'https://www.facebook.com/dialog/oauth',
      scope: 'public_profile,user_friends'
    }
  }
};
