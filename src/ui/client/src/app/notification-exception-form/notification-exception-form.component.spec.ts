import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NotificationExceptionFormComponent } from './notification-exception-form.component';
import { HttpClient } from '@angular/common/http';
import {
  HttpClientTestingModule,
  HttpTestingController,
} from '@angular/common/http/testing';
import { AuthenticationStore } from '../state/authentication.store';
import { NotificationStore } from '../state/notification.store';
import { Validators } from '@angular/forms';
import { NotificationService } from '../notification.service';
import { NotificationException } from '../model/notification.model';
import {
  MatDialogModule,
  MatDialogRef,
  MAT_DIALOG_DATA,
} from '@angular/material/dialog';

describe('NotificationExceptionFormComponent', () => {
  let component: NotificationExceptionFormComponent;
  let fixture: ComponentFixture<NotificationExceptionFormComponent>;

  let httpMock: HttpTestingController;
  let httpClient: HttpClient;

  let service: NotificationService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [NotificationExceptionFormComponent],
      imports: [HttpClientTestingModule, MatDialogModule],
      providers: [
        {
          provide: AuthenticationStore,
          useValue: {
            user: {
              isSignedIn: true,
              email: 'foo@bar.com',
            },
          },
        },
        NotificationStore,
        {
          provide: MatDialogRef,
          useValue: {
            close: () => {},
          },
        },
        {
          provide: MAT_DIALOG_DATA,
          useValue: {
            notificationName: 'mock-notification-name',
          },
        },
      ],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
    httpClient = TestBed.inject(HttpClient);

    service = TestBed.inject(NotificationService);
    fixture = TestBed.createComponent(NotificationExceptionFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('empty form is not valid', () => {
    expect(component.formGroup.valid).toBeFalse();
  });

  it('invalid form cannot be submitted', async () => {
    expect(component.formGroup.valid).toBeFalse();
    spyOn(component, 'onSubmit');

    fixture.debugElement.nativeElement
      .querySelector("button[type='submit']")
      .click();
    fixture.whenStable().then(() => {
      expect(component.onSubmit).toHaveBeenCalledTimes(0);
    });
  });

  it('grpc proto contains all submitted form data', () => {
    const validUntilTime = new Date();
    validUntilTime.setHours(validUntilTime.getHours() + 24);

    component.justificationFormControl.setValue('trust me');
    component.validUntilTimeFormControl.setValue(validUntilTime);
    expect(component.formGroup.valid).toBeTrue();

    const spy = spyOn(service, 'createException$').and.callThrough();
    component.onSubmit();

    let expected = new NotificationException();
    expected.sourceSystem = 'modron';
    expected.userEmail = 'foo@bar.com';
    expected.notificationName = 'mock-notification-name';
    expected.justification = 'trust me';
    expected.validUntilTime = validUntilTime;
    expect(spy).toHaveBeenCalledWith(expected.toProto());
  });

  it('source system is disabled', () => {
    expect(component.sourceSystemFormControl.disabled).toBeTrue();
  });

  it('email is disabled', () => {
    expect(component.emailFormControl.disabled).toBeTrue();
  });

  it('notification name is disabled', () => {
    expect(component.notificationNameFormControl.disabled).toBeTrue();
  });

  it('source system value is correct', () => {
    expect(component.sourceSystemFormControl.value).toBe('modron');
  });

  it('email value is correct', () => {
    expect(component.emailFormControl.value).toBe('foo@bar.com');
  });

  it('source system is required', () => {
    expect(
      component.sourceSystemFormControl.hasValidator(Validators.required)
    ).toBeTrue();
  });

  it('name is required', () => {
    expect(
      component.notificationNameFormControl.hasValidator(Validators.required)
    ).toBeTrue();
  });

  it('email is required', () => {
    expect(
      component.emailFormControl.hasValidator(Validators.required)
    ).toBeTrue();
  });

  it('email is validated', () => {
    expect(
      component.emailFormControl.hasValidator(Validators.email)
    ).toBeTrue();
  });

  it('justification is required', () => {
    expect(
      component.justificationFormControl.hasValidator(Validators.required)
    ).toBeTrue();
  });

  it('expiration date is required', () => {
    expect(
      component.validUntilTimeFormControl.hasValidator(Validators.required)
    ).toBeTrue();
  });
});
