import {AfterViewChecked, Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormControl, Validators} from '@angular/forms';
import {ChatMessage} from '../../domain/chat-message';
import {User} from '../../domain/user';
import {Chat} from '../../domain/chat';

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, AfterViewChecked {

  @Input() className = '';
  @Input() chat: Chat;
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() dismiss: EventEmitter<any> = new EventEmitter<any>();
  messageInput = new FormControl('', [Validators.required]);

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
    if (this.messageInput.valid) {
      this.chat.chatLog.push(new ChatMessage(Date.now(), this.chat.newMessage, this.chat.me.id));
      this.chat.newMessage = '';
      this.scrollToBottom();
      this.chat.receiveMessageFromRandomParticipantIn(Math.floor((Math.random() * 10) + 1));
    }
  }

  getUser(accountId: string): User {
    return this.chat.accounts.get(accountId);
  }

  getAvatar(accountId: string): string {
    const account = this.getUser(accountId);
    return account ? account.avatarUrl : null;
  }

  isHeadMessage(chatMessage: ChatMessage, index: number) {
    return (this.chat.chatLog.length === index + 1) // last message
      || this.chat.chatLog[index + 1].from !== this.chat.chatLog[index].from;
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
