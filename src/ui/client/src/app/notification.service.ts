import { NotificationServiceClient } from "src/proto/notification_pb_service"
import { Injectable } from "@angular/core"
import { Observable, mergeMap, from, EMPTY, concat } from "rxjs"
import { environment } from "src/environments/environment"

import * as pb from "src/proto/notification_pb"

@Injectable({
  providedIn: "root",
})
export class NotificationService {
  public static readonly HOST = environment.production ? "/api" : "";
  private static readonly PAGE_SIZE = 128;

  private _client: NotificationServiceClient

  constructor() {
    this._client = new NotificationServiceClient(NotificationService.HOST)
  }

  createException$(
    exp: pb.NotificationException
  ): Observable<pb.NotificationException> {
    const req = new pb.CreateNotificationExceptionRequest()
    req.setException(exp)

    return new Observable((sub) => {
      this._client.createNotificationException(req, (err, res) => {
        if (err !== null) {
          return sub.error(`createNotificationException: ${err}`)
        }
        if (res === null) {
          return sub.error(
            "createNotificationException: unexpected null response"
          )
        }
        return sub.next(res)
      })
    })
  }

  listExceptions$(userEmail: string): Observable<pb.NotificationException[]> {
    const fetchPage = (
      pageToken: string | null
    ): Observable<pb.ListNotificationExceptionsResponse> => {
      const req = new pb.ListNotificationExceptionsRequest()
      req.setUserEmail(userEmail)
      req.setPageSize(NotificationService.PAGE_SIZE)
      req.setPageToken(pageToken ?? "")

      return new Observable((sub) => {
        this._client.listNotificationExceptions(req, (err, res) => {
          if (err !== null) {
            return sub.error(`listNotificationExceptions: ${err}`)
          }
          if (res === null) {
            return sub.error(
              "listNotificationExceptions: unexpected null response"
            )
          }

          if (res.getNextPageToken() === "") {
            return sub.next(res)
          }
        })
      })
    }
    const fetchExps = (
      pageToken: string | null = null
    ): Observable<pb.NotificationException[]> => {
      return fetchPage(pageToken).pipe(
        mergeMap((res) => {
          const exps$ = from([res.getExceptionsList()])
          const nextExps$ =
            res.getNextPageToken() !== ""
              ? fetchExps(res.getNextPageToken())
              : EMPTY
          return concat(exps$, nextExps$)
        })
      )
    }
    return fetchExps()
  }
}
