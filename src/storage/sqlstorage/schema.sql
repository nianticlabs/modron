CREATE TABLE resources (
    resourceID VARCHAR(50) NOT NULL,
    resourceName VARCHAR(2000),
    resourceGroupName VARCHAR(500),
    collectionID VARCHAR(50),
    recordTime TIMESTAMP,
    parentName VARCHAR(500),
    resourceType VARCHAR(100),
    resourceProto BYTEA,
    PRIMARY KEY (resourceID)
);

CREATE TABLE observations (
    observationID VARCHAR(50) NOT NULL,
    observationName VARCHAR(2000),
    resourceGroupName VARCHAR(500),
    resourceID VARCHAR(50),
    scanID varchar(50),
    recordTime TIMESTAMP,
    observationProto BYTEA,
    PRIMARY KEY (observationID)
);

CREATE TABLE operations (
    operationID VARCHAR(50) NOT NULL,
    resourceGroupName VARCHAR(500),
    opsType VARCHAR(50),
    startTime TIMESTAMP,
    endTime TIMESTAMP,
    status VARCHAR(50),
    reason VARCHAR(1000)
);
CREATE INDEX resources_name ON resources (resourcename);
CREATE INDEX resources_resourcegroupname ON resources (resourcegroupname);
CREATE INDEX observations_resourcegroupname ON observations (resourcegroupname);
