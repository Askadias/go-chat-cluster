import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';
import * as jwt_decode from 'jwt-decode';
import {Router} from "@angular/router";
import {User} from "../domain/user";
import {CookieService} from "ngx-cookie-service";
import {environment as env} from "../../environments/environment";

export const ACCESS_TOKEN: string = 'JWT';
export const USER_ID: string = 'USER_ID';
export const USER_NAME: string = 'USER_NAME';
export const AVATAR_URL: string = 'AVATAR_URL';

@Injectable()
export class AuthService {

  constructor(private http: HttpClient, private router: Router, private cookies: CookieService) {
  }

  logout() {
    localStorage.removeItem(ACCESS_TOKEN);
    localStorage.removeItem(USER_NAME);
    localStorage.removeItem(AVATAR_URL);
    localStorage.removeItem(USER_ID);
    this.cookies.delete(ACCESS_TOKEN);
    this.router.navigate(['/login']);
  }

  loginWithExtAuth(provider, code: string): Observable<any> {
    return this.http.post<any>(`/api/login/${provider}`, {code});
  }

  getToken(): string {
    return localStorage.getItem(ACCESS_TOKEN);
  }

  getProfile(): User {
    return {
      id: localStorage.getItem(USER_ID),
      name: localStorage.getItem(USER_NAME),
      avatarUrl: localStorage.getItem(AVATAR_URL)
    }
  }

  getFriends(): Observable<User[]> {
    return this.http.get<User[]>('/api/friends');
  }

  hasFriendsPermissions(): Observable<any> {
    return this.http.get<any>('/api/friends/permissions');
  }

  getUser(userId: string): Observable<User> {
    return this.http.get<User>(`/api/users/${userId}`);
  }

  getUsers(userIds: string[]): Observable<User[]> {
    return this.http.get<User[]>(`/api/users?${userIds.map((id) => 'userId=' + id).join('&')}`);
  }

  setToken(token: string): void {
    const decoded = jwt_decode(token);
    localStorage.setItem(USER_ID, decoded.jti);
    localStorage.setItem(USER_NAME, decoded.iss);
    localStorage.setItem(AVATAR_URL, decoded.avatar);
    localStorage.setItem(ACCESS_TOKEN, token);
    this.cookies.set(ACCESS_TOKEN, token, this.getTokenExpirationDate(token), '/')
  }

  getTokenExpirationDate(token: string): Date {
    const decoded = jwt_decode(token);

    if (decoded.exp === undefined) return null;

    const date = new Date(0);
    date.setUTCSeconds(decoded.exp);
    return date;
  }

  isTokenExpired(token?: string): boolean {
    if (!token) token = this.getToken();
    if (!token) return true;

    const date = this.getTokenExpirationDate(token);
    if (date === undefined) return false;
    return !(date.valueOf() > new Date().valueOf());
  }

  loginWith(provider: string, isPopup: boolean, force: boolean, redirectBack: boolean) {
    const extAuthURL = this.buildClientAuthUri(env.oauth[provider], provider, isPopup, force, redirectBack);
    if (isPopup) {
      window.open(
        extAuthURL,
        `Login with ${provider}`,
        'width=400,height=600');
    } else {
      window.location.replace(extAuthURL)
    }
  }

  buildClientAuthUri(conf, provider, isPopup: boolean, force: boolean, redirectBack: boolean) {
    const {authUri, clientId, scope} = conf;
    const redirectUri = env.oauth.redirectUri;

    const rawState = btoa(JSON.stringify({
      randomString: Math.random().toString(36).slice(2),
      oauthRedirectUrl: redirectBack ? window.location.href : `${env.oauth.oAuthRedirectUriBase}/${provider}`,
      isPopup: isPopup
    }));
    this.cookies.set('oauth_state', rawState, null, '/');
    const state = (<any>window).encodeURIComponent(rawState);
    var resultUrl = `${authUri}?client_id=${clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scope}&response_type=code&display=popup`;
    if (force) {
      resultUrl += '&auth_type=rerequest';
    }
    return resultUrl;
  }
}
