import {Component, EventEmitter, Input, Output, ViewEncapsulation} from '@angular/core';
import {fadeShiftAnimation} from '../../animations/fade-shift.animation';
import {scaleAnimation} from "../../animations/scale.animation";
import {fadeAnimation} from "../../animations/fade.animation";
import {User} from "../../domain/user";

@Component({
  selector: 'chat-person',
  templateUrl: './person.component.html',
  styleUrls: ['./person.component.scss'],
  encapsulation: ViewEncapsulation.None,
  animations: [fadeShiftAnimation, scaleAnimation, fadeAnimation]
})
export class PersonComponent {
  @Input() className = '';
  @Input() person: User;
  @Input() active = false;
  @Input() compact = false;
  @Output() click: EventEmitter<any> = new EventEmitter<any>();
}
