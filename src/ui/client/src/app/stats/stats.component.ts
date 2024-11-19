import {KeyValue} from "@angular/common"
import {ChangeDetectionStrategy, Component} from "@angular/core"
import {ActivatedRoute} from "@angular/router"
import {ModronStore} from "../state/modron.store"
import {StatsService} from "../stats.service"

import * as pb from "../../proto/modron_pb"
import {StructValueToStringPipe} from "../filter.pipe";

@Component({
  selector: "app-stats",
  templateUrl: "./stats.component.html",
  styleUrls: ["./stats.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class StatsComponent {
  constructor(public store: ModronStore, public stats: StatsService, private route: ActivatedRoute) {
  }

  displaySearchRules: Map<string, boolean> = new Map<string, boolean>();
  mapByType(
    obs: Map<string, Map<string, pb.Observation[]>>
  ): Map<string, pb.Observation[]> {
    const obsByType = new Map<string, pb.Observation[]>()
    obs.forEach((v1) => {
      v1.forEach((v2, k2) => {
        if (!obsByType.has(k2)) {
          obsByType.set(k2, [])
        }
        obsByType.set(k2, (obsByType.get(k2) as pb.Observation[]).concat(v2))
      })
    })
    return obsByType
  }

  exportCsvObs(obs: pb.Observation[], filename: string) {
    const rows: string[][] = [
      ["resource-name", "resource-group", "expected-value", "observed-value", "scan-date"]
    ]
    const obsRows: string[][] = obs.map(
      (v) => {
        return [
          v.getResourceRef()?.getExternalId(),
          v.getResourceRef()?.getGroupName().replace("projects/", ""),
          StructValueToStringPipe.prototype.transform(v.getExpectedValue()),
          StructValueToStringPipe.prototype.transform(v.getObservedValue()),
          v.getTimestamp()?.toDate().toUTCString()
        ] as string[]
      }
      )
    rows.push(...obsRows)
    obsRows.length = 0
      this.exportCsv(
        rows.map((v)=>v.join(",")).join("\n"),
        filename
      )
  }

  exportCsv(data: string, name: string): void {
    const blob = new Blob([data], {type: "text/csv;charset=utf-8"})
    const url = window.URL.createObjectURL(blob)
    const filename = name + ".csv"

    if ((window.navigator as any).msSaveOrOpenBlob) {
      (window.navigator as any).msSaveOrOpenBlob(blob, filename)
    } else {
      const a = document.createElement("a")
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    }
    window.URL.revokeObjectURL(url)
  }

  identity(index: number, item: KeyValue<string, pb.Observation[]>): string {
    return item.key
  }
}
