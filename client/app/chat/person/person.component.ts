import {Component, EventEmitter, Input, Output, ViewEncapsulation} from '@angular/core';
import {User} from "../../domain/user";

@Component({
  selector: 'chat-person',
  templateUrl: './person.component.html',
  styleUrls: ['./person.component.scss'],
  encapsulation: ViewEncapsulation.None
})
export class PersonComponent {
  @Input() className = '';
  @Input() person: User;
  @Input() active = false;
  @Output() click: EventEmitter<any> = new EventEmitter<any>();
}
