import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {HttpClientModule, HttpClientXsrfModule} from '@angular/common/http';

import {FlexLayoutModule} from '@angular/flex-layout';

import {MatCardModule} from '@angular/material/card';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import {MatButtonModule} from '@angular/material/button';
import {OauthAuthorizedComponent} from './oauth-authorized.component';
import {OauthCallbackRoutingModule} from './oauth-authorized-routing.module';
import {AuthService} from "../services/auth.service";

@NgModule({
  declarations: [
    OauthAuthorizedComponent
  ],
  imports: [
    CommonModule,
    OauthCallbackRoutingModule,
    FlexLayoutModule,
    MatCardModule,
    MatFormFieldModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    HttpClientModule,
    HttpClientXsrfModule
  ],
  providers: [
    AuthService
  ]
})
export class OauthAuthorizedModule {
}
