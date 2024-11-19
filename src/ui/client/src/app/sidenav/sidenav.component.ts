import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import {MatIconModule} from "@angular/material/icon";
import {MatListModule} from "@angular/material/list";
import {Router, RouterLink} from "@angular/router";
import {MatRippleModule} from "@angular/material/core";

@Component({
  selector: "app-sidenav",
  standalone: true,
  imports: [CommonModule, MatIconModule, MatListModule, RouterLink, MatRippleModule],
  templateUrl: "./sidenav.component.html",
  styleUrl: "./sidenav.component.scss"
})
export class SidenavComponent {
  constructor(public router: Router) {}

  public navItems = [
    { link: "/modron/resourcegroups", name: "Resource Groups", icon: "folder" },
    { link: "/modron/stats", name: "Stats", icon: "bar_chart" },
    { link: "/modron/exceptions", name: "Exceptions", icon: "notifications_paused" },
  ];
}
