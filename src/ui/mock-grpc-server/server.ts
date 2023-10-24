import { randomUUID } from 'crypto'
import { dirname } from 'path'
import { fileURLToPath } from 'url'

import { GrpcMockServer, ProtoUtils } from '@alenon/grpc-mock-server'
import * as grpc from '@grpc/grpc-js'
import * as proto_loader from '@grpc/proto-loader'

const __dirname = dirname(fileURLToPath(import.meta.url))

class ModronMockGrpcServer {
  private static readonly MODRON_PROTO_PATH: string = __dirname + '/../../proto/modron.proto'
  private static readonly NOTIFICATION_PROTO_PATH: string = __dirname + '/../../proto/notification.proto'
  private static readonly PKG_NAME: string = ''
  private static readonly MODRON_SERVICE_NAME: string = 'ModronService'
  private static readonly NOTIFICATION_SERVICE_NAME: string = 'NotificationService'
  private static readonly BIND_PORT: number = 4202

  private readonly _modronPkgDef: any
  private readonly _notificationPkgDef: any

  readonly mpb: any
  readonly npb: any
  readonly server: GrpcMockServer

  private _scanIds: Map<string, number> = new Map<string, number>()
  private _collectIds: Map<string, number> = new Map<string, number>()
  private _exceptions: any[]

  constructor() {
    // Patch function since we do not have a package name in the proto.
    ProtoUtils.getProtoFromPkgDefinition = (_, pkgDef) => {
      return pkgDef
    }
    this._modronPkgDef = grpc.loadPackageDefinition(
      proto_loader.loadSync(ModronMockGrpcServer.MODRON_PROTO_PATH),
    )
    this._notificationPkgDef = grpc.loadPackageDefinition(
      proto_loader.loadSync(ModronMockGrpcServer.NOTIFICATION_PROTO_PATH),
    )
    this.mpb = ProtoUtils.getProtoFromPkgDefinition(
      ModronMockGrpcServer.PKG_NAME,
      this._modronPkgDef,
    )
    this.npb = ProtoUtils.getProtoFromPkgDefinition(
      ModronMockGrpcServer.PKG_NAME,
      this._notificationPkgDef,
    )
    this._exceptions = [
      new this.npb.NotificationException.constructor({
        uuid: '9e211723-b6ef-4027-8e6b-4bb4badc1ac6',
        sourceSystem: 'modron',
        notificationName: 'project-project1-observation1-observation1', // resourceGroup-resource-rule
        userEmail: 'foo@bar.com',
        justification: 'some justification',
      }),
      new this.npb.NotificationException.constructor({
        uuid: '03662ac3-9475-4fac-9909-8fc0cc0c1890',
        sourceSystem: 'modron',
        notificationName: 'project-project1-observation2-observation2', // resourceGroup-resource-rule
        userEmail: 'foo@bar.com',
        justification: 'some justification',
      }),
      new this.npb.NotificationException.constructor({
        uuid: '03662ac3-9475-4fac-9909-8fc0cc0c1890',
        sourceSystem: 'modron',
        notificationName: 'project-project-project-project-project-project-project-project3-observation1-observation1', // resourceGroup-resource-rule
        userEmail: 'foo@bar.com',
        justification: 'some justification',
      }),
    ]
    this.server = new GrpcMockServer(`0.0.0.0:${ModronMockGrpcServer.BIND_PORT}`)
  }

  public async run(): Promise<void> {
    await this.initMockServer()
  }

  private async initMockServer() {
    const modron_impls = {
      GetStatusCollectAndScan: (call: any, cb: any) => {
        const req = call.request
        let collectStatus = 0
        let scanStatus = 0
        if (this._collectIds.has(req.collectId)) {
          if (this._collectIds.get(req.collectId) === 1) {
            collectStatus = 1
          } else {
            collectStatus = 2
          }
        }
        if (this._scanIds.has(req.scanId)) {
          if (this._scanIds.get(req.scanId) === 1) {
            scanStatus = 1
          } else {
            scanStatus = 2
          }
        }
        let scanInfo = {
          collectStatus: collectStatus,
          scanStatus: scanStatus,
        }
        console.log(`GetStatusCollectAndScan request inbound: ${JSON.stringify(req)} ${JSON.stringify(scanInfo)}`)
        cb(null, new this.mpb.GetStatusCollectAndScanResponse.constructor(scanInfo))
      },
      CollectAndScan: (call: any, cb: any) => {
        const req = call.request
        const collectId = "collect-" + this._collectIds.size
        const scanId = "scan-" + this._scanIds.size
        this._collectIds.set(collectId, 2)
        this._scanIds.set(scanId, 2)
        console.log(`scan request inbound: ${JSON.stringify(req)} ${scanId}`)
        cb(null, new this.mpb.CollectAndScanResponse.constructor({
          collectId: collectId,
          scanId: scanId,
        }))
        setTimeout(() => {
          console.log(`update ${scanId} to 1`)
          this._collectIds.set(collectId, 1)
          this._scanIds.set(scanId, 1)
        }, 10000)
      },
      ListObservations: (call: any, cb: any) => {
        const req = call.request
        console.log(`listObservations request inbound: ${JSON.stringify(req)}`)

        if (req.resourceGroupNames === undefined || req.resourceGroupNames.length === 0) {
          cb(null, new this.mpb.ListObservationsResponse.constructor({
            resourceGroupsObservations: [
              [0, [], "projects/no-findings"],
              [0, [1], "projects/project-1"],
              [0, [1], "projects/project-2"],
              [0, [1], "projects/project-3"],
              [1, [1, 2, 3, 4], "projects/project-4"],
              [2, [1, 2, 3, 4, 5], "projects/project-5"],
              [3, [1], "projects/project-with-a-very-long-name"],
              [0, [1], "projects/project-7"],
              [0, [1], "projects/project-8"],
              [0, [1], "projects/project-9"],
              [1, [1, 2, 3, 4], "projects/project-10"],
              [2, [1, 2, 3, 4, 5], "projects/project-11"],
              [3, [1], "projects/project-12"]].
              map(e => new this.mpb.ResourceGroupObservationsPair.constructor({
                resourceGroupName: e[2],
                rulesObservations: [0, 1, 2, 3, 4, 5, 6].map(ruleNb => new this.mpb.RuleObservationPair.constructor({
                  rule: "observation" + ruleNb,
                  observations: (e[1] as Array<number>).filter(ele => ele === ruleNb).map(e1 => new this.mpb.Observation.constructor({
                    name: "observation" + e1,
                    timestamp: {
                      seconds: 123,
                      nanos: 456,
                    },
                    uid: "5cedca54-a6e0-4de5-8df5-facc533f5903--" + e1,
                    remediation: new this.mpb.Remediation.constructor({
                      description: "some description [title](https://www.example.com)",
                      recommendation: "do something [title](https://www.example.com)",
                    }),
                    resource: new this.mpb.Resource.constructor({
                      name: "project" + e[0] + "[observation" + e1 + "]",
                      resourceGroupName: "project" + e[0],
                      timestamp: {
                        seconds: 1273,
                        nanos: 456,
                      },
                    }),
                  })),
                }))
              })),
            nextPageToken: '',
          }))
        } else {
          cb(null, new this.mpb.ListObservationsResponse.constructor({
            resourceGroupsObservations: [
              [0, [1], req.resourceGroupNames[0]],
            ].
              map(e => new this.mpb.ResourceGroupObservationsPair.constructor({
                resourceGroupName: e[2],
                rulesObservations: [0, 1, 2, 3, 4, 5, 6].map(ruleNb => new this.mpb.RuleObservationPair.constructor({
                  rule: "observation" + ruleNb,
                  observations: (e[1] as Array<number>).filter(ele => ele === ruleNb).map(e1 => new this.mpb.Observation.constructor({
                    name: "observation" + e1,
                    timestamp: {
                      seconds: new Date().getTime() / 1000,
                      nanos: 456,
                    },
                    uid: "5cedca54-a6e0-4de5-8df5-facc533f5903--" + e1,
                    remediation: new this.mpb.Remediation.constructor({
                      description: "some description [title](https://www.example.com)",
                      recommendation: "do something [title](https://www.example.com)",
                    }),
                    resource: new this.mpb.Resource.constructor({
                      name: "project" + e[0] + "[observation" + e1 + "]",
                      resourceGroupName: "project" + e[0],
                      timestamp: {
                        seconds: new Date().getTime() / 1000,
                        nanos: 456,
                      },
                    }),
                  })),
                }))
              })),
            nextPageToken: '',
          }))
        }
      },
    }
    const notification_impls = {
      ListNotificationExceptions: (call: any, cb: any) => {
        const req = call.request
        console.log(`listNotificationExceptions request inbound: ${JSON.stringify(req)}`)
        cb(null, new this.npb.ListNotificationExceptionsResponse.constructor({
          exceptions: this._exceptions,
          nextPageToken: '',
        }))
      },
      CreateNotificationException: (call: any, cb: any) => {
        const req = call.request
        console.log(`createNotificationException request inbound: ${JSON.stringify(req)}`)
        let exp = new this.npb.NotificationException.constructor({
          uuid: randomUUID(),
          sourceSystem: req.exception.sourceSystem,
          notificationName: req.exception.notificationName,
          userEmail: req.exception.userEmail,
          justification: req.exception.justification,
          validUntilTime: req.exception.validUntilTime,
        })
        this._exceptions.push(exp)
        cb(null, exp)
      },
    }
    this.server.addService(
      ModronMockGrpcServer.MODRON_PROTO_PATH,
      ModronMockGrpcServer.PKG_NAME,
      ModronMockGrpcServer.MODRON_SERVICE_NAME,
      modron_impls,
    )
    this.server.addService(
      ModronMockGrpcServer.NOTIFICATION_PROTO_PATH,
      ModronMockGrpcServer.PKG_NAME,
      ModronMockGrpcServer.NOTIFICATION_SERVICE_NAME,
      notification_impls,
    )
    try {
      await this.server.start()
      console.log(`Modron mock gRPC server is listening at: ${this.server.serverAddress}`)
    } catch (error) {
      throw new Error(`failed initializing modron mock gRPC server at: ${this.server.serverAddress}`)
    }
  }
}

const server: ModronMockGrpcServer = new ModronMockGrpcServer()
await server.run()

let sigterm = () => {
  process.stdin.resume()

  var p = new Promise<void>(function (resolve, reject) {
    process.on('SIGTERM', function () {
      process.stdin.pause()
      resolve()
    })
  })

  return p
}

await sigterm()
console.log(`Modron mock gRPC server stopped listening at: ${server.server.serverAddress}`)
