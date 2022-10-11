// package: 
// file: notification.proto

import * as notification_pb from "./notification_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {grpc} from "@improbable-eng/grpc-web";

type NotificationServiceGetNotificationException = {
  readonly methodName: string;
  readonly service: typeof NotificationService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof notification_pb.GetNotificationExceptionRequest;
  readonly responseType: typeof notification_pb.NotificationException;
};

type NotificationServiceCreateNotificationException = {
  readonly methodName: string;
  readonly service: typeof NotificationService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof notification_pb.CreateNotificationExceptionRequest;
  readonly responseType: typeof notification_pb.NotificationException;
};

type NotificationServiceUpdateNotificationException = {
  readonly methodName: string;
  readonly service: typeof NotificationService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof notification_pb.UpdateNotificationExceptionRequest;
  readonly responseType: typeof notification_pb.NotificationException;
};

type NotificationServiceDeleteNotificationException = {
  readonly methodName: string;
  readonly service: typeof NotificationService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof notification_pb.DeleteNotificationExceptionRequest;
  readonly responseType: typeof google_protobuf_empty_pb.Empty;
};

type NotificationServiceListNotificationExceptions = {
  readonly methodName: string;
  readonly service: typeof NotificationService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof notification_pb.ListNotificationExceptionsRequest;
  readonly responseType: typeof notification_pb.ListNotificationExceptionsResponse;
};

export class NotificationService {
  static readonly serviceName: string;
  static readonly GetNotificationException: NotificationServiceGetNotificationException;
  static readonly CreateNotificationException: NotificationServiceCreateNotificationException;
  static readonly UpdateNotificationException: NotificationServiceUpdateNotificationException;
  static readonly DeleteNotificationException: NotificationServiceDeleteNotificationException;
  static readonly ListNotificationExceptions: NotificationServiceListNotificationExceptions;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class NotificationServiceClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  getNotificationException(
    requestMessage: notification_pb.GetNotificationExceptionRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  getNotificationException(
    requestMessage: notification_pb.GetNotificationExceptionRequest,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  createNotificationException(
    requestMessage: notification_pb.CreateNotificationExceptionRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  createNotificationException(
    requestMessage: notification_pb.CreateNotificationExceptionRequest,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  updateNotificationException(
    requestMessage: notification_pb.UpdateNotificationExceptionRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  updateNotificationException(
    requestMessage: notification_pb.UpdateNotificationExceptionRequest,
    callback: (error: ServiceError|null, responseMessage: notification_pb.NotificationException|null) => void
  ): UnaryResponse;
  deleteNotificationException(
    requestMessage: notification_pb.DeleteNotificationExceptionRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: google_protobuf_empty_pb.Empty|null) => void
  ): UnaryResponse;
  deleteNotificationException(
    requestMessage: notification_pb.DeleteNotificationExceptionRequest,
    callback: (error: ServiceError|null, responseMessage: google_protobuf_empty_pb.Empty|null) => void
  ): UnaryResponse;
  listNotificationExceptions(
    requestMessage: notification_pb.ListNotificationExceptionsRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: notification_pb.ListNotificationExceptionsResponse|null) => void
  ): UnaryResponse;
  listNotificationExceptions(
    requestMessage: notification_pb.ListNotificationExceptionsRequest,
    callback: (error: ServiceError|null, responseMessage: notification_pb.ListNotificationExceptionsResponse|null) => void
  ): UnaryResponse;
}

