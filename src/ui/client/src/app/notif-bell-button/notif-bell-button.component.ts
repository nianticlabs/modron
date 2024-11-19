import {Component, Input} from "@angular/core";
import {Observation} from "../../proto/modron_pb";
import {NotificationStore} from "../state/notification.store";
import {NotificationExceptionsFilterPipe} from "../notification-exceptions/notification-exceptions.pipe";
import {NotificationExceptionFormComponent} from "../notification-exception-form/notification-exception-form.component";
import {NotificationException} from "../model/notification.model";
import {MatSnackBar} from "@angular/material/snack-bar";
import {MatDialog} from "@angular/material/dialog";
import {Router} from "@angular/router";

@Component(
  {
    selector: "app-notif-bell-button",
    templateUrl: "./notif-bell-button.component.html",
    styleUrls: ["./notif-bell-button.component.scss"],
  }
)
export class NotificationBellButtonComponent {
  @Input()
  public observation: Observation|undefined;
  static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  constructor(
    public notification: NotificationStore,
    private _dialog: MatDialog,
    private _snackBar: MatSnackBar,
    private _router: Router,
  ) {
  }

  exceptionNameFromObservation(ob: Observation): string {
    const resource = ob.getResourceRef()
    return `${resource?.getGroupName().replace(new RegExp("/"), "_")}-${resource?.getExternalId()}-${ob.getName()}`
  }

  notifyToggle(ob: Observation): void {
    const expName = this.exceptionNameFromObservation(ob);
    if (
      new NotificationExceptionsFilterPipe().transform(
        this.notification.exceptions,
        expName
      ).length == 0
    ) {
      const dialogRef = this._dialog.open(NotificationExceptionFormComponent, {
        data: expName,
      });
      dialogRef
        .afterClosed()
        .subscribe((ret: NotificationException | Error | boolean) => {
          if(ret === false) {
            return
          }
          if (ret instanceof NotificationException) {
            this._snackBar.open(
              "Notification exception created successfully",
              "",
              {
                duration: NotificationBellButtonComponent.SNACKBAR_LINGER_DURATION_MS,
              }
            );
          } else {
            this._snackBar.open("Creating notification exception failed", "", {
              duration: NotificationBellButtonComponent.SNACKBAR_LINGER_DURATION_MS,
            });
          }
        });
    } else {
      this._router.navigate(["modron", "exceptions", expName]);
    }
  }
}
