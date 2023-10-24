import { NotificationService } from "../notification.service"
import { Injectable } from "@angular/core"
import { BehaviorSubject, map, Observable } from "rxjs"
import { NotificationException } from "../model/notification.model"
import { AuthenticationStore } from "./authentication.store"

@Injectable()
export class NotificationStore {
  private _exceptions: BehaviorSubject<NotificationException[]>

  constructor(
    private _service: NotificationService,
    private _auth: AuthenticationStore
  ) {
    this._exceptions = new BehaviorSubject<NotificationException[]>([])
    this.fetchInitialData()
  }

  get exceptions$(): Observable<NotificationException[]> {
    return new Observable((sub) => this._exceptions.subscribe(sub))
  }

  get exceptions(): NotificationException[] {
    return this._exceptions.value
  }

  createException$(
    exp: NotificationException
  ): Observable<NotificationException> {
    return this._service.createException$(exp.toProto()).pipe(
      map((proto) => {
        const exp = NotificationException.fromProto(proto)
        this._exceptions.next(this._exceptions.getValue().concat([exp]))
        return exp
      })
    )
  }

  private listExceptions(userEmail: string) {
    this._service.listExceptions$(userEmail).subscribe((protos) => {
      this._exceptions.next(
        protos.map((proto) => NotificationException.fromProto(proto))
      )
    })
  }

  private fetchInitialData() {
    this.listExceptions(this._auth.user?.email ?? "")
  }
}
