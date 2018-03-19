import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'chat-sidebar',
  template: `
    <div class="sidebar {{className}}" [class.fold]="fold">
      <ng-content></ng-content>
    </div>`,
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
