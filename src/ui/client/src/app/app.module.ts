import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatDialogModule } from '@angular/material/dialog';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CookieService } from 'ngx-cookie-service';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ResourceGroupComponent } from './resource-group/resource-group.component';
import { ModronAppComponent } from './modron-app/modron-app.component';
import { AuthenticationService } from './authentication.service';
import { ModronService } from './modron.service';
import { ResourceGroupsComponent } from './resource-groups/resource-groups.component';
import { StatsComponent } from './stats/stats.component';
import {
  FilterNoObservationsPipe,
  FilterObsPipe,
  reverseSortPipe,
} from './filter.pipe';
import { ObservationDetailsComponent } from './observation-details/observation-details.component';
import { SearchObsComponent } from './search-obs/search-obs.component';
import { FilterKeyValuePipe } from './filter.pipe';
import { ResourceGroupDetailsComponent } from './resource-group-details/resource-group-details.component';
import { HistogramHorizontalComponent } from './histogram-horizontal/histogram-horizontal.component';
import { ModronStore } from './state/modron.store';
import {
  FilterNamePipe,
  MapByTypePipe,
  MapByObservedValuesPipe,
  mapFlatRulesPipe,
} from './resource-group-details/resource-group-details.pipe';
import {
  ObservationsPipe,
  MapPerTypeName,
  ResourceGroupsPipe,
  InvalidProjectNb,
  ObsNbPipe,
} from './resource-groups/resource-groups.pipe';
import { NotificationService } from './notification.service';
import { NotificationStore } from './state/notification.store';
import { NotificationExceptionFormComponent } from './notification-exception-form/notification-exception-form.component';
import { NotificationExceptionsComponent } from './notification-exceptions/notification-exceptions.component';
import { NotificationExceptionsFilterPipe } from './notification-exceptions/notification-exceptions.pipe';
import { AuthenticationStore } from './state/authentication.store';
import { MarkdownModule } from 'ngx-markdown';

@NgModule({
  declarations: [
    AppComponent,
    MapPerTypeName,
    ResourceGroupComponent,
    ModronAppComponent,
    ResourceGroupsComponent,
    StatsComponent,
    FilterObsPipe,
    ObsNbPipe,
    FilterKeyValuePipe,
    FilterNoObservationsPipe,
    MapByTypePipe,
    InvalidProjectNb,
    MapByObservedValuesPipe,
    FilterNamePipe,
    ResourceGroupsPipe,
    ObservationsPipe,
    NotificationExceptionsFilterPipe,
    ResourceGroupDetailsComponent,
    HistogramHorizontalComponent,
    mapFlatRulesPipe,
    reverseSortPipe,
    ObservationDetailsComponent,
    SearchObsComponent,

    NotificationExceptionFormComponent,
    NotificationExceptionsComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    FormsModule,
    MatFormFieldModule,
    MatButtonModule,
    MatCardModule,
    MatInputModule,
    MatDatepickerModule,
    MatProgressBarModule,
    MatSnackBarModule,
    MatTableModule,
    MatNativeDateModule,
    MatIconModule,
    MatDialogModule,
    ReactiveFormsModule,
    MarkdownModule.forRoot(),
  ],
  providers: [
    CookieService,
    AuthenticationService,
    AuthenticationStore,
    ModronService,
    ModronStore,
    NotificationService,
    NotificationStore,
  ],
  bootstrap: [AppComponent],
})
export class AppModule {
  constructor() {}

  ngOnInit() {}
}
