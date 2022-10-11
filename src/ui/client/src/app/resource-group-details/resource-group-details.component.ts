import { Component, ChangeDetectionStrategy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ModronService } from '../modron.service';
import { ModronStore } from '../state/modron.store';
import { KeyValue } from '@angular/common';

import * as pb from 'src/proto/modron_pb';

@Component({
  selector: 'app-resource-group-details',
  templateUrl: './resource-group-details.component.html',
  styleUrls: ['./resource-group-details.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ResourceGroupDetailsComponent implements OnInit {
  constructor(
    private _route: ActivatedRoute,
    public store: ModronStore,
    public modron: ModronService
  ) {}

  public resourceGroupName: string = '';

  public displayObsDetail: Map<string, boolean> = new Map<string, boolean>();

  ngOnInit(): void {
    this.resourceGroupName = this._route.snapshot.paramMap.get('id') as string;
  }

  filterName(
    obs: Map<string, pb.Observation[]>
  ): Map<string, pb.Observation[]> {
    let m = new Map();
    m.set(this.resourceGroupName, obs.get(this.resourceGroupName as string));
    return m;
  }

  getName(
    obs: Map<string, Map<string, pb.Observation[]>>
  ): Map<string, pb.Observation[]> {
    return obs.get(this.resourceGroupName) as Map<string, pb.Observation[]>;
  }

  mapByType(obs: Map<string, pb.Observation[]>): Map<string, pb.Observation[]> {
    let obsByType = new Map<string, pb.Observation[]>();
    for (const ob of [...obs.values()].flat()) {
      if (ob === undefined) {
        continue;
      }
      const type = ob.getName();
      if (!obsByType.has(type)) {
        obsByType.set(type, []);
      }
      obsByType.get(type)?.push(ob);
    }
    return obsByType;
  }

  identity(index: number, item: pb.Observation): string {
    return item.getUid();
  }

  identityKV(index: number, item: KeyValue<string, pb.Observation[]>): string {
    return item.key;
  }

  getObservedValue(ob: pb.Observation): string | undefined {
    return ob.getObservedValue()?.toString()?.replace(/,/g, '');
  }

  getExpectedValue(ob: pb.Observation): string | undefined {
    return ob.getExpectedValue()?.toString()?.replace(/,/g, '');
  }

  toggle(id: string | undefined): void {
    id = id as string;
    if (this.displayObsDetail.has(id)) {
      this.displayObsDetail.set(
        id,
        !(this.displayObsDetail.get(id) as boolean)
      );
    } else {
      this.displayObsDetail.set(id, true);
    }
  }
}
