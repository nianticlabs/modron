import {
  Component,
  OnInit,
  Input,
  ChangeDetectionStrategy,
} from '@angular/core';
import { Observation } from 'src/proto/modron_pb';

@Component({
  selector: 'app-search-obs',
  templateUrl: './search-obs.component.html',
  styleUrls: ['./search-obs.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SearchObsComponent implements OnInit {
  searchResource = '';
  searchObservedVal = '';
  searchGroup = '';
  constructor() {}

  @Input() obs: Observation[] = [];

  ngOnInit(): void {}

  applyFilter(event: Event): string {
    return (event.target as HTMLInputElement).value;
  }

  identity(index: number, item: Observation): string {
    return item.getUid();
  }
}
