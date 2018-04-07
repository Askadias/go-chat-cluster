import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {ChatRoutingModule} from './chat-routing.module';
import {HttpClientModule, HttpClientXsrfModule} from '@angular/common/http';

import {FlexLayoutModule} from '@angular/flex-layout';

import {MatCardModule} from '@angular/material/card';
import {MatIconModule} from '@angular/material/icon';
import {MatButtonModule} from '@angular/material/button';
import {ChatComponent} from './chat.component';
import {SidebarComponent} from "../common/sidebar/sidebar.component";
import {PersonComponent} from "./person/person.component";
import {RoomComponent} from "./room/room.component";
import {MatFormFieldModule} from "@angular/material/form-field";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {MatInputModule} from "@angular/material/input";
import {FriendsFilterPipe} from "./friends-filter.pipe";
import {SkipOwnFilterPipe} from "./skip-own-filter.pipe";
import {PickerModule} from "@ctrl/ngx-emoji-mart";
import {EmojiModule} from '@ctrl/ngx-emoji-mart/ngx-emoji'
import {NgxWigModule} from 'ngx-wig';


@NgModule({
  declarations: [
    PersonComponent,
    SidebarComponent,
    ChatComponent,
    RoomComponent,
    FriendsFilterPipe,
    SkipOwnFilterPipe
  ],
  imports: [
    CommonModule,
    ChatRoutingModule,
    FlexLayoutModule,
    FormsModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    HttpClientModule,
    HttpClientXsrfModule,
    PickerModule,
    EmojiModule,
    NgxWigModule
  ],
  providers: []
})
export class ChatModule {
}
