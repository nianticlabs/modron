import { AppComponent } from "./app.component"
import { AppRoutingModule } from "./app-routing.module"
import { AuthenticationService } from "./authentication.service"
import { AuthenticationStore } from "./state/authentication.store"
import { BrowserAnimationsModule } from "@angular/platform-browser/animations"
import { BrowserModule } from "@angular/platform-browser"
import { CookieService } from "ngx-cookie-service"
import { FilterKeyValuePipe } from "./filter.pipe"
import { FilterNamePipe, MapByTypePipe, MapByObservedValuesPipe, mapFlatRulesPipe } from "./resource-group-details/resource-group-details.pipe"
import { FilterNoObservationsPipe, FilterObsPipe, reverseSortPipe } from "./filter.pipe"
import { FormsModule, ReactiveFormsModule } from "@angular/forms"
import { HistogramHorizontalComponent } from "./histogram-horizontal/histogram-horizontal.component"
import { MarkdownModule } from "ngx-markdown"
import { MatButtonModule } from "@angular/material/button"
import { MatCardModule } from "@angular/material/card"
import { MatDatepickerModule } from "@angular/material/datepicker"
import { MatDialogModule } from "@angular/material/dialog"
import { MatFormFieldModule } from "@angular/material/form-field"
import { MatIconModule } from "@angular/material/icon"
import { MatInputModule } from "@angular/material/input"
import { MatNativeDateModule } from "@angular/material/core"
import { MatProgressBarModule } from "@angular/material/progress-bar"
import { MatSnackBarModule } from "@angular/material/snack-bar"
import { MatTableModule } from "@angular/material/table"
import { ModronAppComponent } from "./modron-app/modron-app.component"
import { ModronService } from "./modron.service"
import { ModronStore } from "./state/modron.store"
import { NgModule } from "@angular/core"
import { NotificationExceptionFormComponent } from "./notification-exception-form/notification-exception-form.component"
import { NotificationExceptionsComponent } from "./notification-exceptions/notification-exceptions.component"
import { NotificationExceptionsFilterPipe } from "./notification-exceptions/notification-exceptions.pipe"
import { NotificationService } from "./notification.service"
import { NotificationStore } from "./state/notification.store"
import { ObservationDetailsComponent } from "./observation-details/observation-details.component"
import { ObservationsPipe, MapPerTypeName, ResourceGroupsPipe, InvalidProjectNb, ObsNbPipe } from "./resource-groups/resource-groups.pipe"
import { ResourceGroupComponent } from "./resource-group/resource-group.component"
import { ResourceGroupDetailsComponent } from "./resource-group-details/resource-group-details.component"
import { ResourceGroupsComponent } from "./resource-groups/resource-groups.component"
import { SearchObsComponent } from "./search-obs/search-obs.component"
import { StatsComponent } from "./stats/stats.component"

@NgModule({
  declarations: [
    AppComponent,
    FilterKeyValuePipe,
    FilterNamePipe,
    FilterNoObservationsPipe,
    FilterObsPipe,
    HistogramHorizontalComponent,
    InvalidProjectNb,
    MapByObservedValuesPipe,
    MapByTypePipe,
    mapFlatRulesPipe,
    MapPerTypeName,
    ModronAppComponent,
    NotificationExceptionFormComponent,
    NotificationExceptionsComponent,
    NotificationExceptionsFilterPipe,
    ObservationDetailsComponent,
    ObservationsPipe,
    ObsNbPipe,
    ResourceGroupComponent,
    ResourceGroupDetailsComponent,
    ResourceGroupsComponent,
    ResourceGroupsPipe,
    reverseSortPipe,
    SearchObsComponent,
    StatsComponent,
  ],
  imports: [
    AppRoutingModule,
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    MarkdownModule.forRoot(),
    MatButtonModule,
    MatCardModule,
    MatDatepickerModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatNativeDateModule,
    MatProgressBarModule,
    MatSnackBarModule,
    MatTableModule,
    ReactiveFormsModule,
  ],
  providers: [
    AuthenticationService,
    AuthenticationStore,
    CookieService,
    ModronService,
    ModronStore,
    NotificationService,
    NotificationStore,
  ],
  bootstrap: [AppComponent],
})
export class AppModule {
}
