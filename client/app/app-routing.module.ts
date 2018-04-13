import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {AuthGuard} from "./login/auth.guard";
import {AuthenticatedGuard} from "./login/authenticated.guard";

const routes: Routes = [
  {
    path: 'login/:provider', loadChildren: 'app/login/login.module#LoginModule',
    canActivate: [AuthenticatedGuard]
  },
  {
    path: 'authorized', loadChildren: 'app/oauth-authorized/oauth-authorized.module#OauthAuthorizedModule'
  },
  {
    path: 'chat', loadChildren: 'app/chat/chat.module#ChatModule',
    canActivate: [AuthGuard]
  },
  {
    path: '**',
    redirectTo: '/login/'
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule {
}
