import { Component } from "@angular/core"
import { environment } from "src/environments/environment"

@Component({
  selector: "app-modron-app",
  templateUrl: "./modron-app.component.html",
  styleUrls: ["./modron-app.component.scss"],
})
export class ModronAppComponent {
  public organization: string

  constructor() {
    this.organization = environment.organization
  }

  get production(): boolean {
    return environment.production
  }
}
