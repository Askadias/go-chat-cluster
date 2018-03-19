import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute, Router} from "@angular/router";
import {AuthService} from "../services/auth.service";

@Component({
  selector: 'chat-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {

  private oauthConfig: any;
  private isPopup = true;
  errors: string[] = [];
  submitting = false;

  constructor(public route: ActivatedRoute,
              protected auth: AuthService,
              protected router: Router) {
    this.oauthConfig = env.oauth;
  }

  ngOnInit() {
    const routeSnapshot = this.route.snapshot;
    const oauthProvider = routeSnapshot.paramMap.get('provider');
    this.isPopup = routeSnapshot.queryParams.isPopup || true;
    const state = routeSnapshot.queryParams.extLoginState;
    let authState = null;
    if (state) {
      try {
        authState = JSON.parse(atob(state));
      } catch (error) {
      }
    }
    if (authState) {
      this.errors = authState.error ? [authState.error] : [];
      if (authState.accessCode) {
        this.submitting = true;
        this.auth.loginWithExtAuth(oauthProvider, authState.accessCode).subscribe(
          (resp) => {
            this.submitting = false;
            this.auth.setToken(resp.token);
            this.router.navigate([`/chat`]);
          }, (error) => {
            this.submitting = false;
            this.errors = [error.message];
          }
        );
      }
    } else {
    }
  }

  loginWith(provider: string) {
    const extAuthURL = this.buildClientAuthUri(this.oauthConfig[provider], provider);
    if (this.isPopup) {
      window.open(
        extAuthURL,
        `Login with ${provider}`,
        'width=400,height=600');
    } else {
      window.location.replace(extAuthURL)
    }
  }

  buildClientAuthUri(conf, provider) {
    const {authUri, clientId, scope} = conf;
    const redirectUri = this.oauthConfig.redirectUri;

    const state = btoa(JSON.stringify({
      randomString: Math.random().toString(36).slice(2),
      oauthRedirectUrl: `${this.oauthConfig.oAuthRedirectUriBase}/${provider}`,
      isPopup: this.isPopup
    }));
    sessionStorage.setItem('oauth_state', state);
    return `${authUri}?client_id=${clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scope}&response_type=code&display=popup`;
  }

}
