import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ActivatedRoute } from "@angular/router";
import { AuthenticationStore } from "../state/authentication.store";
import { NotificationStore } from "../state/notification.store";

import { NotificationExceptionsComponent } from "./notification-exceptions.component";
import { NotificationExceptionsFilterPipe } from "./notification-exceptions.pipe";

describe("NotificationExceptionsComponent", () => {
  let component: NotificationExceptionsComponent;
  let fixture: ComponentFixture<NotificationExceptionsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [
        NotificationExceptionsComponent,
        NotificationExceptionsFilterPipe,
      ],
      providers: [
        NotificationStore,
        AuthenticationStore,
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get(): string {
                  return "mock-notification-name";
                },
              },
            },
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(NotificationExceptionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
