import { ChangeDetectionStrategy, Component, Input } from "@angular/core";
import { MatDialog } from "@angular/material/dialog";
import { MatSnackBar } from "@angular/material/snack-bar";
import { Router } from "@angular/router";
import { Observation, Severity } from "../../proto/modron_pb";
import { NotificationException } from "../model/notification.model";
import { NotificationExceptionFormComponent } from "../notification-exception-form/notification-exception-form.component";
import { NotificationExceptionsFilterPipe } from "../notification-exceptions/notification-exceptions.pipe";
import { NotificationStore } from "../state/notification.store";

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: "app-observation-details",
  templateUrl: "./observation-details.component.html",
  styleUrls: ["./observation-details.component.scss"],
})
export class ObservationDetailsComponent {
  readonly Severity = Severity;
  static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  private readonly BASE_GCP_URL = "https://console.cloud.google.com";
  readonly FOLDER_URL = `${this.BASE_GCP_URL}/welcome?folder=`;
  readonly ORGANIZATION_URL = `${this.BASE_GCP_URL}/welcome?organizationId=`;
  readonly PROJECT_URL = `${this.BASE_GCP_URL}/home/dashboard?project=`;

  @Input() ob: Observation = new Observation();

  @Input()
  public expanded: boolean = true;
  @Input()
  public showActions: boolean = true;

  public notifications: Map<string, boolean> = new Map<string, boolean>();

  constructor(
    public notification: NotificationStore,
    private _dialog: MatDialog,
    private _snackBar: MatSnackBar,
    private _router: Router
  ) {}
  display: Map<string, boolean> = new Map<string, boolean>();

  toggle(name: string) {
    if (this.display.has(name)) {
      this.display.set(name, !(this.display.get(name) as boolean));
    } else {
      this.display.set(name, true);
    }
  }

  getColor(severity: number): string {
    switch (severity) {
      case Severity.SEVERITY_CRITICAL:
        return "red";
      case Severity.SEVERITY_HIGH:
        return "orange";
      case Severity.SEVERITY_MEDIUM:
        return "yellow";
      case Severity.SEVERITY_LOW:
        return "green";
      default:
        return "black";
    }
  }

  getSeverity(severity: number): string {
    switch (severity) {
      case Severity.SEVERITY_CRITICAL:
        return "Critical";
      case Severity.SEVERITY_HIGH:
        return "High";
      case Severity.SEVERITY_MEDIUM:
        return "Medium";
      case Severity.SEVERITY_LOW:
        return "Low";
      case Severity.SEVERITY_INFO:
        return "Info";
      default:
        return "Unknown";
    }
  }

  getCategoryName(category: number): string {
    switch (category) {
      case Observation.Category.CATEGORY_VULNERABILITY:
        return "Vulnerability";
      case Observation.Category.CATEGORY_MISCONFIGURATION:
        return "Misconfiguration";
      case Observation.Category.CATEGORY_TOXIC_COMBINATION:
        return "Toxic Combination";
    }
    return "UNKNOWN";
  }

  getRgLink(observation: Observation): string {
    const rgName = this.getRgName(observation);
    if(rgName.startsWith("folders/")) {
      return `${this.FOLDER_URL}${rgName.replace("folders/", "")}`;
    }
    if(rgName.startsWith("organizations/")) {
      return `${this.ORGANIZATION_URL}${rgName.replace("organizations/", "")}`;
    }
    if(rgName.startsWith("projects/")) {
      return `${this.PROJECT_URL}${rgName.replace("projects/", "")}`;
    }
    return "";
  }

  getRgName(observation: Observation): string {
    const resource = observation.getResourceRef();
    if (resource === undefined) {
      return "";
    }
    return resource.getGroupName();
  }

  getObservedValue(ob: Observation): string | undefined {
    return ob.getObservedValue()?.toString()?.replace(/,/g, "");
  }

  getExpectedValue(ob: Observation): string | undefined {
    return ob.getExpectedValue()?.toString()?.replace(/,/g, "");
  }

  parseName(ob: string | undefined): string | undefined {
    if (!(ob?.includes("[") && ob?.includes("]"))) {
      return ob;
    }
    return ob?.replace(/(\[.*]$)/g, "");
  }

  async notifyToggle(ob: Observation): Promise<void> {
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
                duration:
                  ObservationDetailsComponent.SNACKBAR_LINGER_DURATION_MS,
              }
            );
          } else {
            this._snackBar.open("Creating notification exception failed", "", {
              duration: ObservationDetailsComponent.SNACKBAR_LINGER_DURATION_MS,
            });
          }
        });
    } else {
      await this._router.navigate(["modron", "exceptions", expName]);
    }
  }

  exceptionNameFromObservation(ob: Observation): string {
    const resource = ob.getResourceRef()
    return `${resource?.getGroupName().replace(new RegExp("/"), "_")}-${resource?.getExternalId()}-${ob.getName()}`
  }
}
