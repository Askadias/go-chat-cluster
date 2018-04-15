import {Component, EventEmitter, Input, Output} from '@angular/core';
import {RoomContainer} from "../../domain/room-container";
import {User} from "../../domain/user";

@Component({
  selector: 'chat-group',
  templateUrl: './group.component.html',
  styleUrls: ['./group.component.scss']
})
export class GroupComponent {
  @Input() className = '';
  @Input() roomContainer: RoomContainer;
  @Input() active = false;
  @Input() profile: User;
  @Output() click: EventEmitter<any> = new EventEmitter<any>();
}
