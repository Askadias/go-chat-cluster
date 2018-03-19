import {Component, OnInit} from '@angular/core';
import {environment as env} from '../../environments/environment';
import {ActivatedRoute, Router} from "@angular/router";
import {ChatService} from "../services/chat.service";
import {User} from "../common/user";
import {AuthService} from "../services/auth.service";

@Component({
  selector: 'chat-component',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {

  protected oauthConfig: any;
  errors: string[] = [];
  friends: User[] = [];
  loading = false;
  foldSocialBar = false;

  constructor(public route: ActivatedRoute,
              private auth: AuthService,
              private chat: ChatService) {
    this.oauthConfig = env.oauth;
    this.loading = true;
    this.chat.getFriends().subscribe(
      (friends) => {
        this.loading = false;
        this.friends = friends;
      }, (response) => {
        this.loading = false;
        this.errors = [response.message];
      }
    );
  }

  ngOnInit() {

  }

  logout() {
    this.auth.logout()
  }
}
