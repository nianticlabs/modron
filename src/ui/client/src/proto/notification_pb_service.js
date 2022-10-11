// package: 
// file: notification.proto

var notification_pb = require("./notification_pb");
var google_protobuf_empty_pb = require("google-protobuf/google/protobuf/empty_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var NotificationService = (function () {
  function NotificationService() {}
  NotificationService.serviceName = "NotificationService";
  return NotificationService;
}());

NotificationService.GetNotificationException = {
  methodName: "GetNotificationException",
  service: NotificationService,
  requestStream: false,
  responseStream: false,
  requestType: notification_pb.GetNotificationExceptionRequest,
  responseType: notification_pb.NotificationException
};

NotificationService.CreateNotificationException = {
  methodName: "CreateNotificationException",
  service: NotificationService,
  requestStream: false,
  responseStream: false,
  requestType: notification_pb.CreateNotificationExceptionRequest,
  responseType: notification_pb.NotificationException
};

NotificationService.UpdateNotificationException = {
  methodName: "UpdateNotificationException",
  service: NotificationService,
  requestStream: false,
  responseStream: false,
  requestType: notification_pb.UpdateNotificationExceptionRequest,
  responseType: notification_pb.NotificationException
};

NotificationService.DeleteNotificationException = {
  methodName: "DeleteNotificationException",
  service: NotificationService,
  requestStream: false,
  responseStream: false,
  requestType: notification_pb.DeleteNotificationExceptionRequest,
  responseType: google_protobuf_empty_pb.Empty
};

NotificationService.ListNotificationExceptions = {
  methodName: "ListNotificationExceptions",
  service: NotificationService,
  requestStream: false,
  responseStream: false,
  requestType: notification_pb.ListNotificationExceptionsRequest,
  responseType: notification_pb.ListNotificationExceptionsResponse
};

exports.NotificationService = NotificationService;

function NotificationServiceClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

NotificationServiceClient.prototype.getNotificationException = function getNotificationException(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(NotificationService.GetNotificationException, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

NotificationServiceClient.prototype.createNotificationException = function createNotificationException(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(NotificationService.CreateNotificationException, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

NotificationServiceClient.prototype.updateNotificationException = function updateNotificationException(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(NotificationService.UpdateNotificationException, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

NotificationServiceClient.prototype.deleteNotificationException = function deleteNotificationException(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(NotificationService.DeleteNotificationException, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

NotificationServiceClient.prototype.listNotificationExceptions = function listNotificationExceptions(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(NotificationService.ListNotificationExceptions, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.NotificationServiceClient = NotificationServiceClient;

