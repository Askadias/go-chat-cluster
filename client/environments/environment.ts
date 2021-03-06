// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `.angular-cli.json`.

export const environment = {
  production: false,
  oauth: {
    oAuthRedirectUriBase: 'http://localhost:3000/login',
    redirectUri: 'http://localhost:3000/authorized',
    facebook: {
      clientId: '180253089366075',
      authUri: 'https://www.facebook.com/dialog/oauth',
      scope: 'public_profile,user_friends'
    }
  },
  socket: {
    maxSubscriptionRetries: 10
  },
  chat: {
    messagesLimit: 50,
    closeMessagesRangeSec: 60
  }
};
