import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { FilterKeyValuePipe, FilterNoObservationsPipe } from "../filter.pipe";
import { mapFlatRulesPipe } from "../resource-group-details/resource-group-details.pipe";
import { ModronStore } from "../state/modron.store";
import { RouterTestingModule } from "@angular/router/testing";
import { HttpClientTestingModule } from "@angular/common/http/testing";

import { ResourceGroupsComponent } from "./resource-groups.component";
import {
  InvalidProjectNb,
  ObsNbPipe,
  ResourceGroupsPipe,
} from "./resource-groups.pipe";

describe("ResourceGroupsComponent", () => {
  let component: ResourceGroupsComponent;
  let fixture: ComponentFixture<ResourceGroupsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [
        ResourceGroupsComponent,
        ResourceGroupsPipe,
        FilterKeyValuePipe,
        mapFlatRulesPipe,
        InvalidProjectNb,
        ObsNbPipe,
        FilterNoObservationsPipe,
      ],
      imports: [
        MatSnackBarModule,
        RouterTestingModule,
        HttpClientTestingModule,
        ],
      providers: [ModronStore],
    }).compileComponents();

    fixture = TestBed.createComponent(ResourceGroupsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
