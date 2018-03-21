import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../domain/user";
import {AuthService} from "../services/auth.service";
import {Chat} from "../domain/chat";

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
  chats: Chat[] = [];
  activeChat: Chat;
  loading = false;
  foldSocialBar = false;
  chatOpened = false;

  constructor(public route: ActivatedRoute,
              private auth: AuthService,
              private chat: ChatService) {
    this.oauthConfig = env.oauth;
    this.loading = true;
    this.profile = auth.getProfile();
    this.activeChat = new Chat(this.profile, null);
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
      if (this.activeChat.id !== existingChat.id) {
        this.switchToChat(existingChat)
      }
    } else {
      let newChat = new Chat(this.profile, friend);
      this.chats.push(newChat);
      this.switchToChat(newChat)
    }
  }

  addToCurrentChat(friend: User) {
    this.activeChat.addParticipant(friend)
  }

  removeFromChat(chat: Chat, friend: User) {
    if (chat.participants.length === 1) {
      this.dismissChat(chat);
    } else {
      chat.excludeParticipant(friend);
    }
  }

  isActiveChat(chat: Chat) {
    return this.activeChat && chat.id === this.activeChat.id;
  }

  switchToChat(chat: Chat) {
    this.activeChat = chat;
    this.chatOpened = true;
  }

  canAddToCurrentChat(friend: Account) {
    return this.activeChat.participants.length > 0 && !this.activeChat.accounts.has(friend.id);
  }

  closeChat() {
    this.chatOpened = false;
    this.activeChat = new Chat(this.profile, null);
  }

  dismissChat(chat: Chat) {
    this.chats.splice(
      this.chats.findIndex((it) => it.id === chat.id),
      1
    );
    if (this.activeChat.id === chat.id) {
      this.closeChat();
    }
  }
}
