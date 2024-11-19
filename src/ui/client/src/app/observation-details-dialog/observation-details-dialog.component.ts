import {Component, Inject} from "@angular/core";
import {MAT_DIALOG_DATA} from "@angular/material/dialog";
import {Observation} from "../../proto/modron_pb";

@Component(
  {
    selector: "app-observation-details-dialog",
    templateUrl: "./observation-details-dialog.component.html",
    styleUrls: ["./observation-details-dialog.component.scss"],
  }
)
export class ObservationDetailsDialogComponent {
  constructor(
    @Inject(MAT_DIALOG_DATA) public observation: Observation
  ) {
  }
}
