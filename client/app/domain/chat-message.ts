export class ChatMessage {
  from: string;
  timestamp: number;
  message: string;

  constructor(timestamp: number, message: string, from: string) {
    this.timestamp = timestamp;
    this.message = message;
    this.from = from;
  }
}
