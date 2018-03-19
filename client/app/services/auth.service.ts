import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';
import * as jwt_decode from 'jwt-decode';
import {Router} from "@angular/router";

export const ACCESS_TOKEN: string = 'JWT';
export const USER_NAME: string = 'USER_NAME';
export const AVATAR_URL: string = 'AVATAR_URL';

@Injectable()
export class AuthService {

  constructor(private http: HttpClient, private router: Router) {
  }

  logout() {
    localStorage.removeItem(ACCESS_TOKEN);
    localStorage.removeItem(USER_NAME);
    localStorage.removeItem(AVATAR_URL);
    this.router.navigate(['/login']);
  }

  loginWithExtAuth(provider, code: string): Observable<any> {
    return this.http.post<any>(`/api/login/${provider}`, {code});
  }

  getToken(): string {
    return localStorage.getItem(ACCESS_TOKEN);
  }

  setToken(token: string): void {
    const decoded = jwt_decode(token);
    localStorage.setItem(USER_NAME, decoded.iss);
    localStorage.setItem(AVATAR_URL, decoded.avatar);
    localStorage.setItem(ACCESS_TOKEN, token);
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
