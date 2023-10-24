import {
  Component,
  OnInit,
  Input,
  ChangeDetectionStrategy,
} from "@angular/core"

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: "app-histogram-horizontal",
  templateUrl: "./histogram-horizontal.component.html",
  styleUrls: ["./histogram-horizontal.component.scss"],
})
export class HistogramHorizontalComponent implements OnInit {
  @Input() data: Map<string, number> = new Map<string, number>();

  max = 1;

  ngOnInit(): void {
    this.max = Math.max(...this.data.values())
  }

  getWidth(key: string): string {
    const value = this.data.get(key) as number
    return (value / this.max) * 100 + "%"
  }

  parseKey(key: string): string {
    return key.replace(/,/g, "")
  }
}
