import {environment} from "src/environments/environment"
import {ModronServiceClient} from "../proto/ModronServiceClientPb"

import {Injectable} from "@angular/core"
import {concat, EMPTY, from, mergeMap, Observable} from "rxjs"
import {
  CollectAndScanRequest,
  CollectAndScanResponse,
  GetStatusCollectAndScanRequest,
  GetStatusCollectAndScanResponse,
  ListObservationsRequest,
  ListObservationsResponse,
  Observation
} from "../proto/modron_pb";

@Injectable({
  providedIn: "root",
})
export class ModronService {
  public static readonly SEPARATOR = "::"
  public static readonly HOST = environment.production ? "/api" : "";
  private static readonly PAGE_SIZE = 128;

  private _client: ModronServiceClient

  constructor() {
    this._client = new ModronServiceClient(ModronService.HOST)
  }

  listObservations(
    resourceGroups: string[]
  ): Observable<Map<string, Map<string, Observation[]>>> {
    const fetchPage = (
      pageToken: string | null
    ): Observable<ListObservationsResponse> => {
      const req = new ListObservationsRequest()
      req.setResourceGroupNamesList(resourceGroups)
      req.setPageSize(ModronService.PAGE_SIZE)
      req.setPageToken(pageToken ?? "")

      return new Observable((sub) => {
        this._client.listObservations(req, {}, (err, res) => {
          if (err !== null) {
            return sub.error(`listObservations: ${err}`)
          }
          if (res === null) {
            return sub.error("listObservations: unexpected null response")
          }
          if (res.getNextPageToken() === "") {
            return sub.next(res)
          }
        })
      })
    }
    const fetchObs = (
      pageToken: string | null = null
    ): Observable<Map<string, Map<string, Observation[]>>> => {
      return fetchPage(pageToken).pipe(
        mergeMap((res) => {
          // deepcode ignore CollectionUpdatedButNeverQueried: Used, false positive.
          const obs = new Map<string, Map<string, Observation[]>>()
          res.getResourceGroupsObservationsList().forEach((v) => {
            const map = new Map<string, Observation[]>()
            v.getRulesObservationsList().forEach((r) =>
              map.set(r.getRule(), r.getObservationsList())
            )
            obs.set(v.getResourceGroupName(), map)
          })
          const obs$ = from([obs])
          const nextObs$ =
            res.getNextPageToken() !== ""
              ? fetchObs(res.getNextPageToken())
              : EMPTY
          return concat(obs$, nextObs$)
        })
      )
    }
    return fetchObs()
  }

  collectAndScan(resourceGroups: string[]): Observable<CollectAndScanResponse> {
    const fetchPage = (): Observable<CollectAndScanResponse> => {
      const req = new CollectAndScanRequest()
      req.setResourceGroupNamesList(resourceGroups.map(
        (rg) => {
          if (rg.indexOf("/") === -1) {
            return `projects/${rg}`
          }
          return rg
        }
      ))

      return new Observable((sub) => {
        this._client.collectAndScan(req, {}, (err, res) => {
          if (err !== null) {
            return sub.error(`collectAndScan: ${err}`)
          }
          if (res === null) {
            return sub.error("collect: unexpected null response")
          }
          return sub.next(res)
        })
      })
    }
    return fetchPage()
  }

  collectAndScanAll(): Observable<CollectAndScanResponse> {
    const fetchPage = (): Observable<CollectAndScanResponse> => {
      const req = new CollectAndScanRequest()
      return new Observable((sub) => {
        this._client.collectAndScanAll(req, {}, (err, res) => {
          if (err !== null) {
            return sub.error(`collectAndScanAll: ${err}`)
          }
          if (res === null) {
            return sub.error("collectAndScanAll: unexpected null response")
          }
          return sub.next(res)
        })
      })
    }
    return fetchPage()
  }

  getCollectAndScanStatus(IDs: string): Observable<GetStatusCollectAndScanResponse> {
    const req = new GetStatusCollectAndScanRequest()
    req.setCollectId(IDs.split(ModronService.SEPARATOR)[0])
    req.setScanId(IDs.split(ModronService.SEPARATOR)[1])

    return new Observable((sub) => {
      this._client.getStatusCollectAndScan(req, {}, (err, res) => {
        if (err !== null) {
          return sub.error(`getScanStatus: ${err}`)
        }
        if (res === null) {
          return sub.error("getScanStatus: unexpected null response")
        }
        return sub.next(res)
      })
    })
  }
}
