import { Pipe, PipeTransform } from "@angular/core"
import { Value } from "google-protobuf/google/protobuf/struct_pb"
import * as pb from "src/proto/modron_pb"

@Pipe({ name: "mapByType" })
export class MapByTypePipe implements PipeTransform {
  transform(obs: Map<string, pb.Observation[]>): Map<string, pb.Observation[]> {
    const obsByType = new Map<string, pb.Observation[]>()
    for (const ob of [...obs.values()].flat()) {
      const type = ob.getName()
      if (!obsByType.has(type)) {
        obsByType.set(type, [])
      }
      obsByType.get(type)?.push(ob)
    }
    return obsByType
  }
}

@Pipe({ name: "mapFlatRules" })
export class mapFlatRulesPipe implements PipeTransform {
  transform(
    map: Map<string, Map<string, pb.Observation[]>>
  ): Map<string, pb.Observation[]> {
    const res = new Map<string, pb.Observation[]>()
    map.forEach((v, k) => {
      res.set(k, Array.from(v.values()).flat())
    })
    return res
  }
}

@Pipe({ name: "mapByObservedValues" })
export class MapByObservedValuesPipe implements PipeTransform {
  transform(obs: pb.Observation[]): Map<string, number> {
    const obsByType = new Map<string, number>()
    obs.forEach((o) => {
      const obsValue = o.getObservedValue()
        ? (o.getObservedValue() as Value).toString()
        : "Observation count"
      if (!obsByType.has(obsValue)) {
        obsByType.set(obsValue, 0)
      }
      obsByType.set(obsValue, (obsByType.get(obsValue) as number) + 1)
    })
    return obsByType
  }
}

@Pipe({ name: "filterName" })
export class FilterNamePipe implements PipeTransform {
  transform(
    obs: Map<string, pb.Observation[]> | null,
    name: string | null
  ): Map<string, pb.Observation[]> {
    const obsByType = new Map<string, pb.Observation[]>()
    obsByType.set(
      name ? name : "",
      obs?.get(name ? name : "")
        ? (obs.get(name ? name : "") as pb.Observation[])
        : []
    )
    return obsByType
  }
}
