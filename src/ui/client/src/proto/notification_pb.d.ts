// package: 
// file: notification.proto

import * as jspb from "google-protobuf"
import * as google_protobuf_field_mask_pb from "google-protobuf/google/protobuf/field_mask_pb"
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb"

export class NotificationException extends jspb.Message {
  getUuid(): string
  setUuid(value: string): void

  getSourceSystem(): string
  setSourceSystem(value: string): void

  getUserEmail(): string
  setUserEmail(value: string): void

  getNotificationName(): string
  setNotificationName(value: string): void

  getJustification(): string
  setJustification(value: string): void

  hasCreatedOnTime(): boolean
  clearCreatedOnTime(): void
  getCreatedOnTime(): google_protobuf_timestamp_pb.Timestamp | undefined
  setCreatedOnTime(value?: google_protobuf_timestamp_pb.Timestamp): void

  hasValidUntilTime(): boolean
  clearValidUntilTime(): void
  getValidUntilTime(): google_protobuf_timestamp_pb.Timestamp | undefined
  setValidUntilTime(value?: google_protobuf_timestamp_pb.Timestamp): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): NotificationException.AsObject
  static toObject(includeInstance: boolean, msg: NotificationException): NotificationException.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: NotificationException, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): NotificationException
  static deserializeBinaryFromReader(message: NotificationException, reader: jspb.BinaryReader): NotificationException
}

export namespace NotificationException {
  export type AsObject = {
    uuid: string,
    sourceSystem: string,
    userEmail: string,
    notificationName: string,
    justification: string,
    createdOnTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    validUntilTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class GetNotificationExceptionRequest extends jspb.Message {
  getUuid(): string
  setUuid(value: string): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): GetNotificationExceptionRequest.AsObject
  static toObject(includeInstance: boolean, msg: GetNotificationExceptionRequest): GetNotificationExceptionRequest.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: GetNotificationExceptionRequest, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): GetNotificationExceptionRequest
  static deserializeBinaryFromReader(message: GetNotificationExceptionRequest, reader: jspb.BinaryReader): GetNotificationExceptionRequest
}

export namespace GetNotificationExceptionRequest {
  export type AsObject = {
    uuid: string,
  }
}

export class CreateNotificationExceptionRequest extends jspb.Message {
  hasException(): boolean
  clearException(): void
  getException(): NotificationException | undefined
  setException(value?: NotificationException): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): CreateNotificationExceptionRequest.AsObject
  static toObject(includeInstance: boolean, msg: CreateNotificationExceptionRequest): CreateNotificationExceptionRequest.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: CreateNotificationExceptionRequest, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): CreateNotificationExceptionRequest
  static deserializeBinaryFromReader(message: CreateNotificationExceptionRequest, reader: jspb.BinaryReader): CreateNotificationExceptionRequest
}

export namespace CreateNotificationExceptionRequest {
  export type AsObject = {
    exception?: NotificationException.AsObject,
  }
}

export class UpdateNotificationExceptionRequest extends jspb.Message {
  hasException(): boolean
  clearException(): void
  getException(): NotificationException | undefined
  setException(value?: NotificationException): void

  hasUpdateMask(): boolean
  clearUpdateMask(): void
  getUpdateMask(): google_protobuf_field_mask_pb.FieldMask | undefined
  setUpdateMask(value?: google_protobuf_field_mask_pb.FieldMask): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): UpdateNotificationExceptionRequest.AsObject
  static toObject(includeInstance: boolean, msg: UpdateNotificationExceptionRequest): UpdateNotificationExceptionRequest.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: UpdateNotificationExceptionRequest, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): UpdateNotificationExceptionRequest
  static deserializeBinaryFromReader(message: UpdateNotificationExceptionRequest, reader: jspb.BinaryReader): UpdateNotificationExceptionRequest
}

export namespace UpdateNotificationExceptionRequest {
  export type AsObject = {
    exception?: NotificationException.AsObject,
    updateMask?: google_protobuf_field_mask_pb.FieldMask.AsObject,
  }
}

export class DeleteNotificationExceptionRequest extends jspb.Message {
  getUuid(): string
  setUuid(value: string): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): DeleteNotificationExceptionRequest.AsObject
  static toObject(includeInstance: boolean, msg: DeleteNotificationExceptionRequest): DeleteNotificationExceptionRequest.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: DeleteNotificationExceptionRequest, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): DeleteNotificationExceptionRequest
  static deserializeBinaryFromReader(message: DeleteNotificationExceptionRequest, reader: jspb.BinaryReader): DeleteNotificationExceptionRequest
}

export namespace DeleteNotificationExceptionRequest {
  export type AsObject = {
    uuid: string,
  }
}

export class ListNotificationExceptionsRequest extends jspb.Message {
  getUserEmail(): string
  setUserEmail(value: string): void

  getPageSize(): number
  setPageSize(value: number): void

  getPageToken(): string
  setPageToken(value: string): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): ListNotificationExceptionsRequest.AsObject
  static toObject(includeInstance: boolean, msg: ListNotificationExceptionsRequest): ListNotificationExceptionsRequest.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: ListNotificationExceptionsRequest, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): ListNotificationExceptionsRequest
  static deserializeBinaryFromReader(message: ListNotificationExceptionsRequest, reader: jspb.BinaryReader): ListNotificationExceptionsRequest
}

export namespace ListNotificationExceptionsRequest {
  export type AsObject = {
    userEmail: string,
    pageSize: number,
    pageToken: string,
  }
}

export class ListNotificationExceptionsResponse extends jspb.Message {
  clearExceptionsList(): void
  getExceptionsList(): Array<NotificationException>
  setExceptionsList(value: Array<NotificationException>): void
  addExceptions(value?: NotificationException, index?: number): NotificationException

  getNextPageToken(): string
  setNextPageToken(value: string): void

  serializeBinary(): Uint8Array
  toObject(includeInstance?: boolean): ListNotificationExceptionsResponse.AsObject
  static toObject(includeInstance: boolean, msg: ListNotificationExceptionsResponse): ListNotificationExceptionsResponse.AsObject
  static extensions: { [key: number]: jspb.ExtensionFieldInfo<jspb.Message> }
  static extensionsBinary: { [key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message> }
  static serializeBinaryToWriter(message: ListNotificationExceptionsResponse, writer: jspb.BinaryWriter): void
  static deserializeBinary(bytes: Uint8Array): ListNotificationExceptionsResponse
  static deserializeBinaryFromReader(message: ListNotificationExceptionsResponse, reader: jspb.BinaryReader): ListNotificationExceptionsResponse
}

export namespace ListNotificationExceptionsResponse {
  export type AsObject = {
    exceptionsList: Array<NotificationException.AsObject>,
    nextPageToken: string,
  }
}

