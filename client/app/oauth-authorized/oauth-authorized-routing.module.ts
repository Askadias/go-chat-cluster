import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {OauthAuthorizedComponent} from './oauth-authorized.component';

const routes: Routes = [
  {path: '', component: OauthAuthorizedComponent}
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class OauthCallbackRoutingModule {
}
