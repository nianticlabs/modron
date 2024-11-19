import {ChangeDetectionStrategy, Component} from "@angular/core";
import {Impact, Observation, Remediation, Severity} from "../../proto/modron_pb";
import Category = Observation.Category;

@Component({
  selector: "app-ui-demo",
  templateUrl: "./ui-demo.component.html",
  styleUrls: ["./ui-demo.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class UIDemoComponent {
  protected readonly Severity = Severity;
  protected readonly Object = Object;
  protected readonly severityValues: Severity[] = Object.values(Severity).reverse() as Severity[];
  protected readonly Date = Date;
  date: Date | null = new Date();

  public demoObservation = new Observation();
  constructor() {
    this.demoObservation.setName("EXAMPLE_DEMO_OBSERVATION");
    this.demoObservation.setRiskScore(Severity.SEVERITY_CRITICAL);
    this.demoObservation.setSeverity(Severity.SEVERITY_HIGH);
    this.demoObservation.setImpact(Impact.IMPACT_HIGH);
    this.demoObservation.setCategory(Category.CATEGORY_MISCONFIGURATION)
    this.demoObservation.setImpactReason("environment=production")
    const remediation = new Remediation();
    remediation.setDescription("Example description");
    remediation.setRecommendation("Example recommendation");
    this.demoObservation.setRemediation(remediation);
  }

}
