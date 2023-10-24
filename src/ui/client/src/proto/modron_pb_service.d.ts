// package: 
// file: modron.proto

import * as modron_pb from "./modron_pb"
import { grpc } from "@improbable-eng/grpc-web"

type ModronServiceCollectAndScan = {
  readonly methodName: string
  readonly service: typeof ModronService
  readonly requestStream: false
  readonly responseStream: false
  readonly requestType: typeof modron_pb.CollectAndScanRequest
  readonly responseType: typeof modron_pb.CollectAndScanResponse
}

type ModronServiceListObservations = {
  readonly methodName: string
  readonly service: typeof ModronService
  readonly requestStream: false
  readonly responseStream: false
  readonly requestType: typeof modron_pb.ListObservationsRequest
  readonly responseType: typeof modron_pb.ListObservationsResponse
}

type ModronServiceCreateObservation = {
  readonly methodName: string
  readonly service: typeof ModronService
  readonly requestStream: false
  readonly responseStream: false
  readonly requestType: typeof modron_pb.CreateObservationRequest
  readonly responseType: typeof modron_pb.Observation
}

type ModronServiceGetStatusCollectAndScan = {
  readonly methodName: string
  readonly service: typeof ModronService
  readonly requestStream: false
  readonly responseStream: false
  readonly requestType: typeof modron_pb.GetStatusCollectAndScanRequest
  readonly responseType: typeof modron_pb.GetStatusCollectAndScanResponse
}

export class ModronService {
  static readonly serviceName: string
  static readonly CollectAndScan: ModronServiceCollectAndScan
  static readonly ListObservations: ModronServiceListObservations
  static readonly CreateObservation: ModronServiceCreateObservation
  static readonly GetStatusCollectAndScan: ModronServiceGetStatusCollectAndScan
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void
}
interface ResponseStream<T> {
  cancel(): void
  on(type: "data", handler: (message: T) => void): ResponseStream<T>
  on(type: "end", handler: (status?: Status) => void): ResponseStream<T>
  on(type: "status", handler: (status: Status) => void): ResponseStream<T>
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>
  end(): void
  cancel(): void
  on(type: "end", handler: (status?: Status) => void): RequestStream<T>
  on(type: "status", handler: (status: Status) => void): RequestStream<T>
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>
  end(): void
  cancel(): void
  on(type: "data", handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>
  on(type: "end", handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>
  on(type: "status", handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>
}

export class ModronServiceClient {
  readonly serviceHost: string

  constructor(serviceHost: string, options?: grpc.RpcOptions)
  collectAndScan(
    requestMessage: modron_pb.CollectAndScanRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError | null, responseMessage: modron_pb.CollectAndScanResponse | null) => void
  ): UnaryResponse
  collectAndScan(
    requestMessage: modron_pb.CollectAndScanRequest,
    callback: (error: ServiceError | null, responseMessage: modron_pb.CollectAndScanResponse | null) => void
  ): UnaryResponse
  listObservations(
    requestMessage: modron_pb.ListObservationsRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError | null, responseMessage: modron_pb.ListObservationsResponse | null) => void
  ): UnaryResponse
  listObservations(
    requestMessage: modron_pb.ListObservationsRequest,
    callback: (error: ServiceError | null, responseMessage: modron_pb.ListObservationsResponse | null) => void
  ): UnaryResponse
  createObservation(
    requestMessage: modron_pb.CreateObservationRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError | null, responseMessage: modron_pb.Observation | null) => void
  ): UnaryResponse
  createObservation(
    requestMessage: modron_pb.CreateObservationRequest,
    callback: (error: ServiceError | null, responseMessage: modron_pb.Observation | null) => void
  ): UnaryResponse
  getStatusCollectAndScan(
    requestMessage: modron_pb.GetStatusCollectAndScanRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError | null, responseMessage: modron_pb.GetStatusCollectAndScanResponse | null) => void
  ): UnaryResponse
  getStatusCollectAndScan(
    requestMessage: modron_pb.GetStatusCollectAndScanRequest,
    callback: (error: ServiceError | null, responseMessage: modron_pb.GetStatusCollectAndScanResponse | null) => void
  ): UnaryResponse
}

