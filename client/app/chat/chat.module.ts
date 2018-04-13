import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {ChatRoutingModule} from './chat-routing.module';
import {HttpClientModule, HttpClientXsrfModule} from '@angular/common/http';

import {FlexLayoutModule} from '@angular/flex-layout';

import {MatCardModule} from '@angular/material/card';
import {MatIconModule} from '@angular/material/icon';
import {MatButtonModule} from '@angular/material/button';
import {MatFormFieldModule} from "@angular/material/form-field";
import {MatInputModule} from "@angular/material/input";
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import {MatSidenavModule} from '@angular/material/sidenav';
import {MatToolbarModule} from '@angular/material/toolbar';
import {ChatComponent} from './chat.component';
import {SidebarComponent} from "../common/sidebar/sidebar.component";
import {PersonComponent} from "./person/person.component";
import {RoomComponent} from "./room/room.component";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {FriendsFilterPipe} from "./friends-filter.pipe";
import {SkipOwnFilterPipe} from "./skip-own-filter.pipe";
import {PickerModule} from "@ctrl/ngx-emoji-mart";
import {EmojiModule} from '@ctrl/ngx-emoji-mart/ngx-emoji'
import {NgxWigModule} from 'ngx-wig';
import {NewlinePipe} from "./room/newline.pipe";
import {MediaMatcher} from "@angular/cdk/layout";


@NgModule({
  declarations: [
    PersonComponent,
    SidebarComponent,
    ChatComponent,
    RoomComponent,
    FriendsFilterPipe,
    SkipOwnFilterPipe,
    NewlinePipe
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
    MatProgressSpinnerModule,
    MatSidenavModule,
    MatToolbarModule,
    HttpClientModule,
    HttpClientXsrfModule,
    PickerModule,
    EmojiModule,
    NgxWigModule
  ],
  providers: [
    MediaMatcher
  ]
})
export class ChatModule {
}
