import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {LoginRoutingModule} from './login-routing.module';
import {HttpClientModule, HttpClientXsrfModule} from '@angular/common/http';

import {FlexLayoutModule} from '@angular/flex-layout';

import {MatCardModule} from '@angular/material/card';
import {MatIconModule} from '@angular/material/icon';
import {MatButtonModule} from '@angular/material/button';
import {LoginComponent} from './login.component';


@NgModule({
  declarations: [
    LoginComponent
  ],
  imports: [
    CommonModule,
    LoginRoutingModule,
    FlexLayoutModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    HttpClientModule,
    HttpClientXsrfModule
  ],
  providers: [
  ]
})
export class LoginModule {
}
