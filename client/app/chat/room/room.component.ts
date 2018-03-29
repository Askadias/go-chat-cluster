import {AfterViewChecked, Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormControl} from '@angular/forms';
import {ChatMessage} from '../../domain/chat-message';
import {User} from '../../domain/user';
import {ChatRoom} from '../../domain/chat-room';

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, AfterViewChecked {

  @Input() className = '';
  @Input() room: ChatRoom;
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();
  messageInput = new FormControl('', []);

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
      this.room.chatLog.push(new ChatMessage(this.room.id, this.room.me.id, this.room.newMessage, new Date()));
      this.room.newMessage = '';
      this.scrollToBottom();
      this.room.receiveMessageFromRandomParticipantIn(Math.floor((Math.random() * 10) + 1));
    }
  }

  getUser(accountId: string): User {
    return this.room.accounts.get(accountId);
  }

  getAvatar(accountId: string): string {
    const account = this.getUser(accountId);
    return account ? account.avatarUrl : null;
  }

  isHeadMessage(chatMessage: ChatMessage, index: number) {
    return (this.room.chatLog.length === index + 1) // last message
      || this.room.chatLog[index + 1].from !== this.room.chatLog[index].from;
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
