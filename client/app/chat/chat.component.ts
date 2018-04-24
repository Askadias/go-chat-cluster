import {ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../domain/user";
import {AuthService} from "../services/auth.service";
import {Room} from "../domain/room";
import {RoomContainer} from "../domain/room-container";
import {Message} from "../domain/message";
import {MediaMatcher} from '@angular/cdk/layout';
import {exponentialBackOff} from "../common/utils";
import {MatDialog, MatTabChangeEvent} from "@angular/material";
import {ConfirmDialog} from "../common/confirm/confirm-dialog.component";

const FRIENDS_IDX = 0;
const CHATS_IDX = 1;

@Component({
  selector: 'chat-component',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit, OnDestroy {

  protected oauthConfig: any;
  private readonly isPopup = true;
  private readonly _mediaQueryListener: () => void;
  mobileQuery: MediaQueryList;
  tabletQuery: MediaQueryList;
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
  activeTab = CHATS_IDX;

  constructor(public route: ActivatedRoute,
              public confirmDialog: MatDialog,
              private auth: AuthService,
              private chat: ChatService,
              changeDetectorRef: ChangeDetectorRef,
              media: MediaMatcher) {
    this.mobileQuery = media.matchMedia('(max-width: 600px)');
    this.tabletQuery = media.matchMedia('(max-width: 1200px)');
    this._mediaQueryListener = () => changeDetectorRef.detectChanges();
    this.mobileQuery.addListener(this._mediaQueryListener);
    this.tabletQuery.addListener(this._mediaQueryListener);

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
          this.auth.hasFriendsPermissions().subscribe(
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
    this.chat.getSocket().retryWhen(exponentialBackOff).subscribe((message: Message) => {
      if (message.room) {
        if (message.type === 'update') {
          this.chat.getMemberInfo(message.room, this.profile.id).subscribe(
            memberInfo => {
              this.chat.getRoom(memberInfo.room).subscribe(
                (room) => {
                  room.memberInfo = memberInfo;
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
                  if (error.status == 403 || error.status == 404) {
                    this.removeRoomFromPool(message.room)
                  } else {
                    this.errors = [error.message];
                  }
                }
              )
            },(error) => {
              if (error.status == 403 || error.status == 404) {
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
            targetChat.onMessage(message)
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

  ngOnDestroy(): void {
    this.mobileQuery.removeListener(this._mediaQueryListener);
  }

  isMobile(): boolean {
    return this.mobileQuery.matches
  }

  isTablet(): boolean {
    return this.tabletQuery.matches
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
      this.switchToChat(existingRoom)
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

  onStartChat(event) {
    this.chatWith(event.memberId)
  }

  addToCurrentChat(friend: User) {
    this.addMember(friend);
    this.activeTab = CHATS_IDX;
  }

  addMember(friend: User) {
    this.chat.addMember(this.activeRoom.room.id, friend.id).subscribe((room) => {
      if (room.id != this.activeRoom.room.id) {
        const roomContainer = new RoomContainer(this.profile, room, this.auth, this.chat);
        this.rooms.push(roomContainer);
        this.switchToChat(roomContainer)
      } else {
        // this.activeRoom.room.members.push(friend.id);
        // this.activeRoom.accounts.set(friend.id, friend);
      }
    }, (error) => {
      this.errors = [error.message];
    })
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
    this.activeTab = CHATS_IDX;
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
    let confirmDialogRef = this.confirmDialog.open(ConfirmDialog, {
      width: '320px',
      data: { message: "Are you sure you want to delete the chat room and its conversation history?" }
    });

    confirmDialogRef.afterClosed().subscribe(confirm => {
      if (confirm) {
        this.loadingRooms = true;
        this.chat.deleteRoom(roomContainer.room.id).subscribe(() => {
            this.loadingRooms = false;
            // this.removeRoomFromPool(roomContainer.room.id);
          }, (error) => {
            this.loadingRooms = false;
            this.errors = [error.message];
          }
        )
      }
    });
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

  onTabChange(e: MatTabChangeEvent) {
    this.activeTab = e.index;
  }
}
