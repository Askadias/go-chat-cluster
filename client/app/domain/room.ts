import {User} from './user';
import {MemberInfo} from "./member-info";

export class Room {
  id: string;
  alias?: string;
  owner?: string;
  members: string[] = [];
  memberInfo?: MemberInfo;

  constructor(currentUserId: string, userId: string) {
    this.members = [currentUserId, userId];
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
