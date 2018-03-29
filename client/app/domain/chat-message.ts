export class ChatMessage {
  roomId: string;
  from: string;
  timestamp: Date;
  body: string;
  type: string;

  constructor(roomId: string, from: string, body: string, timestamp: Date, type?: string) {
    this.roomId = roomId;
    this.from = from;
    this.body = body;
    this.timestamp = timestamp;
    this.type = type || 'msg';
  }
}
