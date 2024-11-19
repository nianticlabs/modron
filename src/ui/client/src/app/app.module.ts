import { AppComponent } from "./app.component"
import { AppRoutingModule } from "./app-routing.module"
import { AuthenticationService } from "./authentication.service"
import { AuthenticationStore } from "./state/authentication.store"
import { BrowserAnimationsModule } from "@angular/platform-browser/animations"
import { BrowserModule } from "@angular/platform-browser"
import { CookieService } from "ngx-cookie-service"
import {FilterKeyValuePipe, ParseExternalIdPipe, ShortenDescriptionPipe, StructValueToStringPipe} from "./filter.pipe"
import { FilterNamePipe, MapByTypePipe, MapByObservedValuesPipe, mapFlatRulesPipe } from "./resource-group-details/resource-group-details.pipe"
import { FilterNoObservationsPipe, FilterObsPipe, reverseSortPipe } from "./filter.pipe"
import { FormsModule, ReactiveFormsModule } from "@angular/forms"
import { ObservationsStatsComponent } from "./observations-stats/observations-stats.component"
import { MarkdownModule } from "ngx-markdown"
import { MatButtonModule } from "@angular/material/button"
import { MatCardModule } from "@angular/material/card"
import { MatDatepickerModule } from "@angular/material/datepicker"
import { MatDialogModule } from "@angular/material/dialog"
import { MatExpansionModule } from "@angular/material/expansion"
import { MatFormFieldModule } from "@angular/material/form-field"
import { MatIconModule } from "@angular/material/icon"
import { MatInputModule } from "@angular/material/input"
import { MatListModule } from "@angular/material/list"
import { MatMenuModule } from "@angular/material/menu";
import {MatNativeDateModule, MatRippleModule} from "@angular/material/core"
import { MatProgressBarModule } from "@angular/material/progress-bar"
import { MatSnackBarModule } from "@angular/material/snack-bar"
import { MatSidenavModule } from "@angular/material/sidenav"
import { MatTableModule } from "@angular/material/table"
import { MatToolbarModule } from "@angular/material/toolbar";
import { MatTooltipModule } from "@angular/material/tooltip";
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
import {
  ObservationsPipe,
  MapPerTypeName,
  ResourceGroupsPipe,
  InvalidProjectNb,
  ObsNbPipe,
  MapByRiskScorePipe
} from "./resource-groups/resource-groups.pipe"
import { ResourceGroupComponent } from "./resource-group/resource-group.component"
import { ResourceGroupDetailsComponent } from "./resource-group-details/resource-group-details.component"
import { ResourceGroupsComponent } from "./resource-groups/resource-groups.component"
import { SearchObsComponent } from "./search-obs/search-obs.component"
import { SeverityIndicatorComponent } from "./severity-indicator/severity-indicator.component";
import { StatsComponent } from "./stats/stats.component"
import {SidenavComponent} from "./sidenav/sidenav.component";
import {MatCheckboxModule} from "@angular/material/checkbox";
import {NgOptimizedImage} from "@angular/common";
import {MatBadgeModule} from "@angular/material/badge";
import {FromNowPipe} from "./resource-group/resource-group.pipe";
import {ImpactNamePipe, SeverityAmountPipe, SeverityNamePipe} from "./severity-indicator/severity-indicator.pipe";
import {UIDemoComponent} from "./ui-demo/ui-demo.component";
import {MatSortModule} from "@angular/material/sort";
import {NotificationBellButtonComponent} from "./notif-bell-button/notif-bell-button.component";
import {ObservationDetailsDialogComponent} from "./observation-details-dialog/observation-details-dialog.component";
import {
  ObservationDetailsDialogContentComponent
} from "./observation-details-dialog-content/observation-details-dialog-content.component";
import {ImpactIndicatorComponent} from "./impact-indicator/impact-indicator.component";
import {CategoryNamePipe} from "./observation-details-dialog-content/observation-details-dialog-content.filter";
import {ObservationsTableComponent} from "./observations-table/observations-table.component";
import {BaseChartDirective, provideCharts, withDefaultRegisterables} from "ng2-charts";
import {MatGridList, MatGridTile} from "@angular/material/grid-list";

@NgModule({
  declarations: [
    AppComponent,
    CategoryNamePipe,
    FilterKeyValuePipe,
    FilterNamePipe,
    FilterNoObservationsPipe,
    FilterObsPipe,
    FromNowPipe,
    ObservationsStatsComponent,
    ImpactIndicatorComponent,
    ImpactNamePipe,
    InvalidProjectNb,
    MapByObservedValuesPipe,
    MapByTypePipe,
    mapFlatRulesPipe,
    MapPerTypeName,
    MapByRiskScorePipe,
    ModronAppComponent,
    NotificationBellButtonComponent,
    NotificationExceptionFormComponent,
    NotificationExceptionsComponent,
    NotificationExceptionsFilterPipe,
    ObservationDetailsComponent,
    ObservationDetailsDialogComponent,
    ObservationDetailsDialogContentComponent,
    ObservationsPipe,
    ObservationsTableComponent,
    ObsNbPipe,
    ParseExternalIdPipe,
    ResourceGroupComponent,
    ResourceGroupDetailsComponent,
    ResourceGroupsComponent,
    ResourceGroupsPipe,
    reverseSortPipe,
    SearchObsComponent,
    SeverityIndicatorComponent,
    SeverityAmountPipe,
    SeverityNamePipe,
    ShortenDescriptionPipe,
    StatsComponent,
    StructValueToStringPipe,
    UIDemoComponent,
  ],
  imports: [
    AppRoutingModule,
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    MarkdownModule.forRoot(),
    MatBadgeModule,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatDatepickerModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatListModule,
    MatMenuModule,
    MatNativeDateModule,
    MatProgressBarModule,
    MatSidenavModule,
    MatSnackBarModule,
    MatTableModule,
    MatToolbarModule,
    MatTooltipModule,
    NgOptimizedImage,
    ReactiveFormsModule,
    SidenavComponent,
    MatSortModule,
    MatRippleModule,
    BaseChartDirective,
    MatGridList,
    MatGridTile,
  ],
  providers: [
    AuthenticationService,
    AuthenticationStore,
    CookieService,
    ModronService,
    ModronStore,
    NotificationService,
    NotificationStore,
    provideCharts(withDefaultRegisterables())
  ],
  bootstrap: [AppComponent],
})
export class AppModule {
}
