import {Component, Input} from "@angular/core";
import {Impact, Observation} from "../../proto/modron_pb";

@Component({
  selector: "app-observation-details-dialog-content",
  templateUrl: "./observation-details-dialog-content.component.html",
  styleUrls: ["./observation-details-dialog-content.component.scss"],
})
export class ObservationDetailsDialogContentComponent {
  @Input()
  observation!: Observation;
  constructor() {}

    protected readonly Impact = Impact;
}
