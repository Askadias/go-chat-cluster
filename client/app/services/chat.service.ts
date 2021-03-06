import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';
import {Message} from "../domain/message";
import {Room} from "../domain/room";
import {MemberInfo} from "../domain/member-info";

@Injectable()
export class ChatService {

  private socket: Observable<Message>;

  constructor(private http: HttpClient) {
    const loc = window.location;
    this.socket = Observable.webSocket(`${loc.protocol === 'https:' ? 'wss:' : 'ws:'}//${loc.host}/api/ws`);
  }

  getChatLog(roomId: string, from: number, limit: number): Observable<Message[]> {
    return this.http.get<Message[]>(`/api/rooms/${roomId}/log?from=${from}&limit=${limit}`);
  }

  send(message: Message) {
    return this.http.post<any>(`/api/rooms/${message.room}/log`, message);
  }

  newRoom(room: Room): Observable<Room> {
    return this.http.post<Room>(`/api/rooms`, room);
  }

  addMember(roomId: string, memberId: string): Observable<Room> {
    return this.http.post<Room>(`/api/rooms/${roomId}/members/${memberId}`, {});
  }

  kickMember(roomId: string, memberId: string): Observable<any> {
    return this.http.delete(`/api/rooms/${roomId}/members/${memberId}`);
  }

  getRooms(): Observable<Room[]> {
    return this.http.get<Room[]>(`/api/rooms`);
  }

  getAllMembersInfo(roomId: string): Observable<MemberInfo[]> {
    return this.http.get<MemberInfo[]>(`/api/rooms/${roomId}/members/info`);
  }

  getAllRoomsInfo(): Observable<MemberInfo[]> {
    return this.http.get<MemberInfo[]>(`/api/rooms/members/info`);
  }

  getMemberInfo(roomId: string, memberId: string): Observable<MemberInfo> {
    return this.http.get<MemberInfo>(`/api/rooms/${roomId}/members/${memberId}/info`);
  }

  updateLastReadTime(roomId: string, memberId: string): Observable<any> {
    return this.http.put<any>(`/api/rooms/${roomId}/members/${memberId}/info`, {});
  }

  getRoom(roomId: string): Observable<Room> {
    return this.http.get<Room>(`/api/rooms/${roomId}`);
  }

  deleteRoom(roomId: string): Observable<any> {
    return this.http.delete(`/api/rooms/${roomId}`);
  }

  getSocket(): Observable<Message> {
    return this.socket;
  }
}
