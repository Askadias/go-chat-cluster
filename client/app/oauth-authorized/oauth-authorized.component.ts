import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {isValidURL} from '../common/utils';
import {AuthService} from "../services/auth.service";

@Component({
  selector: 'chat-oauth-callback',
  templateUrl: './oauth-authorized.component.html',
  styleUrls: ['./oauth-authorized.component.scss']
})
export class OauthAuthorizedComponent implements OnInit {

  errors: string[];
  verifying = true;
  isPopup = false;
  returnTo = '/';

  constructor(private auth: AuthService,
              private route: ActivatedRoute,
              private router: Router) {
  }

  ngOnInit() {
    const OAUTH_INVALID_STATE_ERROR = 'Invalid State';
    const OAUTH_INVALID_REDIRECT_URL_ERROR = 'Invalid Redirect URL';
    this.verifying = true;
    this.errors = [];
    const params = this.route.snapshot.queryParams;
    const externalState = params.state;
    const localState = sessionStorage.getItem('oauth_state');
    let state: any;
    if (!localState || externalState !== localState) {
      this.errors.push(OAUTH_INVALID_STATE_ERROR);
    } else {
      try {
        state = JSON.parse(atob(externalState));
      } catch (err) {
        this.errors.push(OAUTH_INVALID_STATE_ERROR);
      }
    }

    if (params.error) {
      this.errors.push(params.error);
    } else {
      if (state) {
        state.accessCode = params.code;
        state.error = this.errors.length > 0 ? this.errors[0] : undefined;
        const redirectUrl = state.oauthRedirectUrl;
        this.returnTo = state.returnTo || '/';
        if (redirectUrl) {
          if (!isValidURL(redirectUrl)) {
            this.errors.push(OAUTH_INVALID_REDIRECT_URL_ERROR);
          } else {
            if (state.isPopup) {
              this.isPopup = true;
              const encodedState = btoa(JSON.stringify(state));
              window.opener.location = `${state.oauthRedirectUrl}?extLoginState=${encodedState}`;
              window.close();
            }
          }
        }
      }
    }
    this.verifying = false;
  }

  cancel() {
    if (this.isPopup) {
      window.close();
    } else {
      this.router.navigate([this.returnTo])
    }
  }
}
