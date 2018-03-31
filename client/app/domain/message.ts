export class Message {
  room: string;
  from: string;
  timestamp: number;
  body: string;
  type: string;

  constructor(roomId: string, from: string, body: string, timestamp: number, type?: string) {
    this.room = roomId;
    this.from = from;
    this.body = body;
    this.timestamp = timestamp;
    this.type = type || 'message';
  }
}
