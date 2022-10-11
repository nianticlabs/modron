import * as pb from 'src/proto/notification_pb';
import * as timestampPb from "google-protobuf/google/protobuf/timestamp_pb"

export class NotificationException {
  private _uuid?: string
  private _createdOnTime?: Date

  sourceSystem = ""
  userEmail = ""
  notificationName = ""
  justification = ""
  validUntilTime?: Date

  get uuid(): string {
    if (this._uuid === undefined) {
      throw new Error("exception has not been created")
    }
    return this._uuid
  }

  get createdOnTime(): Date {
    if (this._createdOnTime === undefined) {
      throw new Error("exception has not been created")
    }
    return this._createdOnTime
  }

  toProto(): pb.NotificationException {
    let proto = new pb.NotificationException()
    proto.setSourceSystem(this.sourceSystem)
    proto.setUserEmail(this.userEmail)
    proto.setNotificationName(this.notificationName)
    proto.setJustification(this.justification)
    if (this.validUntilTime !== undefined) {
      proto.setValidUntilTime(timestampPb.Timestamp.fromDate(this.validUntilTime))
    }
    return proto
  }

  static fromProto(proto: pb.NotificationException): NotificationException {
    let model = new NotificationException()
    model._uuid = proto.getUuid()
    model.sourceSystem = proto.getSourceSystem()
    model.userEmail = proto.getUserEmail()
    model.notificationName = proto.getNotificationName()
    model.justification = proto.getJustification()
    model.validUntilTime = proto.getValidUntilTime()?.toDate()
    model._createdOnTime = proto.getCreatedOnTime()?.toDate()
    return model
  }
}
