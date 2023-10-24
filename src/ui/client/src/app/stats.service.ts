import { Injectable } from "@angular/core"
import { ModronService } from "./modron.service"
import { Observation } from "src/proto/modron_pb"

@Injectable({
  providedIn: "root",
})
export class StatsService {
  constructor(public modron: ModronService) { }

  getObservationsPerType(): Map<string, Observation[]> {
    throw Error("unimplemented")
  }
}
