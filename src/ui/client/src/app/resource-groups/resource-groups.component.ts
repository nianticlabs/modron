import { Component, ChangeDetectionStrategy } from '@angular/core';
import { ModronStore } from '../state/modron.store';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Observation } from 'src/proto/modron_pb';
import { map, Observable } from 'rxjs';
import * as pb from 'src/proto/modron_pb';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: 'app-resource-groups',
  templateUrl: './resource-groups.component.html',
  styleUrls: ['./resource-groups.component.scss'],
})
export class ResourceGroupsComponent {
  public static readonly STATUS_CHECK_INTERVAL_MS = 10000;
  private static readonly SNACKBAR_LINGER_DURATION_MS = 2500;

  searchText = '';
  removeNoObs = false;

  constructor(public store: ModronStore, public snackBar: MatSnackBar) { }

  ngOnInit() {
    setInterval(() => {
      this.store.collectInfo.forEach((v, k) => {
        if (v.state !== 1) {
          this.store.getScanStatus$(k).subscribe({
            next: (scanInfo) => {
                if (scanInfo.state === pb.RequestStatus.DONE) {
                this.snackBar.open('Scan complete', '', {
                  duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS,
                });
                this.store.fetchObservations([]);
              }
            },
            error: () =>
              this.snackBar.open(
                'An unexpected error has occurred while scanning',
                '',
                {
                  duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS,
                }
              ),
          });
        }
      });
    }, ResourceGroupsComponent.STATUS_CHECK_INTERVAL_MS);
  }

  getColor(balance: number): string {
    return balance > 0 ? '#da1e28' : '#24a148';
  }

  collectAndScan(resourceGroups: string[]): void {
    this.store.collectAndScan$(resourceGroups).subscribe({
      next: () =>
        this.snackBar.open('Scanning all resource groups...', '', {
          duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS,
        }),
      error: () =>
        this.snackBar.open(
          'An unexpected error has occurred while starting the collection',
          '',
          { duration: ResourceGroupsComponent.SNACKBAR_LINGER_DURATION_MS }
        ),
    });
  }

  isCollectionRunning$(): Observable<boolean> {
    return this.store.collectInfo$.pipe(
      map((info) => {
        let running = false;
        for (const [_, v] of info) {
          if (v.state !== 1) {
            if (v.resourceGroups.length == 0) {
              running = true;
            }
          }
        }
        return running;
      })
    );
  }

  getDate(obs: any[]): string {
    obs as Observation[];
    if (obs.length > 0) {
      return obs[0].getTimestamp()?.toDate().toUTCString().slice(4);
    }
    return '';
  }
}
