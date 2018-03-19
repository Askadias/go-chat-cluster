import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {fadeShiftAnimation} from '../../animations/fade-shift.animation';

@Component({
  selector: 'chat-person',
  templateUrl: './person.component.html',
  styleUrls: ['./person.component.scss'],
  animations: [fadeShiftAnimation]
})
export class PersonComponent {
  @Input() className = '';
  @Input() avatar = '';
  @Input() name = '';
  @Input() active = false;
  @Input() compact = false;
  @Output() click: EventEmitter<any> = new EventEmitter<any>();
  @Input() showPersonControls = false;
}
