import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {ChatRoutingModule} from './chat-routing.module';
import {HttpClientModule, HttpClientXsrfModule} from '@angular/common/http';

import {FlexLayoutModule} from '@angular/flex-layout';

import {MatCardModule} from '@angular/material/card';
import {MatIconModule} from '@angular/material/icon';
import {MatButtonModule} from '@angular/material/button';
import {ChatComponent} from './chat.component';
import {SidebarComponent} from "./sidebar/sidebar.component";
import {PersonComponent} from "./person/person.component";


@NgModule({
  declarations: [
    PersonComponent,
    SidebarComponent,
    ChatComponent
  ],
  imports: [
    CommonModule,
    ChatRoutingModule,
    FlexLayoutModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    HttpClientModule,
    HttpClientXsrfModule
  ],
  providers: []
})
export class ChatModule {
}
