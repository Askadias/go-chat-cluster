import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';
import * as jwt_decode from 'jwt-decode';
import {Router} from "@angular/router";
import {User} from "../domain/user";
import {CookieService} from "ngx-cookie-service";

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
}
