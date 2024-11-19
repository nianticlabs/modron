import { Component, EventEmitter, Input, Output } from "@angular/core"
import { environment } from "../../environments/environment"
import { Router } from "@angular/router";

@Component({
  selector: "app-modron-app",
  templateUrl: "./modron-app.component.html",
  styleUrls: ["./modron-app.component.scss"],
})
export class ModronAppComponent {
  public organization: string
  public href= "";

  @Input() isExpanded: boolean = false;
  @Output() toggleMenu = new EventEmitter();

  constructor(private router: Router) {
    this.organization = environment.organization
  }

  get production(): boolean {
    return environment.production
  }

  get currentUrl(): string {
    return this.router.url;
  }

  public navItems = [
    { link: "/modron/resourcegroups", name: "Resource Groups", icon: "folder" },
    { link: "/modron/stats", name: "Stats", icon: "bar_chart" },
    { link: "/modron/exceptions", name: "Exceptions", icon: "notifications_paused" },
  ];
}
