import { Component, ChangeDetectionStrategy, OnInit } from "@angular/core"
import { ModronStore } from "../state/modron.store"
import { MatSnackBar } from "@angular/material/snack-bar"
import { Observation } from "src/proto/modron_pb"
import { ActivatedRoute, Params, Router } from "@angular/router"
import { map, Observable } from "rxjs"

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: "app-resource-groups",
  templateUrl: "./resource-groups.component.html",
  styleUrls: ["./resource-groups.component.scss"],
})
export class ResourceGroupsComponent implements OnInit {
  private static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  searchText = "";
  removeNoObs = false;

  constructor(
    public store: ModronStore,
    public snackBar: MatSnackBar,
    private activatedRoute: ActivatedRoute,
    private router: Router
  ) { }

  ngOnInit() {
    let filterText = ""
    if (this.activatedRoute.snapshot.queryParamMap.get("filter")?.length != 0) {
      filterText = this.activatedRoute.snapshot.queryParamMap.get("filter")!
    }
    this.searchText = filterText
  }

  getColor(balance: number): string {
    return balance > 0 ? "#da1e28" : "#24a148"
  }

  collectAndScanAll(): void {
    this.store.collectAndScanAll$().subscribe({
      next: () =>
        this.snackBar.open("Scanning all resource groups ...", "", {
          duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS,
        }),
      error: () =>
        this.snackBar.open(
          "An unexpected error has occurred while starting the scan",
          "",
          { duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS }
        ),
    })
  }

  isScanRunning$(): Observable<boolean> {
    return this.store.scanInfo$.pipe(
      map((info) => {
        let running = false
        for (const v of info.values()) {
          if (v.state !== 1) {
            if (v.resourceGroups.length == 0) {
              running = true
            }
          }
        }
        return running
      })
    )
  }

  getDate(obs: Observation[]): Date | null {
    if (obs.length > 0) {
      return obs[0].getTimestamp()?.toDate() || null;
    }
    return null
  }

  public updateFilterUrlParam() {
    const queryParams: Params = { filter: this.searchText }

    this.router.navigate(
      [],
      {
        relativeTo: this.activatedRoute,
        queryParams: queryParams, replaceUrl: true,
        queryParamsHandling: "merge", // remove to replace all query params by provided
      })
  }
}
