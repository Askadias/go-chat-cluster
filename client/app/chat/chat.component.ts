import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../domain/user";
import {AuthService} from "../services/auth.service";
import {Room} from "../domain/room";
import {RoomContainer} from "../domain/room-container";

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
  rooms: RoomContainer[] = [];
  activeRoom: RoomContainer;
  loading = false;
  foldSocialBar = false;
  chatOpened = false;

  constructor(public route: ActivatedRoute,
              private auth: AuthService,
              private chat: ChatService) {
    this.oauthConfig = env.oauth;
    this.loading = true;
    this.profile = auth.getProfile();
    this.activeRoom = null;
    this.auth.getFriends().subscribe(
      (friends) => {
        this.loading = false;
        this.friends = friends;
        this.chat.getRooms().subscribe(
          (rooms) => {
            this.loading = false;
            if (rooms) {
              this.rooms = rooms.map((room) => new RoomContainer(this.profile, room, this.auth, this.chat));
            }
          }, (response) => {
            this.loading = false;
            this.errors = [response.message];
          }
        );
      }, (response) => {
        this.loading = false;
        this.errors = [response.message];
      }
    );
  }

  ngOnInit() {
    this.chat.getEventListener().subscribe(event => {
      if (event.type === 'message') {
        if (event.data.room) {
          if (event.data.type === 'update') {
            this.loading = false;
            this.chat.getRoom(event.data.room).subscribe(
              (room) => {
                this.loading = false;
                const roomContainer = new RoomContainer(this.profile, room, this.auth, this.chat);
                const idx = this.rooms.findIndex(roomContainer =>
                  roomContainer.room.id === event.data.room
                );
                if (idx > -1) {
                  this.rooms[idx] = roomContainer;
                  if (this.activeRoom && this.activeRoom.room.id == room.id) {
                    this.activeRoom = roomContainer;
                  }
                } else {
                  this.rooms.push(roomContainer);
                }
              }, (response) => {
                this.loading = false;
                if (response.status == 404) {
                  this.dismissChat(this.rooms.find(rm => rm.room.id === event.data.room))
                } else {
                  this.errors = [response.message];
                }
              }
            )
          } else {
            const targetChat = this.rooms.find((roomContainer) =>
              roomContainer.room.id === event.data.room
            );
            if (targetChat) {
              targetChat.onMessageReceive(event.data)
            } else {
              this.loading = false;
              this.chat.getRoom(event.data.room).subscribe(
                (room) => {
                  this.loading = false;
                  this.rooms.push(new RoomContainer(this.profile, room, this.auth, this.chat));
                }, (response) => {
                  this.loading = false;
                  this.errors = [response.message];
                }
              )
            }
          }
        }
      }
      if (event.type == "close") {
        if (event.data.room) {
          const targetChat = this.rooms.find((roomContainer) =>
            roomContainer.room.id === event.data.room
          );
          if (targetChat) {
            let senderId = event.data.from;
            if (senderId) {
              let sender = targetChat.accounts.get(senderId);
              if (sender) {
                sender.online = false
              }
            }
          }
        }
      }
      if (event.type == "open") {
        if (event.data.room) {
          const targetChat = this.rooms.find((roomContainer) =>
            roomContainer.room.id === event.data.room
          );
          if (targetChat) {
            let senderId = event.data.from;
            if (senderId) {
              let sender = targetChat.accounts.get(senderId);
              if (sender) {
                sender.online = true
              }
            }
          }
        }
      }
    })
  }

  logout() {
    this.auth.logout()
  }

  chatWith(userId: string) {
    const existingRoom = this.rooms.find((roomContainer) =>
      roomContainer.room.members.length === 2
      && !!roomContainer.room.members.find((id) => id == userId)
    );
    if (existingRoom) {
      if (this.activeRoom.room.id !== existingRoom.room.id) {
        this.switchToChat(existingRoom)
      }
    } else {
      this.chat.newRoom(new Room(this.profile.id, userId)).subscribe(
        (newRoom) => {
          this.loading = false;
          const roomContainer = new RoomContainer(this.profile, newRoom, this.auth, this.chat);
          this.rooms.push(roomContainer);
          this.switchToChat(roomContainer)
        }, (response) => {
          this.loading = false;
          this.errors = [response.message];
        }
      );
    }
  }

  addToCurrentChat(friend: User) {
    this.activeRoom.addMember(friend);
  }

  removeFromChat(roomContainer: RoomContainer, userId: string) {
    if (roomContainer.room.members.length === 1) {
      this.dismissChat(roomContainer);
    } else {
      roomContainer.kickMember(userId);
    }
  }

  isActiveChat(chat: RoomContainer) {
    return this.activeRoom && chat.room.id === this.activeRoom.room.id;
  }

  switchToChat(chat: RoomContainer) {
    this.activeRoom = chat;
    this.chatOpened = true;
  }

  canAddToCurrentChat(userId: string) {
    return this.activeRoom != null
      && this.activeRoom.room.members.length > 0
      && this.activeRoom.room.members.indexOf(userId) == -1;
  }

  closeChat() {
    this.chatOpened = false;
    this.activeRoom = null;
  }

  dismissChat(chat: RoomContainer) {
    if (chat) {
      this.rooms.splice(
        this.rooms.findIndex((it) => it.room.id === chat.room.id),
        1
      );
      if (this.activeRoom.room.id === chat.room.id) {
        this.closeChat();
      }
    }
  }

  trackByUserId(index: number, friend: User): string {
    return friend.id;
  }
}
