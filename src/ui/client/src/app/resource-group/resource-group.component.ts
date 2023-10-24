import { Component, Input } from "@angular/core"
import { map, Observable } from "rxjs"
import { ModronStore } from "../state/modron.store"
import { MatSnackBar } from "@angular/material/snack-bar"
import * as pb from "src/proto/modron_pb"

@Component({
  selector: "app-resource-group",
  templateUrl: "./resource-group.component.html",
  styleUrls: ["./resource-group.component.scss"],
})
export class ResourceGroupComponent {
  private static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  @Input()
  name = "";

  @Input()
  lastScanDate = "";

  @Input()
  provider = "";

  @Input()
  observationCount = -1;

  constructor(public store: ModronStore, public snackBar: MatSnackBar) { }

  collectAndScan(resourceGroups: string[]): void {
    this.store
      .collectAndScan$(resourceGroups)
      .subscribe({
        next: () =>
          this.snackBar.open("Scanning " + resourceGroups.join(",") + " ...", "", {
            duration: ResourceGroupComponent.SNACKBAR_LINGER_DURATION_MS,
          }),
        error: () =>
          this.snackBar.open(
            "An unexpected error has occurred while starting the collection",
            "",
            { duration: ResourceGroupComponent.SNACKBAR_LINGER_DURATION_MS }
          ),
      })
  }

  isScanRunning$(project: string): Observable<boolean> {
    return this.store.scanInfo$.pipe(
      map((info) => {
        for (const v of info.values()) {
          if (v.state === pb.RequestStatus.RUNNING) {
            if (
              v.resourceGroups.includes(project) ||
              v.resourceGroups.length === 0
            ) {
              return true
            }
          }
        }
        return false
      })
    )
  }
}
