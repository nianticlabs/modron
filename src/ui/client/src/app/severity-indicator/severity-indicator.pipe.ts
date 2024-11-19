import {Pipe, PipeTransform} from "@angular/core";
import {Impact, Severity} from "../../proto/modron_pb";

@Pipe({name: "severityName"})
export class SeverityNamePipe implements PipeTransform {
  transform(severity: number): string {
    switch(severity) {
      case Severity.SEVERITY_CRITICAL:
        return "Critical";
      case Severity.SEVERITY_HIGH:
        return "High";
      case Severity.SEVERITY_MEDIUM:
        return "Medium";
      case Severity.SEVERITY_LOW:
        return "Low";
      case Severity.SEVERITY_INFO:
        return "Info";
      default:
        return "Unknown";
    }
  }
}

@Pipe({name: "impactName"})
export class ImpactNamePipe implements PipeTransform {
  transform(impact: number): string {
    switch(impact) {
      case Impact.IMPACT_HIGH:
        return "High";
      case Impact.IMPACT_MEDIUM:
        return "Medium";
      case Impact.IMPACT_LOW:
        return "Low";
      default:
        return "Unknown";
    }
  }
}

@Pipe({name: "severityAmount"})
export class SeverityAmountPipe implements PipeTransform {
  transform(count: number): string {
    if(count > 99) {
      return "99+";
    }
    return count.toString();
  }
}
