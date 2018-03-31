import {EventEmitter, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs/Observable';
import {AuthService} from "./auth.service";
import {Message} from "../domain/message";
import {Room} from "../domain/room";
import {User} from "../domain/user";

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

  getChatLog(roomId: string, from: number, limit: number): Observable<Message[]> {
    return this.http.get<Message[]>(`/api/rooms/${roomId}/log?from${from}&limit=${limit}`);
  }

  newRoom(room: Room): Observable<Room> {
    return this.http.post<Room>(`/api/rooms`, room);
  }

  addMember(roomId: string, memberId: string): Observable<any> {
    return this.http.post(`/api/rooms/${roomId}/members/${memberId}`, {});
  }

  kickMember(roomId: string, memberId: string): Observable<any> {
    return this.http.delete(`/api/rooms/${roomId}/members/${memberId}`);
  }

  getRooms(): Observable<Room[]> {
    return this.http.get<Room[]>(`/api/rooms`);
  }

  getRoom(roomId: string): Observable<Room> {
    return this.http.get<Room>(`/api/rooms/${roomId}`);
  }

  deleteRoom(roomId: string): Observable<any> {
    return this.http.delete(`/api/rooms/${roomId}`);
  }

  send(message: Message) {
    this.socket.send(JSON.stringify(message));
  }

  close() {
    this.socket.close();
  }

  getEventListener() {
    return this.listener;
  }
}
