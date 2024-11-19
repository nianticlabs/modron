import {RequestStatus, ScanType} from "../../proto/modron_pb";

export type StatusInfo = {
  state: RequestStatus
  resourceGroups: string[]
  scanType: ScanType
}
