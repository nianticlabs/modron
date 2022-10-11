// package: 
// file: modron.proto

import * as jspb from "google-protobuf";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as google_protobuf_struct_pb from "google-protobuf/google/protobuf/struct_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class ExportedCredentials extends jspb.Message {
  hasCreationDate(): boolean;
  clearCreationDate(): void;
  getCreationDate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setCreationDate(value?: google_protobuf_timestamp_pb.Timestamp): void;

  hasExpirationDate(): boolean;
  clearExpirationDate(): void;
  getExpirationDate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setExpirationDate(value?: google_protobuf_timestamp_pb.Timestamp): void;

  hasLastUsage(): boolean;
  clearLastUsage(): void;
  getLastUsage(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setLastUsage(value?: google_protobuf_timestamp_pb.Timestamp): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExportedCredentials.AsObject;
  static toObject(includeInstance: boolean, msg: ExportedCredentials): ExportedCredentials.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExportedCredentials, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExportedCredentials;
  static deserializeBinaryFromReader(message: ExportedCredentials, reader: jspb.BinaryReader): ExportedCredentials;
}

export namespace ExportedCredentials {
  export type AsObject = {
    creationDate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    expirationDate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    lastUsage?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class VmInstance extends jspb.Message {
  getPublicIp(): string;
  setPublicIp(value: string): void;

  getPrivateIp(): string;
  setPrivateIp(value: string): void;

  getIdentity(): string;
  setIdentity(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): VmInstance.AsObject;
  static toObject(includeInstance: boolean, msg: VmInstance): VmInstance.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: VmInstance, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): VmInstance;
  static deserializeBinaryFromReader(message: VmInstance, reader: jspb.BinaryReader): VmInstance;
}

export namespace VmInstance {
  export type AsObject = {
    publicIp: string,
    privateIp: string,
    identity: string,
  }
}

export class Network extends jspb.Message {
  clearIpsList(): void;
  getIpsList(): Array<string>;
  setIpsList(value: Array<string>): void;
  addIps(value: string, index?: number): string;

  getGcpPrivateGoogleAccessV4(): boolean;
  setGcpPrivateGoogleAccessV4(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Network.AsObject;
  static toObject(includeInstance: boolean, msg: Network): Network.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Network, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Network;
  static deserializeBinaryFromReader(message: Network, reader: jspb.BinaryReader): Network;
}

export namespace Network {
  export type AsObject = {
    ipsList: Array<string>,
    gcpPrivateGoogleAccessV4: boolean,
  }
}

export class KubernetesCluster extends jspb.Message {
  clearMasterAuthorizedNetworksList(): void;
  getMasterAuthorizedNetworksList(): Array<string>;
  setMasterAuthorizedNetworksList(value: Array<string>): void;
  addMasterAuthorizedNetworks(value: string, index?: number): string;

  getPrivateCluster(): boolean;
  setPrivateCluster(value: boolean): void;

  getMasterVersion(): string;
  setMasterVersion(value: string): void;

  getNodesVersion(): string;
  setNodesVersion(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): KubernetesCluster.AsObject;
  static toObject(includeInstance: boolean, msg: KubernetesCluster): KubernetesCluster.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: KubernetesCluster, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): KubernetesCluster;
  static deserializeBinaryFromReader(message: KubernetesCluster, reader: jspb.BinaryReader): KubernetesCluster;
}

export namespace KubernetesCluster {
  export type AsObject = {
    masterAuthorizedNetworksList: Array<string>,
    privateCluster: boolean,
    masterVersion: string,
    nodesVersion: string,
  }
}

export class Database extends jspb.Message {
  getType(): string;
  setType(value: string): void;

  getVersion(): string;
  setVersion(value: string): void;

  getEncryption(): Database.EncryptionTypeMap[keyof Database.EncryptionTypeMap];
  setEncryption(value: Database.EncryptionTypeMap[keyof Database.EncryptionTypeMap]): void;

  getAddress(): string;
  setAddress(value: string): void;

  getAutoResize(): boolean;
  setAutoResize(value: boolean): void;

  getBackupConfig(): Database.BackupConfigurationMap[keyof Database.BackupConfigurationMap];
  setBackupConfig(value: Database.BackupConfigurationMap[keyof Database.BackupConfigurationMap]): void;

  getPasswordPolicy(): Database.PasswordPolicyMap[keyof Database.PasswordPolicyMap];
  setPasswordPolicy(value: Database.PasswordPolicyMap[keyof Database.PasswordPolicyMap]): void;

  getTlsRequired(): boolean;
  setTlsRequired(value: boolean): void;

  getAuthorizedNetworksSettingAvailable(): Database.AuthorizedNetworksMap[keyof Database.AuthorizedNetworksMap];
  setAuthorizedNetworksSettingAvailable(value: Database.AuthorizedNetworksMap[keyof Database.AuthorizedNetworksMap]): void;

  clearAuthorizedNetworksList(): void;
  getAuthorizedNetworksList(): Array<string>;
  setAuthorizedNetworksList(value: Array<string>): void;
  addAuthorizedNetworks(value: string, index?: number): string;

  getAvailabilityType(): Database.AvailabilityTypeMap[keyof Database.AvailabilityTypeMap];
  setAvailabilityType(value: Database.AvailabilityTypeMap[keyof Database.AvailabilityTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Database.AsObject;
  static toObject(includeInstance: boolean, msg: Database): Database.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Database, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Database;
  static deserializeBinaryFromReader(message: Database, reader: jspb.BinaryReader): Database;
}

export namespace Database {
  export type AsObject = {
    type: string,
    version: string,
    encryption: Database.EncryptionTypeMap[keyof Database.EncryptionTypeMap],
    address: string,
    autoResize: boolean,
    backupConfig: Database.BackupConfigurationMap[keyof Database.BackupConfigurationMap],
    passwordPolicy: Database.PasswordPolicyMap[keyof Database.PasswordPolicyMap],
    tlsRequired: boolean,
    authorizedNetworksSettingAvailable: Database.AuthorizedNetworksMap[keyof Database.AuthorizedNetworksMap],
    authorizedNetworksList: Array<string>,
    availabilityType: Database.AvailabilityTypeMap[keyof Database.AvailabilityTypeMap],
  }

  export interface EncryptionTypeMap {
    ENCRYPTION_UNKNOWN: 0;
    INSECURE_CLEAR_TEXT: 1;
    ENCRYPTION_MANAGED: 2;
    ENCRYPTION_USER_MANAGED: 3;
  }

  export const EncryptionType: EncryptionTypeMap;

  export interface BackupConfigurationMap {
    BACKUP_CONFIG_UNKNOWN: 0;
    BACKUP_CONFIG_DISABLED: 1;
    BACKUP_CONFIG_MANAGED: 2;
  }

  export const BackupConfiguration: BackupConfigurationMap;

  export interface PasswordPolicyMap {
    PASSWORD_POLICY_UNKNOWN: 0;
    PASSWORD_POLICY_WEAK: 1;
    PASSWORD_POLICY_STRONG: 2;
  }

  export const PasswordPolicy: PasswordPolicyMap;

  export interface AuthorizedNetworksMap {
    AUTHORIZED_NETWORKS_UNKNOWN: 0;
    AUTHORIZED_NETWORKS_NOT_SET: 1;
    AUTHORIZED_NETWORKS_SET: 2;
  }

  export const AuthorizedNetworks: AuthorizedNetworksMap;

  export interface AvailabilityTypeMap {
    HA_UNKNOWN: 0;
    NO_HA: 1;
    HA_ZONAL: 2;
    HA_REGIONAL: 3;
    HA_GLOBAL: 4;
  }

  export const AvailabilityType: AvailabilityTypeMap;
}

export class Bucket extends jspb.Message {
  hasCreationDate(): boolean;
  clearCreationDate(): void;
  getCreationDate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setCreationDate(value?: google_protobuf_timestamp_pb.Timestamp): void;

  hasRetentionPolicy(): boolean;
  clearRetentionPolicy(): void;
  getRetentionPolicy(): Bucket.RetentionPolicy | undefined;
  setRetentionPolicy(value?: Bucket.RetentionPolicy): void;

  hasEncryptionPolicy(): boolean;
  clearEncryptionPolicy(): void;
  getEncryptionPolicy(): Bucket.EncryptionPolicy | undefined;
  setEncryptionPolicy(value?: Bucket.EncryptionPolicy): void;

  getAccessType(): Bucket.AccessTypeMap[keyof Bucket.AccessTypeMap];
  setAccessType(value: Bucket.AccessTypeMap[keyof Bucket.AccessTypeMap]): void;

  getAccessControlType(): Bucket.AccessControlTypeMap[keyof Bucket.AccessControlTypeMap];
  setAccessControlType(value: Bucket.AccessControlTypeMap[keyof Bucket.AccessControlTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Bucket.AsObject;
  static toObject(includeInstance: boolean, msg: Bucket): Bucket.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Bucket, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Bucket;
  static deserializeBinaryFromReader(message: Bucket, reader: jspb.BinaryReader): Bucket;
}

export namespace Bucket {
  export type AsObject = {
    creationDate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    retentionPolicy?: Bucket.RetentionPolicy.AsObject,
    encryptionPolicy?: Bucket.EncryptionPolicy.AsObject,
    accessType: Bucket.AccessTypeMap[keyof Bucket.AccessTypeMap],
    accessControlType: Bucket.AccessControlTypeMap[keyof Bucket.AccessControlTypeMap],
  }

  export class RetentionPolicy extends jspb.Message {
    hasPeriod(): boolean;
    clearPeriod(): void;
    getPeriod(): google_protobuf_duration_pb.Duration | undefined;
    setPeriod(value?: google_protobuf_duration_pb.Duration): void;

    getIsLocked(): boolean;
    setIsLocked(value: boolean): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RetentionPolicy.AsObject;
    static toObject(includeInstance: boolean, msg: RetentionPolicy): RetentionPolicy.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RetentionPolicy, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RetentionPolicy;
    static deserializeBinaryFromReader(message: RetentionPolicy, reader: jspb.BinaryReader): RetentionPolicy;
  }

  export namespace RetentionPolicy {
    export type AsObject = {
      period?: google_protobuf_duration_pb.Duration.AsObject,
      isLocked: boolean,
    }
  }

  export class EncryptionPolicy extends jspb.Message {
    getIsEnabled(): boolean;
    setIsEnabled(value: boolean): void;

    getIsKeyCustomerManaged(): boolean;
    setIsKeyCustomerManaged(value: boolean): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EncryptionPolicy.AsObject;
    static toObject(includeInstance: boolean, msg: EncryptionPolicy): EncryptionPolicy.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EncryptionPolicy, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EncryptionPolicy;
    static deserializeBinaryFromReader(message: EncryptionPolicy, reader: jspb.BinaryReader): EncryptionPolicy;
  }

  export namespace EncryptionPolicy {
    export type AsObject = {
      isEnabled: boolean,
      isKeyCustomerManaged: boolean,
    }
  }

  export interface AccessTypeMap {
    ACCESS_UNKNOWN: 0;
    PRIVATE: 1;
    PUBLIC: 2;
  }

  export const AccessType: AccessTypeMap;

  export interface AccessControlTypeMap {
    ACCESS_CONTROL_UNKNOWN: 0;
    NON_UNIFORM: 1;
    UNIFORM: 2;
  }

  export const AccessControlType: AccessControlTypeMap;
}

export class APIKey extends jspb.Message {
  clearScopesList(): void;
  getScopesList(): Array<string>;
  setScopesList(value: Array<string>): void;
  addScopes(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): APIKey.AsObject;
  static toObject(includeInstance: boolean, msg: APIKey): APIKey.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: APIKey, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): APIKey;
  static deserializeBinaryFromReader(message: APIKey, reader: jspb.BinaryReader): APIKey;
}

export namespace APIKey {
  export type AsObject = {
    scopesList: Array<string>,
  }
}

export class Permission extends jspb.Message {
  getRole(): string;
  setRole(value: string): void;

  clearPrincipalsList(): void;
  getPrincipalsList(): Array<string>;
  setPrincipalsList(value: Array<string>): void;
  addPrincipals(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Permission.AsObject;
  static toObject(includeInstance: boolean, msg: Permission): Permission.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Permission, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Permission;
  static deserializeBinaryFromReader(message: Permission, reader: jspb.BinaryReader): Permission;
}

export namespace Permission {
  export type AsObject = {
    role: string,
    principalsList: Array<string>,
  }
}

export class IamPolicy extends jspb.Message {
  hasResource(): boolean;
  clearResource(): void;
  getResource(): Resource | undefined;
  setResource(value?: Resource): void;

  clearPermissionsList(): void;
  getPermissionsList(): Array<Permission>;
  setPermissionsList(value: Array<Permission>): void;
  addPermissions(value?: Permission, index?: number): Permission;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IamPolicy.AsObject;
  static toObject(includeInstance: boolean, msg: IamPolicy): IamPolicy.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: IamPolicy, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IamPolicy;
  static deserializeBinaryFromReader(message: IamPolicy, reader: jspb.BinaryReader): IamPolicy;
}

export namespace IamPolicy {
  export type AsObject = {
    resource?: Resource.AsObject,
    permissionsList: Array<Permission.AsObject>,
  }
}

export class ServiceAccount extends jspb.Message {
  clearExportedCredentialsList(): void;
  getExportedCredentialsList(): Array<ExportedCredentials>;
  setExportedCredentialsList(value: Array<ExportedCredentials>): void;
  addExportedCredentials(value?: ExportedCredentials, index?: number): ExportedCredentials;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServiceAccount.AsObject;
  static toObject(includeInstance: boolean, msg: ServiceAccount): ServiceAccount.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ServiceAccount, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServiceAccount;
  static deserializeBinaryFromReader(message: ServiceAccount, reader: jspb.BinaryReader): ServiceAccount;
}

export namespace ServiceAccount {
  export type AsObject = {
    exportedCredentialsList: Array<ExportedCredentials.AsObject>,
  }
}

export class ResourceGroup extends jspb.Message {
  getEnvironment(): string;
  setEnvironment(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResourceGroup.AsObject;
  static toObject(includeInstance: boolean, msg: ResourceGroup): ResourceGroup.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResourceGroup, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResourceGroup;
  static deserializeBinaryFromReader(message: ResourceGroup, reader: jspb.BinaryReader): ResourceGroup;
}

export namespace ResourceGroup {
  export type AsObject = {
    environment: string,
  }
}

export class LoadBalancer extends jspb.Message {
  getType(): LoadBalancer.TypeMap[keyof LoadBalancer.TypeMap];
  setType(value: LoadBalancer.TypeMap[keyof LoadBalancer.TypeMap]): void;

  clearCertificatesList(): void;
  getCertificatesList(): Array<Certificate>;
  setCertificatesList(value: Array<Certificate>): void;
  addCertificates(value?: Certificate, index?: number): Certificate;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadBalancer.AsObject;
  static toObject(includeInstance: boolean, msg: LoadBalancer): LoadBalancer.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: LoadBalancer, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadBalancer;
  static deserializeBinaryFromReader(message: LoadBalancer, reader: jspb.BinaryReader): LoadBalancer;
}

export namespace LoadBalancer {
  export type AsObject = {
    type: LoadBalancer.TypeMap[keyof LoadBalancer.TypeMap],
    certificatesList: Array<Certificate.AsObject>,
  }

  export interface TypeMap {
    UNKNOWN_TYPE: 0;
    EXTERNAL: 1;
    INTERNAL: 2;
  }

  export const Type: TypeMap;
}

export class Certificate extends jspb.Message {
  getType(): Certificate.TypeMap[keyof Certificate.TypeMap];
  setType(value: Certificate.TypeMap[keyof Certificate.TypeMap]): void;

  getDomainName(): string;
  setDomainName(value: string): void;

  clearSubjectAlternativeNamesList(): void;
  getSubjectAlternativeNamesList(): Array<string>;
  setSubjectAlternativeNamesList(value: Array<string>): void;
  addSubjectAlternativeNames(value: string, index?: number): string;

  hasCreationDate(): boolean;
  clearCreationDate(): void;
  getCreationDate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setCreationDate(value?: google_protobuf_timestamp_pb.Timestamp): void;

  hasExpirationDate(): boolean;
  clearExpirationDate(): void;
  getExpirationDate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setExpirationDate(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getIssuer(): string;
  setIssuer(value: string): void;

  getSignatureAlgorithm(): string;
  setSignatureAlgorithm(value: string): void;

  getPemCertificateChain(): string;
  setPemCertificateChain(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Certificate.AsObject;
  static toObject(includeInstance: boolean, msg: Certificate): Certificate.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Certificate, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Certificate;
  static deserializeBinaryFromReader(message: Certificate, reader: jspb.BinaryReader): Certificate;
}

export namespace Certificate {
  export type AsObject = {
    type: Certificate.TypeMap[keyof Certificate.TypeMap],
    domainName: string,
    subjectAlternativeNamesList: Array<string>,
    creationDate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    expirationDate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    issuer: string,
    signatureAlgorithm: string,
    pemCertificateChain: string,
  }

  export interface TypeMap {
    UNKNOWN: 0;
    IMPORTED: 1;
    MANAGED: 2;
  }

  export const Type: TypeMap;
}

export class Resource extends jspb.Message {
  getUid(): string;
  setUid(value: string): void;

  getCollectionUid(): string;
  setCollectionUid(value: string): void;

  hasTimestamp(): boolean;
  clearTimestamp(): void;
  getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getDisplayName(): string;
  setDisplayName(value: string): void;

  getLink(): string;
  setLink(value: string): void;

  getName(): string;
  setName(value: string): void;

  getParent(): string;
  setParent(value: string): void;

  getResourceGroupName(): string;
  setResourceGroupName(value: string): void;

  hasIamPolicy(): boolean;
  clearIamPolicy(): void;
  getIamPolicy(): IamPolicy | undefined;
  setIamPolicy(value?: IamPolicy): void;

  hasVmInstance(): boolean;
  clearVmInstance(): void;
  getVmInstance(): VmInstance | undefined;
  setVmInstance(value?: VmInstance): void;

  hasNetwork(): boolean;
  clearNetwork(): void;
  getNetwork(): Network | undefined;
  setNetwork(value?: Network): void;

  hasKubernetesCluster(): boolean;
  clearKubernetesCluster(): void;
  getKubernetesCluster(): KubernetesCluster | undefined;
  setKubernetesCluster(value?: KubernetesCluster): void;

  hasServiceAccount(): boolean;
  clearServiceAccount(): void;
  getServiceAccount(): ServiceAccount | undefined;
  setServiceAccount(value?: ServiceAccount): void;

  hasLoadBalancer(): boolean;
  clearLoadBalancer(): void;
  getLoadBalancer(): LoadBalancer | undefined;
  setLoadBalancer(value?: LoadBalancer): void;

  hasResourceGroup(): boolean;
  clearResourceGroup(): void;
  getResourceGroup(): ResourceGroup | undefined;
  setResourceGroup(value?: ResourceGroup): void;

  hasExportedCredentials(): boolean;
  clearExportedCredentials(): void;
  getExportedCredentials(): ExportedCredentials | undefined;
  setExportedCredentials(value?: ExportedCredentials): void;

  hasApiKey(): boolean;
  clearApiKey(): void;
  getApiKey(): APIKey | undefined;
  setApiKey(value?: APIKey): void;

  hasBucket(): boolean;
  clearBucket(): void;
  getBucket(): Bucket | undefined;
  setBucket(value?: Bucket): void;

  hasCertificate(): boolean;
  clearCertificate(): void;
  getCertificate(): Certificate | undefined;
  setCertificate(value?: Certificate): void;

  hasDatabase(): boolean;
  clearDatabase(): void;
  getDatabase(): Database | undefined;
  setDatabase(value?: Database): void;

  getTypeCase(): Resource.TypeCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Resource.AsObject;
  static toObject(includeInstance: boolean, msg: Resource): Resource.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Resource, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Resource;
  static deserializeBinaryFromReader(message: Resource, reader: jspb.BinaryReader): Resource;
}

export namespace Resource {
  export type AsObject = {
    uid: string,
    collectionUid: string,
    timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    displayName: string,
    link: string,
    name: string,
    parent: string,
    resourceGroupName: string,
    iamPolicy?: IamPolicy.AsObject,
    vmInstance?: VmInstance.AsObject,
    network?: Network.AsObject,
    kubernetesCluster?: KubernetesCluster.AsObject,
    serviceAccount?: ServiceAccount.AsObject,
    loadBalancer?: LoadBalancer.AsObject,
    resourceGroup?: ResourceGroup.AsObject,
    exportedCredentials?: ExportedCredentials.AsObject,
    apiKey?: APIKey.AsObject,
    bucket?: Bucket.AsObject,
    certificate?: Certificate.AsObject,
    database?: Database.AsObject,
  }

  export enum TypeCase {
    TYPE_NOT_SET = 0,
    VM_INSTANCE = 100,
    NETWORK = 101,
    KUBERNETES_CLUSTER = 102,
    SERVICE_ACCOUNT = 103,
    LOAD_BALANCER = 104,
    RESOURCE_GROUP = 105,
    EXPORTED_CREDENTIALS = 106,
    API_KEY = 107,
    BUCKET = 108,
    CERTIFICATE = 109,
    DATABASE = 110,
  }
}

export class Remediation extends jspb.Message {
  getDescription(): string;
  setDescription(value: string): void;

  getRecommendation(): string;
  setRecommendation(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Remediation.AsObject;
  static toObject(includeInstance: boolean, msg: Remediation): Remediation.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Remediation, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Remediation;
  static deserializeBinaryFromReader(message: Remediation, reader: jspb.BinaryReader): Remediation;
}

export namespace Remediation {
  export type AsObject = {
    description: string,
    recommendation: string,
  }
}

export class Observation extends jspb.Message {
  getUid(): string;
  setUid(value: string): void;

  getScanUid(): string;
  setScanUid(value: string): void;

  hasTimestamp(): boolean;
  clearTimestamp(): void;
  getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): void;

  hasResource(): boolean;
  clearResource(): void;
  getResource(): Resource | undefined;
  setResource(value?: Resource): void;

  getName(): string;
  setName(value: string): void;

  hasExpectedValue(): boolean;
  clearExpectedValue(): void;
  getExpectedValue(): google_protobuf_struct_pb.Value | undefined;
  setExpectedValue(value?: google_protobuf_struct_pb.Value): void;

  hasObservedValue(): boolean;
  clearObservedValue(): void;
  getObservedValue(): google_protobuf_struct_pb.Value | undefined;
  setObservedValue(value?: google_protobuf_struct_pb.Value): void;

  hasRemediation(): boolean;
  clearRemediation(): void;
  getRemediation(): Remediation | undefined;
  setRemediation(value?: Remediation): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Observation.AsObject;
  static toObject(includeInstance: boolean, msg: Observation): Observation.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Observation, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Observation;
  static deserializeBinaryFromReader(message: Observation, reader: jspb.BinaryReader): Observation;
}

export namespace Observation {
  export type AsObject = {
    uid: string,
    scanUid: string,
    timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    resource?: Resource.AsObject,
    name: string,
    expectedValue?: google_protobuf_struct_pb.Value.AsObject,
    observedValue?: google_protobuf_struct_pb.Value.AsObject,
    remediation?: Remediation.AsObject,
  }
}

export class ScanResultsList extends jspb.Message {
  clearObservationsList(): void;
  getObservationsList(): Array<Observation>;
  setObservationsList(value: Array<Observation>): void;
  addObservations(value?: Observation, index?: number): Observation;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ScanResultsList.AsObject;
  static toObject(includeInstance: boolean, msg: ScanResultsList): ScanResultsList.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ScanResultsList, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ScanResultsList;
  static deserializeBinaryFromReader(message: ScanResultsList, reader: jspb.BinaryReader): ScanResultsList;
}

export namespace ScanResultsList {
  export type AsObject = {
    observationsList: Array<Observation.AsObject>,
  }
}

export class GetStatusCollectAndScanResponse extends jspb.Message {
  getCollectStatus(): RequestStatusMap[keyof RequestStatusMap];
  setCollectStatus(value: RequestStatusMap[keyof RequestStatusMap]): void;

  getScanStatus(): RequestStatusMap[keyof RequestStatusMap];
  setScanStatus(value: RequestStatusMap[keyof RequestStatusMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetStatusCollectAndScanResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetStatusCollectAndScanResponse): GetStatusCollectAndScanResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetStatusCollectAndScanResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetStatusCollectAndScanResponse;
  static deserializeBinaryFromReader(message: GetStatusCollectAndScanResponse, reader: jspb.BinaryReader): GetStatusCollectAndScanResponse;
}

export namespace GetStatusCollectAndScanResponse {
  export type AsObject = {
    collectStatus: RequestStatusMap[keyof RequestStatusMap],
    scanStatus: RequestStatusMap[keyof RequestStatusMap],
  }
}

export class GetStatusCollectAndScanRequest extends jspb.Message {
  getCollectId(): string;
  setCollectId(value: string): void;

  getScanId(): string;
  setScanId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetStatusCollectAndScanRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetStatusCollectAndScanRequest): GetStatusCollectAndScanRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetStatusCollectAndScanRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetStatusCollectAndScanRequest;
  static deserializeBinaryFromReader(message: GetStatusCollectAndScanRequest, reader: jspb.BinaryReader): GetStatusCollectAndScanRequest;
}

export namespace GetStatusCollectAndScanRequest {
  export type AsObject = {
    collectId: string,
    scanId: string,
  }
}

export class CollectAndScanRequest extends jspb.Message {
  clearResourceGroupNamesList(): void;
  getResourceGroupNamesList(): Array<string>;
  setResourceGroupNamesList(value: Array<string>): void;
  addResourceGroupNames(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CollectAndScanRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CollectAndScanRequest): CollectAndScanRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CollectAndScanRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CollectAndScanRequest;
  static deserializeBinaryFromReader(message: CollectAndScanRequest, reader: jspb.BinaryReader): CollectAndScanRequest;
}

export namespace CollectAndScanRequest {
  export type AsObject = {
    resourceGroupNamesList: Array<string>,
  }
}

export class CollectAndScanResponse extends jspb.Message {
  getCollectId(): string;
  setCollectId(value: string): void;

  getScanId(): string;
  setScanId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CollectAndScanResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CollectAndScanResponse): CollectAndScanResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CollectAndScanResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CollectAndScanResponse;
  static deserializeBinaryFromReader(message: CollectAndScanResponse, reader: jspb.BinaryReader): CollectAndScanResponse;
}

export namespace CollectAndScanResponse {
  export type AsObject = {
    collectId: string,
    scanId: string,
  }
}

export class ListObservationsRequest extends jspb.Message {
  getPageToken(): string;
  setPageToken(value: string): void;

  getPageSize(): number;
  setPageSize(value: number): void;

  clearResourceGroupNamesList(): void;
  getResourceGroupNamesList(): Array<string>;
  setResourceGroupNamesList(value: Array<string>): void;
  addResourceGroupNames(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListObservationsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListObservationsRequest): ListObservationsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListObservationsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListObservationsRequest;
  static deserializeBinaryFromReader(message: ListObservationsRequest, reader: jspb.BinaryReader): ListObservationsRequest;
}

export namespace ListObservationsRequest {
  export type AsObject = {
    pageToken: string,
    pageSize: number,
    resourceGroupNamesList: Array<string>,
  }
}

export class CreateObservationRequest extends jspb.Message {
  hasObservation(): boolean;
  clearObservation(): void;
  getObservation(): Observation | undefined;
  setObservation(value?: Observation): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateObservationRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateObservationRequest): CreateObservationRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateObservationRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateObservationRequest;
  static deserializeBinaryFromReader(message: CreateObservationRequest, reader: jspb.BinaryReader): CreateObservationRequest;
}

export namespace CreateObservationRequest {
  export type AsObject = {
    observation?: Observation.AsObject,
  }
}

export class RuleObservationPair extends jspb.Message {
  getRule(): string;
  setRule(value: string): void;

  clearObservationsList(): void;
  getObservationsList(): Array<Observation>;
  setObservationsList(value: Array<Observation>): void;
  addObservations(value?: Observation, index?: number): Observation;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RuleObservationPair.AsObject;
  static toObject(includeInstance: boolean, msg: RuleObservationPair): RuleObservationPair.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RuleObservationPair, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RuleObservationPair;
  static deserializeBinaryFromReader(message: RuleObservationPair, reader: jspb.BinaryReader): RuleObservationPair;
}

export namespace RuleObservationPair {
  export type AsObject = {
    rule: string,
    observationsList: Array<Observation.AsObject>,
  }
}

export class ResourceGroupObservationsPair extends jspb.Message {
  getResourceGroupName(): string;
  setResourceGroupName(value: string): void;

  clearRulesObservationsList(): void;
  getRulesObservationsList(): Array<RuleObservationPair>;
  setRulesObservationsList(value: Array<RuleObservationPair>): void;
  addRulesObservations(value?: RuleObservationPair, index?: number): RuleObservationPair;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResourceGroupObservationsPair.AsObject;
  static toObject(includeInstance: boolean, msg: ResourceGroupObservationsPair): ResourceGroupObservationsPair.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResourceGroupObservationsPair, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResourceGroupObservationsPair;
  static deserializeBinaryFromReader(message: ResourceGroupObservationsPair, reader: jspb.BinaryReader): ResourceGroupObservationsPair;
}

export namespace ResourceGroupObservationsPair {
  export type AsObject = {
    resourceGroupName: string,
    rulesObservationsList: Array<RuleObservationPair.AsObject>,
  }
}

export class ListObservationsResponse extends jspb.Message {
  clearResourceGroupsObservationsList(): void;
  getResourceGroupsObservationsList(): Array<ResourceGroupObservationsPair>;
  setResourceGroupsObservationsList(value: Array<ResourceGroupObservationsPair>): void;
  addResourceGroupsObservations(value?: ResourceGroupObservationsPair, index?: number): ResourceGroupObservationsPair;

  getNextPageToken(): string;
  setNextPageToken(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListObservationsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListObservationsResponse): ListObservationsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListObservationsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListObservationsResponse;
  static deserializeBinaryFromReader(message: ListObservationsResponse, reader: jspb.BinaryReader): ListObservationsResponse;
}

export namespace ListObservationsResponse {
  export type AsObject = {
    resourceGroupsObservationsList: Array<ResourceGroupObservationsPair.AsObject>,
    nextPageToken: string,
  }
}

export interface RequestStatusMap {
  UNKNOWN: 0;
  DONE: 1;
  RUNNING: 2;
  ALREADY_RUNNING: 3;
  CANCELLED: 4;
}

export const RequestStatus: RequestStatusMap;

