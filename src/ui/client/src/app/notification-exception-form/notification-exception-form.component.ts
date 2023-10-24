import { NotificationStore } from "../state/notification.store"
import { NotificationException } from "../model/notification.model"

import { Component, Inject } from "@angular/core"
import {
  FormControl,
  Validators,
  FormBuilder,
  FormGroup,
} from "@angular/forms"
import { AuthenticationStore } from "../state/authentication.store"
import { MatDialogRef, MAT_DIALOG_DATA } from "@angular/material/dialog"

@Component({
  selector: "app-notification-exception-form",
  templateUrl: "./notification-exception-form.component.html",
  styleUrls: ["./notification-exception-form.component.scss"],
})
export class NotificationExceptionFormComponent {
  submitting = false;
  sourceSystemFormControl = new FormControl("modron", [Validators.required]);
  justificationFormControl = new FormControl("", [Validators.required]);
  validUntilTimeFormControl = new FormControl(new Date(), [
    Validators.required,
  ]);
  notificationNameFormControl: FormControl
  emailFormControl: FormControl
  formGroup: FormGroup

  constructor(
    @Inject(MAT_DIALOG_DATA) notificationName: string,
    auth: AuthenticationStore,
    private _dialogRef: MatDialogRef<NotificationExceptionFormComponent>,
    private _notification: NotificationStore
  ) {
    this.emailFormControl = new FormControl(auth.user.email, [
      Validators.required,
      Validators.email,
    ])
    this.notificationNameFormControl = new FormControl(notificationName, [
      Validators.required,
    ])
    this.notificationNameFormControl.disable()
    this.sourceSystemFormControl.disable()
    this.emailFormControl.disable()
    this.formGroup = new FormBuilder().group({
      sourceSystem: this.sourceSystemFormControl,
      email: this.emailFormControl,
      notificationName: this.notificationNameFormControl,
      justification: this.justificationFormControl,
      validUntilTime: this.validUntilTimeFormControl,
    })
  }

  onSubmit() {
    const exception = new NotificationException()
    exception.sourceSystem = this.sourceSystemFormControl.value ?? ""
    exception.userEmail = this.emailFormControl.value ?? ""
    exception.notificationName = this.notificationNameFormControl.value ?? ""
    exception.justification = this.justificationFormControl.value ?? ""
    exception.validUntilTime =
      this.validUntilTimeFormControl.value ?? undefined
    this.submitting = true
    this._notification.createException$(exception).subscribe({
      next: (exp) => {
        this.submitting = false
        this._dialogRef.close(exp)
      },
      error: (e) => {
        this.submitting = false
        this._dialogRef.close(e)
      },
    })
  }
}
