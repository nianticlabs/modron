import { ComponentFixture, TestBed } from "@angular/core/testing"
import { ActivatedRoute } from "@angular/router"
import { reverseSortPipe } from "../filter.pipe"
import {
  MapByTypePipe,
  mapFlatRulesPipe,
} from "../resource-group-details/resource-group-details.pipe"
import {
  InvalidProjectNb,
  ObservationsPipe,
} from "../resource-groups/resource-groups.pipe"
import { ModronStore } from "../state/modron.store"

import { StatsComponent } from "./stats.component"

describe("StatsComponent", () => {
  let component: StatsComponent
  let fixture: ComponentFixture<StatsComponent>

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [
        StatsComponent,
        InvalidProjectNb,
        ObservationsPipe,
        mapFlatRulesPipe,
        reverseSortPipe,
        MapByTypePipe,
      ],
      providers: [
        ModronStore,
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get(): string {
                  return "mock-observation-id"
                },
              },
            },
          },
        },
      ],
    }).compileComponents()

    fixture = TestBed.createComponent(StatsComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it("should create", () => {
    expect(component).toBeTruthy()
  })
})
