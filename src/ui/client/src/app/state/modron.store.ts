import { Injectable } from "@angular/core"
import { BehaviorSubject, map, Observable } from "rxjs"
import { ModronService } from "../modron.service"
import { StatusInfo } from "../model/modron.model"

import * as pb from "src/proto/modron_pb"

@Injectable()
export class ModronStore {
  public static readonly STATUS_CHECK_INTERVAL_MS = 10000
  private static statusCheckerIsRunning = false

  private _observations: BehaviorSubject<
    Map<string, Map<string, pb.Observation[]>>
  >
  private _runningScans: Map<string, StatusInfo>
  private _scanIdsStatus: BehaviorSubject<Map<string, StatusInfo>>

  constructor(private _service: ModronService) {
    this._observations = new BehaviorSubject<
      Map<string, Map<string, pb.Observation[]>>
    >(new Map())
    this._scanIdsStatus = new BehaviorSubject<Map<string, StatusInfo>>(
      new Map<string, StatusInfo>()
    )
    this._runningScans = new Map<string, StatusInfo>()
    this.fetchInitialData()
  }

  get observations$(): Observable<Map<string, Map<string, pb.Observation[]>>> {
    return new Observable((sub) => this._observations.subscribe(sub))
  }

  get observations(): Map<string, Map<string, pb.Observation[]>> {
    return this._observations.value
  }

  get scanInfo$(): Observable<Map<string, StatusInfo>> {
    return new Observable((sub) => this._scanIdsStatus.subscribe(sub))
  }

  get scanInfo(): Map<string, StatusInfo> {
    return this._scanIdsStatus.value
  }

  fetchObservations(resourceGroups: string[]): Observable<Map<string, Map<string, pb.Observation[]>>> {
    return this._service.listObservations(resourceGroups)
  }

  collectAndScan$(resourceGroups: string[]): Observable<pb.CollectAndScanResponse> {
    this.checkScansStatus()
    return this._service.collectAndScan(resourceGroups).pipe(
      map((res) => {
        // A shallow copy here is enough
        const scanInfo = new Map(this.scanInfo)
        scanInfo.set(res.getCollectId() + ModronService.SEPARATOR + res.getScanId(), {
          state: 2,
          resourceGroups: resourceGroups,
        })
        this._runningScans.set(res.getCollectId() + ModronService.SEPARATOR + res.getScanId(), {
          state: 2,
          resourceGroups: resourceGroups,
        })
        this._scanIdsStatus.next(scanInfo)
        return res
      })
    )
  }

  checkScansStatus() {
    if (ModronStore.statusCheckerIsRunning) {
      return
    }
    ModronStore.statusCheckerIsRunning = true
    const checker = setInterval(() => {
      if (this._runningScans.size === 0) {
        clearInterval(checker)
        ModronStore.statusCheckerIsRunning = false
      }
      this._runningScans.forEach((k, v) => {
        this._service.getCollectAndScanStatus(v).subscribe(
          (res) => {
            let s = res.getCollectStatus()
            if (s === pb.RequestStatus.DONE) {
              s = res.getScanStatus()
            }
            const scanInfo = new Map(this.scanInfo)
            scanInfo.set(v, { state: s, resourceGroups: k.resourceGroups })
            if (s === pb.RequestStatus.DONE) {
              this._runningScans.delete(v)
              this.fetchObservations(k.resourceGroups).subscribe((obs) => {
                const allObs = new Map(this._observations.value)
                obs.forEach((rgObs, rg) => {
                  allObs.set(rg, rgObs)
                })
                this._observations.next(allObs)
              })
            }
            this._scanIdsStatus.next(scanInfo)
          })
      })
    }, ModronStore.STATUS_CHECK_INTERVAL_MS)
  }

  private fetchInitialData() {
    this.fetchObservations([]).subscribe((obs) => {
      this._observations.next(obs)
    })
  }
}
