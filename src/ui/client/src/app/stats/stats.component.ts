import { KeyValue } from "@angular/common"
import { ChangeDetectionStrategy, Component } from "@angular/core"
import { ActivatedRoute } from "@angular/router"
import { ModronStore } from "../state/modron.store"
import { StatsService } from "../stats.service"

import * as pb from "src/proto/modron_pb"

@Component({
  selector: "app-stats",
  templateUrl: "./stats.component.html",
  styleUrls: ["./stats.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class StatsComponent {
  constructor(public store: ModronStore, public stats: StatsService, private route: ActivatedRoute) { }

  displaySearchRules: Map<string, boolean> = new Map<string, boolean>();

  toggleSearch(rule: string): void {
    if (this.displaySearchRules.has(rule)) {
      this.displaySearchRules.set(
        rule,
        !(this.displaySearchRules.get(rule) as boolean)
      )
    } else {
      this.displaySearchRules.set(rule, true)
    }
  }

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

  exportCsvMap(data: Map<string, number>, filename: string) {
    const csvData = Array.from(
      data,
      ([k, v]) => `${k.replace(/,/g, "")},${v}`
    ).reduce((prev, curr) => `${prev}\n${curr}`)
    this, this.exportCsv(csvData, filename)
  }

  exportCsvObs(obs: pb.Observation[], filename: string) {
    const header = "resource-name,resource-group,observed-value,scan-date\n"
    const data = obs.map(
      (v) =>
        `${v.getResource()?.getName()},${v
          .getResource()
          ?.getResourceGroupName().replace("projects/", "")},${v.getObservedValue()},'${v
            .getTimestamp()
            ?.toDate()
            .toUTCString()}'`
    )
    this,
      this.exportCsv(
        header + data.reduce((prev, curr) => `${prev}\n${curr}`),
        filename
      )
  }

  exportCsv(data: string, name: string): void {
    const blob = new Blob([data], { type: "text/csv" })
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
