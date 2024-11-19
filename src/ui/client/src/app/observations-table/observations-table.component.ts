import {
  AfterViewInit,
  ChangeDetectionStrategy, Component, Input, OnInit,
  ViewChild
} from "@angular/core"
import { ModronService } from "../modron.service"
import { ModronStore } from "../state/modron.store"

import * as pb from "src/proto/modron_pb"
import {Observation} from "src/proto/modron_pb";
import {MatSort, Sort} from "@angular/material/sort";
import {MatTableDataSource} from "@angular/material/table";
import {MatDialog} from "@angular/material/dialog";
import {ObservationDetailsDialogComponent} from "../observation-details-dialog/observation-details-dialog.component";

type ObsMap = Map<string, pb.Observation[]>
type RgObsMap = Map<string, ObsMap>

@Component({
  selector: "app-observations-table",
  templateUrl: "./observations-table.component.html",
  styleUrls: ["./observations-table.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ObservationsTableComponent implements OnInit,AfterViewInit {
  public dataSource: MatTableDataSource<pb.Observation> = new MatTableDataSource();
  public sortedData: MatTableDataSource<pb.Observation> = new MatTableDataSource();

  @Input()
  public obs: pb.Observation[] = [];
  constructor(
    public store: ModronStore,
    public modron: ModronService,
    private dialog: MatDialog,
  ){

  }

  @Input()
  public columns = ["riskScore", "category", "shortDesc", "resource", "actions"];
  public resourceGroupName = "";


  async ngOnInit(): Promise<void> {
      this.dataSource.data = this.obs
      this.sortData({active: "riskScore", direction: "desc"})
  }

  @ViewChild(MatSort) sort: MatSort|undefined

  sortData(sort: Sort) {
    if (!sort.active || sort.direction === "") {
      this.sortedData.data = this.dataSource.data;
      return;
    }

    this.sortedData.data = this.dataSource.data.slice().sort((a, b) => {
      let sortResult = 0;
      switch(sort.active) {
        case "riskScore":
          sortResult = a.getRiskScore() - b.getRiskScore();
          break;
        case "category":
          sortResult = a.getName().localeCompare(b.getName());
          break;
      }
      return sort.direction === "asc" ? sortResult : -sortResult;
    });
  }

  ngAfterViewInit(): void {
    this.dataSource.sort = this.sort as MatSort
  }

  getName(obs: RgObsMap): ObsMap | undefined {
    return obs.get(this.resourceGroupName)
  }
  identity(index: number, item: pb.Observation): string {
    return item.getUid()
  }

  protected readonly JSON = JSON;
  protected readonly Object = Object;
  protected readonly Observation = Observation;

  r(row: unknown): pb.Observation {
    return row as pb.Observation
  }

  showObservationDetails(row: Observation) {
    this.dialog.open(ObservationDetailsDialogComponent, {
      data: row,
      width: "50%",
      hasBackdrop: true,
    })
  }
}
