import {AfterViewChecked, Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormControl} from '@angular/forms';
import {Message} from '../../domain/message';
import {User} from '../../domain/user';
import {Room} from '../../domain/room';
import {RoomContainer} from "../../domain/room-container";

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, AfterViewChecked {

  @Input() className = '';
  @Input() room: RoomContainer;
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();
  messageInput = new FormControl('', []);
  members: User[];
  messages: Message[];
  me: User;

  @ViewChild('chatLogElement') private chatLogContainer: ElementRef;

  constructor() {
  }

  ngOnInit() {
    this.scrollToBottom();
  }

  ngAfterViewChecked() {
    this.scroll();
  }

  sendMessage() {
    if (this.messageInput.valid && this.room.newMessage !== '') {
      this.room.chat.send(new Message(this.room.room.id, this.room.me.id, this.room.newMessage, Date.now()));
      this.room.newMessage = '';
      this.scrollToBottom();
    }
  }

  getAvatar(userId: string): string {
    const account = this.room.getUser(userId);
    return account ? account.avatarUrl : null;
  }

  isHeadMessage(chatMessage: Message, index: number) {
    return (this.room.messages.length === index + 1) // last message
      || this.room.messages[index + 1].from !== this.room.messages[index].from;
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
