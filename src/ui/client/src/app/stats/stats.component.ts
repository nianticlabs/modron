import { Component, OnInit, ChangeDetectionStrategy } from '@angular/core'
import { ModronStore } from '../state/modron.store'
import { StatsService } from '../stats.service'
import { KeyValue } from '@angular/common'

import * as pb from 'src/proto/modron_pb'

@Component({
  selector: 'app-stats',
  templateUrl: './stats.component.html',
  styleUrls: ['./stats.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class StatsComponent implements OnInit {
  constructor(public store: ModronStore, public stats: StatsService) { }

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

  ngOnInit(): void {
    // if (this.store.observations.size === 0) {
    //   this.store.fetchObservations([])
    // }
  }

  mapByType(
    obs: Map<string, Map<string, pb.Observation[]>>
  ): Map<string, pb.Observation[]> {
    let obsByType = new Map<string, pb.Observation[]>()
    obs.forEach((v1, k1, m1) => {
      v1.forEach((v2, k2, m2) => {
        if (!obsByType.has(k2)) {
          obsByType.set(k2, [])
        }
        obsByType.set(k2, (obsByType.get(k2) as pb.Observation[]).concat(v2))
      })
    })
    return obsByType
  }

  exportCsvMap(data: Map<string, number>, filename: string) {
    var csvData = Array.from(
      data,
      ([k, v]) => `${k.replace(/,/g, '')},${v}`
    ).reduce((prev, curr) => `${prev}\n${curr}`)
    this, this.exportCsv(csvData, filename)
  }

  exportCsvObs(obs: pb.Observation[], filename: string) {
    let header = 'resource-name,resource-group,observed-value,scan-date\n'
    let data = obs.map(
      (v, i, a) =>
        `${v.getResource()?.getName()},${v
          .getResource()
          ?.getResourceGroupName()},${v.getObservedValue()},'${v
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
    var blob = new Blob([data], { type: 'text/csv' })
    var url = window.URL.createObjectURL(blob)
    var filename = name + '.csv'

    if ((window.navigator as any).msSaveOrOpenBlob) {
      (window.navigator as any).msSaveOrOpenBlob(blob, filename)
    } else {
      var a = document.createElement('a')
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
