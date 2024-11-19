import {Component, Input} from "@angular/core";
import {Severity} from "../../proto/modron_pb";

@Component({
  selector: "app-severity-indicator",
  templateUrl: "./severity-indicator.component.html",
  styleUrls: ["./severity-indicator.component.scss"],
})
export class SeverityIndicatorComponent {
  @Input()
  severity: Severity = Severity.SEVERITY_UNKNOWN;

  @Input()
  count: number | undefined;

  getIcon(): string {
    switch (this.severity) {
      case Severity.SEVERITY_CRITICAL:
        return "C";
      case Severity.SEVERITY_HIGH:
        return "H";
      case Severity.SEVERITY_MEDIUM:
        return "M";
      case Severity.SEVERITY_LOW:
        return "L";
      case Severity.SEVERITY_INFO:
        return "I";
      default:
        return "?";
    }
  }

  constructor() {}

  protected readonly Severity = Severity;
}
