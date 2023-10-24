import { ChangeDetectionStrategy, Component, Input } from "@angular/core"
import { MatDialog } from "@angular/material/dialog"
import { MatSnackBar } from "@angular/material/snack-bar"
import { Router } from "@angular/router"
import { Observation } from "src/proto/modron_pb"
import { NotificationException } from "../model/notification.model"
import { NotificationExceptionFormComponent } from "../notification-exception-form/notification-exception-form.component"
import { NotificationExceptionsFilterPipe } from "../notification-exceptions/notification-exceptions.pipe"
import { NotificationStore } from "../state/notification.store"

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: "app-observation-details",
  templateUrl: "./observation-details.component.html",
  styleUrls: ["./observation-details.component.scss"],
})
export class ObservationDetailsComponent {
  private static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  private readonly BASE_GCP_URL = "https://console.cloud.google.com"
  readonly FOLDER_URL = `${this.BASE_GCP_URL}/welcome?folder=`
  readonly ORGANIZATION_URL = `${this.BASE_GCP_URL}/welcome?organizationId=`
  readonly PROJECT_URL = `${this.BASE_GCP_URL}/home/dashboard?project=`

  @Input() ob: Observation = new Observation();

  public notifications: Map<string, boolean> = new Map<string, boolean>();

  constructor(
    public notification: NotificationStore,
    private _dialog: MatDialog,
    private _snackBar: MatSnackBar,
    private _router: Router
  ) { }

  display: Map<string, boolean> = new Map<string, boolean>();

  toggle(name: string) {
    if (this.display.has(name)) {
      this.display.set(name, !(this.display.get(name) as boolean))
    } else {
      this.display.set(name, true)
    }
  }

  getObservedValue(ob: Observation): string | undefined {
    return ob.getObservedValue()?.toString()?.replace(/,/g, "")
  }

  getExpectedValue(ob: Observation): string | undefined {
    return ob.getExpectedValue()?.toString()?.replace(/,/g, "")
  }

  parseName(ob: string | undefined): string | undefined {
    if (!(ob?.includes("[") && ob?.includes("]"))) {
      return ob
    }
    return ob?.replace(/(\[.*\]$)/g, "")
  }

  notifyToggle(ob: Observation): void {
    const expName = this.exceptionNameFromObservation(ob)
    if (
      new NotificationExceptionsFilterPipe().transform(
        this.notification.exceptions,
        expName
      ).length == 0
    ) {
      const dialogRef = this._dialog.open(NotificationExceptionFormComponent, {
        data: expName,
      })
      dialogRef
        .afterClosed()
        .subscribe((ret: NotificationException | Error) => {
          const isNotificationException = (
            ret: NotificationException | Error
          ): ret is NotificationException => {
            return ret !== undefined
          }
          if (isNotificationException(ret)) {
            this._snackBar.open(
              "Notification exception created successfully",
              "",
              {
                duration:
                  ObservationDetailsComponent.SNACKBAR_LINGER_DURATION_MS,
              }
            )
          } else {
            this._snackBar.open("Creating notification exception failed", "", {
              duration: ObservationDetailsComponent.SNACKBAR_LINGER_DURATION_MS,
            })
          }
        })
    } else {
      this._router.navigate(["modron", "exceptions", expName])
    }
  }

  exceptionNameFromObservation(ob: Observation): string {
    const resource = ob.getResource()
    return `${resource?.getResourceGroupName().replace(new RegExp("/"), "_")}-${resource?.getName()}-${ob.getName()}`
  }
}
