import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../domain/user";
import {AuthService} from "../services/auth.service";
import {ChatRoom} from "../domain/chat-room";

@Component({
  selector: 'chat-component',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {

  protected oauthConfig: any;
  errors: string[] = [];
  profile: User;
  friends: User[] = [];
  chats: ChatRoom[] = [];
  activeRoom: ChatRoom;
  loading = false;
  foldSocialBar = false;
  chatOpened = false;

  constructor(public route: ActivatedRoute,
              private auth: AuthService,
              private chat: ChatService) {
    this.oauthConfig = env.oauth;
    this.loading = true;
    this.profile = auth.getProfile();
    this.activeRoom = new ChatRoom(this.profile, null);
    this.chat.getFriends().subscribe(
      (friends) => {
        this.loading = false;
        this.friends = friends;
      }, (response) => {
        this.loading = false;
        this.errors = [response.message];
      }
    );
  }

  ngOnInit() {
    this.chat.getEventListener().subscribe(event => {
      if (event.type === 'message') {
        let data = event.data.content;
        if (event.data.roomId) {
          const targetChat = this.chats.find((chat) =>
            chat.id === event.data.roomId
          );
          if (targetChat) {
            let senderId = event.data.sender;
            if (senderId) {
              let sender = targetChat.participants.find((user) =>
                user.id === senderId
              );
              if (sender) {
                targetChat.onMessageReceive(data, sender.id)
              }
            }
          }
        }
      }
      if(event.type == "close") {
        if (event.data.roomId) {
          const targetChat = this.chats.find((chat) =>
            chat.id === event.data.roomId
          );
          if (targetChat) {
            let senderId = event.data.sender;
            if (senderId) {
              let sender = targetChat.participants.find((user) =>
                user.id === senderId
              );
              this.removeFromChat(targetChat, sender)
            }
          }
        }
      }
      if(event.type == "open") {
        // this.messages.push("/The socket connection has been established");
      }
    })
  }

  logout() {
    this.auth.logout()
  }

  chatWith(friend: User) {
    const existingChat = this.chats.find((chat) =>
      chat.participants.length === 1
      && chat.participants[0].id === friend.id
    );
    if (existingChat) {
      if (this.activeRoom.id !== existingChat.id) {
        this.switchToChat(existingChat)
      }
    } else {
      let newChat = new ChatRoom(this.profile, friend);
      this.chats.push(newChat);
      this.switchToChat(newChat)
    }
  }

  addToCurrentChat(friend: User) {
    this.activeRoom.addParticipant(friend)
  }

  removeFromChat(chat: ChatRoom, friend: User) {
    if (chat.participants.length === 1) {
      this.dismissChat(chat);
    } else {
      chat.excludeParticipant(friend);
    }
  }

  isActiveChat(chat: ChatRoom) {
    return this.activeRoom && chat.id === this.activeRoom.id;
  }

  switchToChat(chat: ChatRoom) {
    this.activeRoom = chat;
    this.chatOpened = true;
  }

  canAddToCurrentChat(friend: Account) {
    return this.activeRoom.participants.length > 0 && !this.activeRoom.accounts.has(friend.id);
  }

  closeChat() {
    this.chatOpened = false;
    this.activeRoom = new ChatRoom(this.profile, null);
  }

  dismissChat(chat: ChatRoom) {
    this.chats.splice(
      this.chats.findIndex((it) => it.id === chat.id),
      1
    );
    if (this.activeRoom.id === chat.id) {
      this.closeChat();
    }
  }

  trackByUserId(index: number, friend: User): string {
    return friend.id;
  }
}
