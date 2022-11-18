import { Injectable } from '@angular/core';
import { BehaviorSubject, map, Observable } from 'rxjs';
import { ModronService } from '../modron.service';
import { StatusInfo } from '../model/modron.model';

import * as pb from 'src/proto/modron_pb';

@Injectable()
export class ModronStore {
  private _observations: BehaviorSubject<
    Map<string, Map<string, pb.Observation[]>>
  >;
  private _scanIdsStatus: BehaviorSubject<Map<string, StatusInfo>>;
  private _collectIdsStatus: BehaviorSubject<Map<string, StatusInfo>>;

  constructor(private _service: ModronService) {
    this._observations = new BehaviorSubject<
      Map<string, Map<string, pb.Observation[]>>
    >(new Map());
    this._scanIdsStatus = new BehaviorSubject<Map<string, StatusInfo>>(
      new Map<string, StatusInfo>()
    );
    this._collectIdsStatus = new BehaviorSubject<Map<string, StatusInfo>>(
      new Map<string, StatusInfo>()
    );
    this.fetchInitialData();
  }

  get observations$(): Observable<Map<string, Map<string, pb.Observation[]>>> {
    return new Observable((sub) => this._observations.subscribe(sub));
  }

  get observations(): Map<string, Map<string, pb.Observation[]>> {
    return this._observations.value;
  }

  get scanInfo$(): Observable<Map<string, StatusInfo>> {
    return new Observable((sub) => this._scanIdsStatus.subscribe(sub));
  }

  get scanInfo(): Map<string, StatusInfo> {
    return this._scanIdsStatus.value;
  }

  // TODO: Make this return an observable.
  fetchObservations(resourceGroups: string[]) {
    this._service.listObservations(resourceGroups).subscribe((obs) => {
      this._observations.next(obs);
    });
  }

  getScanStatus$(scanId: string): Observable<StatusInfo> {
    return this._service.getScanStatus(scanId).pipe(
      map((res) => {
        // A shallow copy here is enough
        let scanInfo = new Map(this.scanInfo);
        const info = {
          state: res.getCollectStatus() !== pb.RequestStatus.DONE ? res.getCollectStatus() : res.getScanStatus(),
          resourceGroups: this.scanInfo.get(scanId)?.resourceGroups as string[],
        };

        if (info.state === pb.RequestStatus.UNKNOWN) {
          console.error(
            `state of scan ${scanId} is UNKNOWN, the modron service may have restarted`
          );
          scanInfo.delete(scanId);
        } else if (info.state === pb.RequestStatus.DONE || info.state === pb.RequestStatus.CANCELLED) {
          // Scan is done, remove from state
          scanInfo.delete(scanId);
        } else {
          scanInfo.set(scanId, info);
        }
        this._scanIdsStatus.next(scanInfo);
        return info;
      })
    );
  }

  collectAndScan$(resourceGroups: string[]): Observable<pb.CollectAndScanResponse> {
    return this._service.collectAndScan(resourceGroups).pipe(
      map((res) => {
        // A shallow copy here is enough
        let scanInfo = new Map(this.scanInfo);
        scanInfo.set(res.getScanId(), {
          state: -1,
          resourceGroups: resourceGroups,
        })
        this._collectIdsStatus.next(scanInfo);
        return res;
      })
    );
  }

  private fetchInitialData() {
    this.fetchObservations([]);
  }
}
