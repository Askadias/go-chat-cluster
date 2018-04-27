import {Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {Message} from '../../domain/message';
import {User} from '../../domain/user';
import {RoomContainer} from "../../domain/room-container";
import {environment as env} from "../../../environments/environment";
import {EmojiEvent} from "@ctrl/ngx-emoji-mart/ngx-emoji";
import {ChatService} from "../../services/chat.service";
import {scaleAnimation} from "../../animations/scale.animation";
import TurndownService from 'turndown'
import {gfm} from 'turndown-plugin-gfm'
import anchorme from "anchorme";
import {isImageURL} from '../../common/utils';

@Component({
  selector: 'chat-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss'],
  animations: [scaleAnimation]
})
export class RoomComponent implements OnInit {

  @Input() className = '';
  @Input() isMobile = false;
  @Input() isTablet = false;
  @Output() back: EventEmitter<any> = new EventEmitter<any>();
  @Output() close: EventEmitter<any> = new EventEmitter<any>();
  @Output() delete: EventEmitter<any> = new EventEmitter<any>();
  @Output() startChat: EventEmitter<string> = new EventEmitter<string>();

  @ViewChild('chatLogElement') private chatLogContainer: ElementRef;

  _room: RoomContainer;
  members: User[];
  loading = false;
  sending = false;
  errors: string[] = [];
  emojiPickerOpened = false;
  initialized = false;
  showScrollDown = false;
  turndownService = new TurndownService({
    headingStyle: 'atx',
    codeBlockStyle: 'fenced'
  });

  constructor(private chat: ChatService) {
    this.turndownService.use(gfm);
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
            }, 100);
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

  scrollDown() {
    setTimeout(() => {
      const native = this.chatLogContainer.nativeElement;
      native.scrollTop = native.scrollHeight;
    });
  }

  ngOnInit() {
  }

  sendMessage() {
    let msg = this._room.newMessage;
    if (msg !== '') {
      this.sending = true;
      this.errors = [];
      anchorme(msg, {list: true}).forEach(url => {
        if (url && isImageURL(url.raw)) {
          msg += `\n\r![](${url.raw})`
        }
      });
      this._room.chat.send(new Message(this._room.room.id, this._room.me.id, msg, Date.now()))
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

  onScroll() {
    const native = this.chatLogContainer.nativeElement;
    if (this.initialized && !this.loading) {

      if (native.scrollTop < 1) {
        this.loadMessages()
      }
    }

    this.showScrollDown = native.scrollHeight - (native.scrollTop + native.clientHeight) > 300;
  }

  toggleEmojiPicker(): void {
    this.emojiPickerOpened = !this.emojiPickerOpened;
  }

  pickEmoji(event: EmojiEvent): void {
    if (event) {
      this._room.newMessage += event.emoji.native
    }
  }

  sendOnEnter(event) {
    if (event.keyCode == 13 && !event.shiftKey) {
      this.sendMessage();
      return false;
    }
  }

  trackByMessageId(index: number, message: Message): string {
    return message.id;
  }

  isOwner() {
    return this._room.amIOwner()
  }

  onCopy(e: ClipboardEvent) {
    e.clipboardData.setData('text/plain', this.turndownService.turndown(this.getHTMLOfSelection()));
    e.preventDefault();
  }

  getHTMLOfSelection() {
    let range;
    if (window.getSelection) {
      let selection = window.getSelection();
      if (selection.rangeCount > 0) {
        range = selection.getRangeAt(0);
        let clonedSelection = range.cloneContents();
        let div = document.createElement('div');
        div.appendChild(clonedSelection);
        return div.innerHTML;
      }
      else {
        return '';
      }
    }
    else {
      return '';
    }
  }
}
