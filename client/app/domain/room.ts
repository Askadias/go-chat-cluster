import {User} from './user';

export class Room {
  id: string;
  owner: string;
  members: string[] = [];

  constructor(ownerId: string, userId: string) {
    this.owner = ownerId;
    this.members = [ownerId, userId];
  }

  addMember(newMember: User) {
    if (!this.members.find((it) => it === newMember.id)) {
      this.members.push(newMember.id);
    }
  }

  kickMember(member: User) {
    this.members.splice(
      this.members.findIndex((it) => it === member.id),
      1
    );
  }
}
