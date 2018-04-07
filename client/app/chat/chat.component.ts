import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../domain/user";
import {AuthService} from "../services/auth.service";
import {Room} from "../domain/room";
import {RoomContainer} from "../domain/room-container";
import {Message} from "../domain/message";
import {range} from 'rxjs/observable/range'
import {zip} from 'rxjs/observable/zip'
import {timer} from 'rxjs/observable/timer'

@Component({
  selector: 'chat-component',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {

  protected oauthConfig: any;
  private isPopup = true;
  errors: string[] = [];
  profile: User;
  friends: User[] = [];
  rooms: RoomContainer[] = [];
  activeRoom: RoomContainer;
  loadingFriends = false;
  loadingRooms = false;
  foldSocialBar = false;
  chatOpened = false;
  hasFriendsPermissions = false;

  constructor(public route: ActivatedRoute,
              private auth: AuthService,
              private chat: ChatService) {
    this.oauthConfig = env.oauth;
    this.loadingFriends = true;
    this.profile = auth.getProfile();
    this.activeRoom = null;
    const routeSnapshot = this.route.snapshot;
    this.isPopup = routeSnapshot.queryParams.isPopup || true;
    this.auth.getFriends().subscribe(
      (friends) => {
        this.loadingFriends = false;
        this.friends = friends;
        if (!friends || friends.length == 0) {
          this.chat.getRooms().subscribe(
            () => {
              this.hasFriendsPermissions = false;
            }, (error) => {
              this.hasFriendsPermissions = false;
              this.errors = [error.message];
            }
          )
        } else {
          this.hasFriendsPermissions = true;
          this.loadingRooms = true;
          this.chat.getRooms().subscribe(
            (rooms) => {
              this.loadingRooms = false;
              if (rooms) {
                this.rooms = rooms.map((room) => new RoomContainer(this.profile, room, this.auth, this.chat));
              }
            }, (error) => {
              this.loadingRooms = false;
              this.errors = [error.message];
            }
          );
        }
      }, (error) => {
        this.loadingFriends = false;
        this.errors = [error.message];
      }
    );
  }

  ngOnInit() {
    this.chat.getSocket().retryWhen(function (errors) {
      return zip(
        range(1, env.socket.maxSubscriptionRetries), errors, function (i, e) {
          return i
        })
        .flatMap(function (i) {
          console.log("delay retry by " + i + " second(s)");
          return timer(i * 1000);
        });
    }).subscribe((message: Message) => {
      if (message.room) {
        if (message.type === 'update') {
          this.loadingRooms = false;
          this.chat.getRoom(message.room).subscribe(
            (room) => {
              this.loadingRooms = false;
              const roomContainer = new RoomContainer(this.profile, room, this.auth, this.chat);
              const idx = this.rooms.findIndex(roomContainer =>
                roomContainer.room.id === message.room
              );
              if (idx > -1) {
                this.rooms[idx] = roomContainer;
                if (this.activeRoom && this.activeRoom.room.id == room.id) {
                  this.activeRoom = roomContainer;
                }
              } else {
                this.rooms.push(roomContainer);
              }
            }, (error) => {
              this.loadingRooms = false;
              if (error.status == 404) {
                this.removeRoomFromPool(message.room)
              } else {
                this.errors = [error.message];
              }
            }
          )
        } else {
          const targetChat = this.rooms.find((roomContainer) =>
            roomContainer.room.id === message.room
          );
          if (targetChat) {
            targetChat.onMessageReceive(message)
          } else {
            this.loadingRooms = false;
            this.chat.getRoom(message.room).subscribe(
              (room) => {
                this.loadingRooms = false;
                this.rooms.push(new RoomContainer(this.profile, room, this.auth, this.chat));
              }, (error) => {
                this.loadingRooms = false;
                this.errors = [error.message];
              }
            )
          }
        }
      }
    })
  }

  loginWith(provider: string) {
    this.auth.loginWith(provider, this.isPopup, true, true);
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
          this.loadingRooms = false;
          const roomContainer = new RoomContainer(this.profile, newRoom, this.auth, this.chat);
          this.rooms.push(roomContainer);
          this.switchToChat(roomContainer)
        }, (error) => {
          this.loadingRooms = false;
          this.errors = [error.message];
        }
      );
    }
  }

  addToCurrentChat(friend: User) {
    this.activeRoom.addMember(friend);
  }

  removeFromChat(roomContainer: RoomContainer, userId: string) {
    if (roomContainer.room.members.length === 1) {
      this.removeRoomFromPool(roomContainer.room.id);
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

  dismissChat(roomContainer: RoomContainer) {
    this.loadingRooms = true;
    this.chat.deleteRoom(roomContainer.room.id).subscribe(() => {
        this.loadingRooms = false;
        this.removeRoomFromPool(roomContainer.room.id);
      }, (error) => {
        this.loadingRooms = false;
        this.errors = [error.message];
      }
    )
  }

  removeRoomFromPool(roomId) {
    this.rooms.splice(
      this.rooms.findIndex((it) => it.room.id === roomId),
      1
    );
    if (this.activeRoom.room.id === roomId) {
      this.closeChat();
    }
  }

  trackByUserId(index: number, friend: User): string {
    return friend.id;
  }
}
