import { KeyValue, ViewportScroller } from "@angular/common"
import { ChangeDetectionStrategy, Component, OnInit } from "@angular/core"
import { ActivatedRoute } from "@angular/router"
import { ModronService } from "../modron.service"
import { ModronStore } from "../state/modron.store"

import * as pb from "src/proto/modron_pb"
import { first } from "rxjs"

@Component({
  selector: "app-resource-group-details",
  templateUrl: "./resource-group-details.component.html",
  styleUrls: ["./resource-group-details.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ResourceGroupDetailsComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    public store: ModronStore,
    public modron: ModronService,
    private viewportScroller: ViewportScroller,
  ) { }

  public resourceGroupName = "";
  private readonly BASE_GCP_URL = "https://console.cloud.google.com"
  readonly FOLDER_URL = `${this.BASE_GCP_URL}/welcome?folder=`
  readonly ORGANIZATION_URL = `${this.BASE_GCP_URL}/welcome?organizationId=`
  readonly PROJECT_URL = `${this.BASE_GCP_URL}/home/dashboard?project=`

  public displayObsDetail: Map<string, boolean> = new Map<string, boolean>();

  ngOnInit(): void {
    this.resourceGroupName = (this.route.snapshot.paramMap.get("id") as string).replace(new RegExp("-"), "/")
  }

  // Wait for https://github.com/angular/angular/issues/30139 to be fixed.
  // The bug prevents us from scrolling to a fragment that is dynamically loaded.
  ngAfterViewInit(): void {
    this.store.observations$.subscribe(() =>
      this.route.fragment.pipe(first()).subscribe(fragment => {
        console.log(fragment)
        this.viewportScroller.scrollToAnchor(fragment!)
      })
    )
  }

  filterName(
    obs: Map<string, pb.Observation[]>
  ): Map<string, pb.Observation[]> {
    const m = new Map()
    m.set(this.resourceGroupName, obs.get(this.resourceGroupName as string))
    return m
  }

  getName(
    obs: Map<string, Map<string, pb.Observation[]>>
  ): Map<string, pb.Observation[]> {
    return obs.get(this.resourceGroupName) as Map<string, pb.Observation[]>
  }

  mapByType(obs: Map<string, pb.Observation[]>): Map<string, pb.Observation[]> {
    const obsByType = new Map<string, pb.Observation[]>()
    for (const ob of [...obs.values()].flat()) {
      if (ob === undefined) {
        continue
      }
      const type = ob.getName()
      if (!obsByType.has(type)) {
        obsByType.set(type, [])
      }
      obsByType.get(type)?.push(ob)
    }
    return obsByType
  }

  identity(index: number, item: pb.Observation): string {
    return item.getUid()
  }

  identityKV(index: number, item: KeyValue<string, pb.Observation[]>): string {
    return item.key
  }

  getObservedValue(ob: pb.Observation): string | undefined {
    return ob.getObservedValue()?.toString()?.replace(/,/g, "")
  }

  getExpectedValue(ob: pb.Observation): string | undefined {
    return ob.getExpectedValue()?.toString()?.replace(/,/g, "")
  }

  toggle(id: string | undefined): void {
    id = id as string
    if (this.displayObsDetail.has(id)) {
      this.displayObsDetail.set(
        id,
        !(this.displayObsDetail.get(id) as boolean)
      )
    } else {
      this.displayObsDetail.set(id, true)
    }
  }
}
