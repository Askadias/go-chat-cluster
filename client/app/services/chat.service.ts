import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';

@Injectable()
export class ChatService {

  constructor(private http: HttpClient) {
  }

  getFriends(): Observable<any> {
    return this.http.get<any>('/api/friends');
  }
}
