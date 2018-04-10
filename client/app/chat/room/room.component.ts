import {Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {Message} from '../../domain/message';
import {User} from '../../domain/user';
import {RoomContainer} from "../../domain/room-container";
import {environment as env} from "../../../environments/environment";
import {EmojiEvent} from "@ctrl/ngx-emoji-mart/ngx-emoji";
import {ChatService} from "../../services/chat.service";

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit {

  @Input() className = '';
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();

  @ViewChild('chatLogElement') private chatLogContainer: ElementRef;

  _room: RoomContainer;
  members: User[];
  me: User;
  loading = false;
  sending = false;
  errors: string[] = [];
  emojiPickerOpened = false;
  initialized = false;

  constructor(private chat: ChatService) {
  }

  @Input()
  set room(room: RoomContainer) {
    this._room = room;
    this.initialized = false;
    this._room.messages = [];
    this.loadMessages();
    this._room.onMessageDelivered.subscribe(() => {
      setTimeout(() => {
        const native = this.chatLogContainer.nativeElement;
        native.scrollTop = native.scrollHeight;
      });
    });
    this._room.onMessageReceived.subscribe(() => {
      setTimeout(() => {
        const native = this.chatLogContainer.nativeElement;
        if (native.scrollHeight - (native.scrollTop + native.clientHeight) < 100) {
          native.scrollTop = native.scrollHeight;
        }
      });
    })
  }

  loadMessages() {
    let from = Date.now();
    if (this._room.messages.length > 1) {
      from = this._room.messages[0].timestamp;
    }
    this.loading = true;
    this.chat.getChatLog(this._room.room.id, from, env.chat.messagesLimit).subscribe((messages: Message[]) => {
        this.loading = false;
        if (messages) {
          messages.reverse().push(...this._room.messages);
          this._room.messages = messages;
          if (!this.initialized) { // scroll to the last message on chat opening
            setTimeout(() => {
              const native = this.chatLogContainer.nativeElement;
              native.scrollTop = native.scrollHeight;
            });
          } else {
            setTimeout(() => {
              const native = this.chatLogContainer.nativeElement;
              native.scrollTop = 10;
            });
          }
        }
        this.initialized = true;
      }, (error) => {
        this.loading = false;
        this.errors = [error.message];
        this.initialized = true;
      }
    );
  }

  ngOnInit() {
  }

  sendMessage() {
    if (this._room.newMessage !== '') {
      this.sending = true;
      this.errors = [];
      this._room.chat.send(new Message(this._room.room.id, this._room.me.id, this._room.newMessage, Date.now()))
        .subscribe(() => {
            this.sending = false;
          },
          (error) => {
            this.sending = false;
            this.errors = [error.message];
          });
      this._room.newMessage = '';
    }
  }

  getAvatar(userId: string): string {
    const account = this._room.getUser(userId);
    return account ? account.avatarUrl : null;
  }

  isHeadMessage(chatMessage: Message, index: number) {
    return (this._room.messages.length === index + 1) // last message
      || this._room.messages[index + 1].from !== this._room.messages[index].from;
  }

  getTimestamp(chatMessage: Message) {
    return chatMessage.timestamp * 1000
  }

  isNearestMessage(chatMessage: Message, index: number) {
    return (index < this._room.messages.length - 1) // last message
      && this._room.messages[index + 1].from === this._room.messages[index].from
      && this._room.messages[index + 1].timestamp - this._room.messages[index].timestamp < env.chat.closeMessagesRangeSec;
  }

  loadHistory() {
    if (this.initialized) {
      const native = this.chatLogContainer.nativeElement;

      if (native.scrollTop < 1) {
        this.loadMessages()
      }
    }
  }

  toggleEmojiPicker(): void {
    this.emojiPickerOpened = !this.emojiPickerOpened;
  }

  pickEmoji(event: EmojiEvent): void {
    if (event) {
      this._room.newMessage += event.emoji.native
    }
  }

  trackByMessageId(index: number, message: Message): string {
    return message.id;
  }
}
