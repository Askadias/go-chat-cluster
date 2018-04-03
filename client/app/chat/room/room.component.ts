import {AfterViewChecked, Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormControl} from '@angular/forms';
import {Message} from '../../domain/message';
import {User} from '../../domain/user';
import {RoomContainer} from "../../domain/room-container";

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, AfterViewChecked {

  @Input() className = '';
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();
  _room: RoomContainer;
  messageInput = new FormControl('', []);
  members: User[];
  messages: Message[];
  me: User;
  loading = false;
  errors: string[] = [];

  @ViewChild('chatLogElement') private chatLogContainer: ElementRef;

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
    if (this.messageInput.valid && this._room.newMessage !== '') {
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
}
