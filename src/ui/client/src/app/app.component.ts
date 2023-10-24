import { Component } from "@angular/core"
import { environment } from "src/environments/environment"
import { AuthenticationStore } from "./state/authentication.store"
import { ModronStore } from "./state/modron.store"
import { NotificationStore } from "./state/notification.store"

@Component({
  selector: "app-root",
  templateUrl: "./app.component.html",
  styleUrls: ["./app.component.scss"],
})
export class AppComponent {
  constructor(
    public auth: AuthenticationStore,
    public modron: ModronStore,
    public notification: NotificationStore
  ) { }

  get local(): boolean {
    return environment.local
  }
}
