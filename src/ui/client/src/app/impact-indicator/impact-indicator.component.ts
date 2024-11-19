import {Component, Input} from "@angular/core";
import {Impact, Severity} from "../../proto/modron_pb";

@Component(
  {
    selector: "app-impact-indicator",
    templateUrl: "./impact-indicator.component.html",
    styleUrls: ["./impact-indicator.component.scss"],
  }
)
export class ImpactIndicatorComponent {
  @Input()
  public impact: Impact = Impact.IMPACT_UNKNOWN;
  constructor() {
  }

  // We reuse the severity indicator component to display the impact
  public severity(): Severity {
    switch (this.impact) {
      case Impact.IMPACT_HIGH:
        return Severity.SEVERITY_HIGH;
      case Impact.IMPACT_MEDIUM:
        return Severity.SEVERITY_MEDIUM;
      case Impact.IMPACT_LOW:
        return Severity.SEVERITY_LOW;
      default:
        return Severity.SEVERITY_UNKNOWN;
    }
  }
}
