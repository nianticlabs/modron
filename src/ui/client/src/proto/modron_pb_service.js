// package: 
// file: modron.proto

var modron_pb = require("./modron_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var ModronService = (function () {
  function ModronService() {}
  ModronService.serviceName = "ModronService";
  return ModronService;
}());

ModronService.CollectAndScan = {
  methodName: "CollectAndScan",
  service: ModronService,
  requestStream: false,
  responseStream: false,
  requestType: modron_pb.CollectAndScanRequest,
  responseType: modron_pb.CollectAndScanResponse
};

ModronService.ListObservations = {
  methodName: "ListObservations",
  service: ModronService,
  requestStream: false,
  responseStream: false,
  requestType: modron_pb.ListObservationsRequest,
  responseType: modron_pb.ListObservationsResponse
};

ModronService.CreateObservation = {
  methodName: "CreateObservation",
  service: ModronService,
  requestStream: false,
  responseStream: false,
  requestType: modron_pb.CreateObservationRequest,
  responseType: modron_pb.Observation
};

ModronService.GetStatusCollectAndScan = {
  methodName: "GetStatusCollectAndScan",
  service: ModronService,
  requestStream: false,
  responseStream: false,
  requestType: modron_pb.GetStatusCollectAndScanRequest,
  responseType: modron_pb.GetStatusCollectAndScanResponse
};

exports.ModronService = ModronService;

function ModronServiceClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

ModronServiceClient.prototype.collectAndScan = function collectAndScan(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(ModronService.CollectAndScan, {
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

ModronServiceClient.prototype.listObservations = function listObservations(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(ModronService.ListObservations, {
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

ModronServiceClient.prototype.createObservation = function createObservation(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(ModronService.CreateObservation, {
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

ModronServiceClient.prototype.getStatusCollectAndScan = function getStatusCollectAndScan(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(ModronService.GetStatusCollectAndScan, {
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

exports.ModronServiceClient = ModronServiceClient;

