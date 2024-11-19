import {Pipe, PipeTransform} from "@angular/core";
import {Observation} from "../../proto/modron_pb";
import Category = Observation.Category;

@Pipe({
  name: "categoryName"
})
export class CategoryNamePipe implements PipeTransform {
  transform(cat: Category): string {
    switch(cat) {
      case Category.CATEGORY_VULNERABILITY:
        return "Vulnerability";
      case Category.CATEGORY_MISCONFIGURATION:
        return "Misconfiguration";
      case Category.CATEGORY_TOXIC_COMBINATION:
        return "Toxic Combination";
      default:
        return "Unknown";
    }
  }
}
