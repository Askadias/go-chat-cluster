import {Component, Input, OnInit, ViewEncapsulation} from '@angular/core';

@Component({
  selector: 'chat-sidebar',
  template: `
    <div class="sidebar {{className}}" [class.fold]="fold">
      <ng-content></ng-content>
    </div>`,
  encapsulation: ViewEncapsulation.None,
  styleUrls: ['./sidebar.component.scss']
})
export class SidebarComponent implements OnInit {

  @Input() fold = false;
  @Input() className = '';

  constructor() {
  }

  ngOnInit() {
  }

}
