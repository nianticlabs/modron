import { Component, OnInit, Input } from '@angular/core';
import { map, Observable } from 'rxjs';
import { ModronStore } from '../state/modron.store';

@Component({
  selector: 'app-resource-group',
  templateUrl: './resource-group.component.html',
  styleUrls: ['./resource-group.component.scss'],
})
export class ResourceGroupComponent implements OnInit {
  @Input()
  name: string = '';

  @Input()
  lastScanDate = '';

  @Input()
  provider = '';

  @Input()
  observationCount: number = -1;

  constructor(public store: ModronStore) {}

  ngOnInit(): void {}

  collectAndScan(resourceGroups: string[]): void {
    this.store
      .collectAndScan$(resourceGroups)
      .subscribe((res) => console.log(res.getCollectId()));
  }

  isCollectionRunning$(project: string): Observable<boolean> {
    return this.store.collectInfo$.pipe(
      map((info) => {
        let running = false;
        for (const [_, v] of info) {
          if (v.state !== 1) {
            if (
              v.resourceGroups.includes(project) ||
              v.resourceGroups.length === 0
            ) {
              running = true;
            }
          }
        }
        return running;
      })
    );
  }
}
