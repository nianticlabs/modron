import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ModronAppComponent } from './modron-app/modron-app.component';
import { NotificationExceptionFormComponent } from './notification-exception-form/notification-exception-form.component';
import { NotificationExceptionsComponent } from './notification-exceptions/notification-exceptions.component';
import { ResourceGroupDetailsComponent } from './resource-group-details/resource-group-details.component';
import { ResourceGroupsComponent } from './resource-groups/resource-groups.component';
import { StatsComponent } from './stats/stats.component';

const routes: Routes = [
  {
    path: 'modron',
    component: ModronAppComponent,
    children: [
      { path: 'projects', component: ResourceGroupsComponent },
      { path: 'project/:id', component: ResourceGroupDetailsComponent },
      { path: 'stats', component: StatsComponent },
      { path: 'exceptions', component: NotificationExceptionsComponent },
      {
        path: 'exceptions/:notificationName',
        component: NotificationExceptionsComponent,
      },
      {
        path: 'exceptions/new/:notificationName',
        component: NotificationExceptionFormComponent,
      },
    ],
  },

  // otherwise redirect to home
  { path: '**', redirectTo: 'modron/projects' },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
