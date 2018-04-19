import {Component, Input, ViewEncapsulation} from '@angular/core';
import {fadeAnimation} from "../../animations/fade.animation";
import {User} from "../../domain/user";
import {scaleAnimation} from "../../animations/scale.animation";

@Component({
  selector: 'chat-person',
  templateUrl: './person.component.html',
  styleUrls: ['./person.component.scss'],
  encapsulation: ViewEncapsulation.None,
  animations: [scaleAnimation, fadeAnimation]
})
export class PersonComponent {
  @Input() className = '';
  @Input() person: User;
  @Input() active = false;
  @Input() compact = false;
}
