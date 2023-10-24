INSERT INTO operations(operationID, resourceGroupName, opsType, startTime, status) VALUES("uid1", "project-id", "COLLECT", date("now", "-7 days"), "STARTED");
UPDATE operations SET status = "COMPLETED", endTime = date("now", "-7 days") WHERE operationID = "uid1";
INSERT INTO resources(resourceID, resourceName, resourceGroupName, collectionID, recordTime, parentName, resourceType) VALUES ("resourceUid1", "resourceName", "project-id", "uid1", date("now", "-7 days"), "parent-name", "resource type");

INSERT INTO operations(operationID, resourceGroupName, opsType, startTime, status) VALUES("uid2", "project-id", "COLLECT", date("now"), "STARTED");
UPDATE operations SET status = "COMPLETED", endTime = date("now") WHERE operationID = "uid2";
INSERT INTO resources(resourceID, resourceName, resourceGroupName, collectionID, recordTime, parentName, resourceType) VALUES ("resourceUid2", "resourceName", "project-id", "uid2", date("now"), "parent-name", "resource type");
