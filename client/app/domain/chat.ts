import {ChatMessage} from './chat-message';
import {User} from './user';
import {v4 as uuid} from 'uuid';

export class Chat {
  id: string;
  me: User;
  participants: User[] = [];
  chatLog: ChatMessage[];
  newMessage = '';

  accounts: Map<string, User> = new Map<string, User>();

  constructor(me: User, participant: User, chatLog: ChatMessage[] = []) {
    this.me = me;
    if (participant) {
      this.participants = [participant];
    }
    this.chatLog = chatLog;
    this.id = uuid();

    if (participant) {
      this.accounts.set(participant.id, participant);
    }
    this.accounts.set(this.me.id, this.me);
  }

  addParticipant(newParticipant: User) {
    if (!this.participants.find((it) => it.id === newParticipant.id)) {
      this.participants.push(newParticipant);
      this.accounts.set(newParticipant.id, newParticipant);
    }
  }

  excludeParticipant(participant: User) {
    this.participants.splice(
      this.participants.findIndex((it) => it.id === participant.id),
      1
    );
  }

  receiveMessageFromRandomParticipantIn(seconds: number) {
    const count = this.participants.length;
    const actor = this.participants[Math.floor((Math.random() * count))];
    setTimeout(() => this.onMessageReceive('Hi there!', actor.id), seconds * 1000);
  }

  onMessageReceive(message: string, accountId: string) {
    this.chatLog.push(new ChatMessage(Date.now(), message, accountId));
  }
}
