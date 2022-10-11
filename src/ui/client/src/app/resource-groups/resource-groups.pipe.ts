import { Pipe, PipeTransform } from '@angular/core';
import * as pb from 'src/proto/modron_pb';

@Pipe({ name: 'resourceGroups' })
export class ResourceGroupsPipe implements PipeTransform {
  transform(obs: Map<string, pb.Observation[]>): string[] {
    return [...obs.keys()];
  }
}

@Pipe({ name: 'mapPerTypeName' })
export class MapPerTypeName implements PipeTransform {
  transform(obs: Map<string, pb.Observation[]>): Map<string, pb.Observation[]> {
    let obsn: Map<string, pb.Observation[]> = new Map<
      string,
      pb.Observation[]
    >();

    obsn.set('START', []);
    obsn.set(JSON.stringify(obs.size), []);
    obsn.set('END', []);

    return obsn;
  }
}

@Pipe({ name: 'observations' })
export class ObservationsPipe implements PipeTransform {
  transform(obs: Map<string, pb.Observation[]>): pb.Observation[] {
    return [...obs.values()].flat();
  }
}

@Pipe({ name: 'invalidProjectNb' })
export class InvalidProjectNb implements PipeTransform {
  transform(obs: any[]): number {
    let res: number = 0;
    obs.forEach((e) => {
      if (e.value.length > 0) {
        res += 1;
      }
    });
    return res;
  }
}

@Pipe({ name: 'obsNb' })
export class ObsNbPipe implements PipeTransform {
  transform(obs: any[]): number {
    let res: number = 0;
    obs.forEach((e) => (res += e.value.length));
    return res;
  }
}
