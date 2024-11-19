import {
  Component,
  OnInit,
  Input,
  ChangeDetectionStrategy,
} from "@angular/core"
import {ChartData, ChartOptions} from "chart.js";

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: "app-observations-stats",
  templateUrl: "./observations-stats.component.html",
  styleUrls: ["./observations-stats.component.scss"],
})
export class ObservationsStatsComponent implements OnInit {
  @Input() data: Map<string, number> = new Map<string, number>();
  public options: ChartOptions = {
    scales: {

    },
    indexAxis: "y",
    plugins: {
      legend: {
        display: false,
      },
    }
  }
  public chartData: ChartData = {
    labels: [] as string[],
    datasets: [
      {
        label: "Observations",
        data: [] as number[],
      }
    ]
  };
  max = 1;

  ngOnInit(): void {
    this.max = Math.max(...this.data.values())
    this.chartData.labels = Array.from(this.data.keys());
    this.chartData.datasets[0].data = Array.from(this.data.values());
  }
}
