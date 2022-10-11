import { ModronServiceClient } from 'src/proto/modron_pb_service';
import { environment } from 'src/environments/environment';

import { Injectable } from '@angular/core';
import { Observable, mergeMap, from, EMPTY, concat } from 'rxjs';

import * as pb from 'src/proto/modron_pb';

@Injectable({
  providedIn: 'root',
})
export class ModronService {
  public static readonly HOST = environment.production ? '/api' : '';
  private static readonly PAGE_SIZE = 128;

  private _client: ModronServiceClient;

  constructor() {
    this._client = new ModronServiceClient(ModronService.HOST);
  }

  listObservations(
    resourceGroups: string[]
  ): Observable<Map<string, Map<string, pb.Observation[]>>> {
    const fetchPage = (
      pageToken: string | null
    ): Observable<pb.ListObservationsResponse> => {
      let req = new pb.ListObservationsRequest();
      req.setResourceGroupNamesList(resourceGroups);
      req.setPageSize(ModronService.PAGE_SIZE);
      req.setPageToken(pageToken ?? '');

      return new Observable((sub) => {
        this._client.listObservations(req, (err, res) => {
          if (err !== null) {
            return sub.error(`listObservations: ${err}`);
          }
          if (res === null) {
            return sub.error(`listObservations: unexpected null response`);
          }
          if (res.getNextPageToken() === '') {
            return sub.next(res);
          }
        });
      });
    };
    const fetchObs = (
      pageToken: string | null = null
    ): Observable<Map<string, Map<string, pb.Observation[]>>> => {
      return fetchPage(pageToken).pipe(
        mergeMap((res) => {
          const obs = new Map<string, Map<string, pb.Observation[]>>();
          res.getResourceGroupsObservationsList().forEach((v) => {
            let map = new Map<string, pb.Observation[]>();
            v.getRulesObservationsList().forEach((r) =>
              map.set(r.getRule(), r.getObservationsList())
            );
            obs.set(v.getResourceGroupName(), map);
          });
          const obs$ = from([obs]);
          const nextObs$ =
            res.getNextPageToken() !== ''
              ? fetchObs(res.getNextPageToken())
              : EMPTY;
          return concat(obs$, nextObs$);
        })
      );
    };
    return fetchObs();
  }

  collectAndScan(resourceGroups: string[]): Observable<pb.CollectAndScanResponse> {
    const fetchPage = (): Observable<pb.CollectAndScanResponse> => {
      let req = new pb.CollectAndScanRequest();
      req.setResourceGroupNamesList(resourceGroups);

      return new Observable((sub) => {
        this._client.collectAndScan(req, (err, res) => {
          if (err !== null) {
            return sub.error(`collectAndScan: ${err}`);
          }
          if (res === null) {
            return sub.error(`collect: unexpected null response`);
          }
          return sub.next(res);
        });
      });
    };
    return fetchPage();
  }

  getScanStatus(scanId: string): Observable<pb.GetStatusCollectAndScanResponse> {
    const fetchPage = (): Observable<pb.GetStatusCollectAndScanResponse> => {
      let req = new pb.GetStatusCollectAndScanRequest();
      req.setScanId(scanId);

      return new Observable((sub) => {
        this._client.getStatusCollectAndScan(req, (err, res) => {
          if (err !== null) {
            return sub.error(`getScanStatus: ${err}`);
          }
          if (res === null) {
            return sub.error(`getScanStatus: unexpected null response`);
          }
          return sub.next(res);
        });
      });
    };
    return fetchPage();
  }
}
