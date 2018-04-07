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

  private isPopup = true;
  errors: string[] = [];
  submitting = false;

  constructor(public route: ActivatedRoute,
              protected auth: AuthService,
              protected router: Router) {
  }

  ngOnInit() {
    const routeSnapshot = this.route.snapshot;
    const oauthProvider = routeSnapshot.paramMap.get('provider');
    this.isPopup = routeSnapshot.queryParams.isPopup || true;
    const state = routeSnapshot.queryParams.extLoginState;
    let authState = null;
    if (state) {
      try {
        authState = JSON.parse((<any>window).decodeURIComponent(atob(state)));
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
    this.auth.loginWith(provider, this.isPopup, false, false);
  }
}
