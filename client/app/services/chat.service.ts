import {EventEmitter, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';
import {AuthService} from "./auth.service";
import {ChatMessage} from "../domain/chat-message";

@Injectable()
export class ChatService {

  private socket: WebSocket;
  private listener: EventEmitter<any> = new EventEmitter();

  constructor(private http: HttpClient) {
    const loc = window.location;
    this.socket = new WebSocket(`${loc.protocol === 'https:' ? 'wss:' : 'ws:'}//${loc.host}/api/ws`);
    this.socket.onopen = event => {
      this.listener.emit({type: 'open', data: event});
    };
    this.socket.onclose = event => {
      this.listener.emit({type: 'close', data: event});
    };
    this.socket.onmessage = event => {
      this.listener.emit({type: 'message', data: JSON.parse(event.data)});
    };
  }

  getFriends(): Observable<any> {
    return this.http.get<any>('/api/friends');
  }

  getChatLog(roomId: string, from: number, limit: number): Observable<ChatMessage[]> {
    return this.http.get<ChatMessage[]>(`/api/rooms/${roomId}/log?from${from}&limit=${limit}`);
  }

  send(message: ChatMessage) {
    this.socket.send(JSON.stringify(message));
  }

  close() {
    this.socket.close();
  }

  getEventListener() {
    return this.listener;
  }
}
