import {Message} from './message';
import {User} from './user';
import {Room} from "./room";
import {AuthService} from "../services/auth.service";
import {ChatService} from "../services/chat.service";
import "rxjs/add/operator/mergeMap";
import {EventEmitter} from "@angular/core";

export class RoomContainer {
  messages: Message[] = [];
  accounts: Map<string, User> = new Map<string, User>();
  loading: boolean;
  errors: string[] = [];
  newMessage: string;
  usersToFetch: string[] = [];
  onMessageDelivered: EventEmitter<Message> = new EventEmitter<Message>();
  onMessageReceived: EventEmitter<Message> = new EventEmitter<Message>();

  constructor(public me: User,
              public room: Room,
              public auth: AuthService,
              public chat: ChatService) {
    this.errors = [];
    this.loading = true;
    this.newMessage = '';
    auth.getUsers(this.room.members).subscribe((users: User[]) => {
        this.loading = false;
        if (users) {
          users.forEach(user => this.accounts.set(user.id, user))
        }
      }, (error) => {
        this.loading = false;
        this.errors = [error.message];
      }
    );
    this.fetchUsers();
  }

  addMember(friend: User) {
    this.loading = true;
    this.chat.addMember(this.room.id, friend.id).subscribe(() => {
      this.loading = false;
      this.room.members.push(friend.id);
      this.accounts.set(friend.id, friend);
      this.chat.send(new Message(this.room.id, this.me.id, '', Date.now(), 'update'))
    }, (error) => {
      this.loading = false;
      this.errors = [error.message];
    })
  }

  onMessage(message: Message) {
    this.messages.push(message);
    if (message.from == this.me.id) {
      this.onMessageDelivered.emit(message);
    } else {
      this.onMessageReceived.emit(message);
    }
  }

  amIOwner() {
    return this.me.id == this.room.owner
  }

  getUser(userId: string): User {
    let user = this.accounts.get(userId);
    if (!user) {
      this.usersToFetch.push(userId);
    }
    return user
  }

  fetchUsers() {
    if (this.usersToFetch.length > 0) {
      this.auth.getUsers(this.usersToFetch).subscribe((users) => {
        if (users && users.length > 0) {
          users.forEach(user => {
            this.accounts.set(user.id, user);
            this.usersToFetch.slice(this.usersToFetch.indexOf(user.id), 1);
          });
        }
        // populate unknown users
        this.usersToFetch.forEach(unknownUserId => {
          this.accounts.set(unknownUserId, {
            id: unknownUserId,
            name: 'Unknown',
            online: false,
            avatarUrl: null
          })
        });
        this.usersToFetch = [];
        setTimeout(this.fetchUsers, 60000);
      })
    }
  }

  kickMember(userId: string) {
    this.loading = true;
    this.chat.kickMember(this.room.id, userId).subscribe(() => {
      this.loading = false;
      this.room.members.slice(this.room.members.indexOf(userId), 1);
      this.chat.send(new Message(this.room.id, this.me.id, '', Date.now(), 'update'))
    }, (error) => {
      this.loading = false;
      this.errors = [error.message];
    })
  }
}
