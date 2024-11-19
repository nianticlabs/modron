import {
  ChangeDetectionStrategy, ChangeDetectorRef,
  Component, OnDestroy,
  OnInit,
} from "@angular/core"
import { ActivatedRoute } from "@angular/router"
import { ModronService } from "../modron.service"
import { ModronStore } from "../state/modron.store"
import * as pb from "src/proto/modron_pb"
import {Subscription, tap} from "rxjs";

type ObsMap = Map<string, pb.Observation[]>
type RgObsMap = Map<string, ObsMap>

@Component({
  selector: "app-resource-group-details",
  templateUrl: "./resource-group-details.component.html",
  styleUrls: ["./resource-group-details.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ResourceGroupDetailsComponent implements OnInit,OnDestroy {
  public loading = true;
  private subscription: Subscription | undefined;
  public obs: pb.Observation[] = [];
  constructor(
    private route: ActivatedRoute,
    public store: ModronStore,
    public modron: ModronService,
    private cdr: ChangeDetectorRef,
  ){

  }
  public resourceGroupName = "";
  private readonly BASE_GCP_URL = "https://console.cloud.google.com"
  readonly FOLDER_URL = `${this.BASE_GCP_URL}/welcome?folder=`
  readonly ORGANIZATION_URL = `${this.BASE_GCP_URL}/welcome?organizationId=`
  readonly PROJECT_URL = `${this.BASE_GCP_URL}/home/dashboard?project=`


  async ngOnInit(): Promise<void> {
    this.resourceGroupName = (this.route.snapshot.paramMap.get("id") as string).replace(new RegExp("-"), "/")
    this.subscription = this.store.observations$.pipe(
      tap((obs) => {
        if(obs.size > 0) {
          this.loading = false
          this.cdr.markForCheck()
        }
        this.obs = this.getObservations(obs)
      })
    ).subscribe()
  }

  ngOnDestroy(): void {
    this.subscription?.unsubscribe()
  }


  getName(obs: RgObsMap): ObsMap | undefined {
    return obs.get(this.resourceGroupName)
  }

  getObservations(obs?: RgObsMap): pb.Observation[] {
    if(obs === undefined) {
      return []
    }
    const obsMap = this.getName(obs)
    if(obsMap === undefined) {
      return []
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    return Array.from(obsMap).map(([_, v]) => v).flat()
  }
  identity(index: number, item: pb.Observation): string {
    return item.getUid()
  }
  getObservedValue(ob: pb.Observation): string | undefined {
    return ob.getObservedValue()?.toString()?.replace(/,/g, "")
  }

  getExpectedValue(ob: pb.Observation): string | undefined {
    return ob.getExpectedValue()?.toString()?.replace(/,/g, "")
  }

  r(row: unknown): pb.Observation {
    return row as pb.Observation
  }

  resourceRef(row: pb.Observation) {
    return row.getResourceRef()?.getExternalId()
  }
}
