import {AuthService} from "../services/auth.service";
import {
  HttpErrorResponse,
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest,
  HttpResponse
} from "@angular/common/http";
import {Observable} from "rxjs/Observable";
import "rxjs/add/operator/do";
import {Injector} from "@angular/core";

export class AuthInterceptor implements HttpInterceptor {

  constructor(private injector: Injector) {
  }

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const auth = this.injector.get(AuthService);
    request = request.clone({
      setHeaders: {
        Authorization: `Bearer ${auth.getToken()}`
      }
    });
    return next.handle(request).do((event: HttpEvent<any>) => {
      if (event instanceof HttpResponse) {
        // do stuff with response if you want
      }
    }, (err: any) => {
      if (err instanceof HttpErrorResponse) {
        if (err.status === 401) {
          auth.logout();
        }
      }
    });
  }
}
