import { KeyValue } from "@angular/common"
import { Pipe, PipeTransform } from "@angular/core"
import { Value } from "google-protobuf/google/protobuf/struct_pb"
import { Observation, ResourceRef } from "../proto/modron_pb"

@Pipe({
  name: "filterObs",
})
export class FilterObsPipe implements PipeTransform {
  transform(
    items: Observation[],
    resource: string,
    group: string,
    value: string
  ): Observation[] {
    if (!items) {
      return []
    }
    if (!resource && !group && !value) {
      return items
    }

    resource = resource.toLocaleLowerCase()
    group = group.toLocaleLowerCase()
    value = value.toLocaleLowerCase()

    return items.filter((it) => {
      return (
        (it.getResourceRef() as ResourceRef)
          .getExternalId()
          .toLocaleLowerCase()
          .includes(resource) &&
        (it.getResourceRef() as ResourceRef)
          .getGroupName().replace("projects/", "")
          .toLocaleLowerCase()
          .includes(group) &&
        (it.getObservedValue()
          ? (it.getObservedValue() as Value)
            .toString()
            .toLocaleLowerCase()
            .includes(value)
          : true)
      )
    })
  }
}

// Filter by group name
@Pipe({
  name: "filterKeyValue",
})
export class FilterKeyValuePipe implements PipeTransform {
  transform(items: any[], searchText: string): any[] {
    if (!items) {
      return []
    }
    if (!searchText) {
      return items
    }
    searchText = searchText.toLocaleLowerCase()

    return items.filter((it) => {
      return it.key.toLocaleLowerCase().includes(searchText)
    })
  }
}

// Filter pipe to remove all elements without observations
@Pipe({
  name: "filterNoObservations",
})
export class FilterNoObservationsPipe implements PipeTransform {
  transform(items: any[], removeNoObs: boolean): any[] {
    if (!items) {
      return []
    }
    if (!removeNoObs) {
      return items
    }

    return items.filter((it) => {
      return it.value.length != 0
    })
  }
}

@Pipe({
  name: "reverseSortByLength",
})
export class reverseSortPipe implements PipeTransform {
  transform(items: KeyValue<string, any[]>[]): KeyValue<string, any[]>[] {
    if (!items) {
      return []
    }
    return items.sort((a, b) => a.value.length - b.value.length).reverse()
  }
}

@Pipe({
  name: "shortenDescription",
})
export class ShortenDescriptionPipe implements PipeTransform {
  transform(value: string | undefined | null) {
    if(value === undefined || value === null) {
      return "";
    }
    return value.split("\n")[0]
  }
}

@Pipe({
  name: "parseExternalId",
})
export class ParseExternalIdPipe implements PipeTransform {
  transform(value: string | undefined | null) {
    if(value === undefined || value === null) {
      return "";
    }

    const regex = /^\/\/container\.googleapis\.com\/projects\/[^\\/]+\/locations\/[^\\/]+\/clusters\/[^\\/]+\/k8s\/namespaces\/[^\\/]+\/apps\/((?:deployments|daemonsets)\/[^\\/]+)$/;
    const matches = value.match(regex);
    if (matches) {
      return matches[1];
    }

    return value.split("/", -1).pop() || ""
  }
}

@Pipe({
  name: "structValueToString"
})
export class StructValueToStringPipe implements PipeTransform {
  transform(value: Value | null | undefined): string {
      if (value === undefined || value === null) {
        return ""
      }
      if(value.hasBoolValue()) {
        return value.getBoolValue().toString()
      }
      if(value.hasStringValue()) {
        return value.getStringValue()
      }
      if(value.hasNumberValue()) {
        return value.getNumberValue().toString()
      }
      return value.toString()
  }
}
