import {AfterViewChecked, Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {Message} from '../../domain/message';
import {User} from '../../domain/user';
import {RoomContainer} from "../../domain/room-container";
import {environment as env} from "../../../environments/environment";
import {EmojiEvent} from "@ctrl/ngx-emoji-mart/ngx-emoji";

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, AfterViewChecked {

  @Input() className = '';
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();

  @ViewChild('chatLogElement') private chatLogContainer: ElementRef;

  _room: RoomContainer;
  members: User[];
  messages: Message[];
  me: User;
  loading = false;
  errors: string[] = [];
  emojiPickerOpened = false;

  constructor() {
  }

  @Input()
  set room(room: RoomContainer) {
    this._room = room;
    setTimeout(() => {
      const native = this.chatLogContainer.nativeElement;
      native.scrollTop = native.scrollHeight;
    });
  }

  ngOnInit() {
  }

  ngAfterViewChecked() {
    this.scroll();
  }

  sendMessage() {
    if (this._room.newMessage !== '') {
      this.loading = true;
      this.errors = [];
      this._room.chat.send(new Message(this._room.room.id, this._room.me.id, this._room.newMessage, Date.now()))
        .subscribe(() => {
            this.loading = false;
          },
          (error) => {
            this.loading = false;
            this.errors = [error.message];
          });
      this._room.newMessage = '';
      this.scrollToBottom();
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

  scroll(): void {
    const native = this.chatLogContainer.nativeElement;
    if (native.scrollHeight - (native.scrollTop + native.clientHeight) < 100) {
      this.scrollToBottom();
    }
  }

  scrollToBottom(): void {
    try {
      const native = this.chatLogContainer.nativeElement;
      native.scrollTop = native.scrollHeight;
    } catch (err) {
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
}
