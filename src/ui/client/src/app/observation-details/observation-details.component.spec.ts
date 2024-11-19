import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatDialogModule } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NotificationExceptionsFilterPipe } from "../notification-exceptions/notification-exceptions.pipe";
import { AuthenticationStore } from "../state/authentication.store";
import { NotificationStore } from "../state/notification.store";

import { ObservationDetailsComponent } from "./observation-details.component";
import { ParseExternalIdPipe, ShortenDescriptionPipe } from "../filter.pipe";
import {ImpactNamePipe, SeverityNamePipe} from "../severity-indicator/severity-indicator.pipe";

describe("ObservationDetailsComponent", () => {
  let component: ObservationDetailsComponent;
  let fixture: ComponentFixture<ObservationDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MatDialogModule, MatSnackBarModule],
      declarations: [
        ObservationDetailsComponent,
        NotificationExceptionsFilterPipe,
        ImpactNamePipe,
        ParseExternalIdPipe,
        SeverityNamePipe,
        ShortenDescriptionPipe,
      ],
      providers: [AuthenticationStore, NotificationStore],
    }).compileComponents();

    fixture = TestBed.createComponent(ObservationDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
