import { Pipe, PipeTransform } from "@angular/core"
import { NotificationException } from "../model/notification.model"

@Pipe({
  name: "filterExceptions",
})
export class NotificationExceptionsFilterPipe implements PipeTransform {
  transform(exps: NotificationException[], searchText: string): NotificationException[] {
    searchText = searchText.toLocaleLowerCase()

    return exps.filter((exp) => {
      return exp.notificationName.toLocaleLowerCase().includes(searchText)
    })
  }
}
