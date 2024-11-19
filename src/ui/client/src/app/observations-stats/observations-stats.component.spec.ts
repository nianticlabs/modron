import { ComponentFixture, TestBed } from "@angular/core/testing";

import { ObservationsStatsComponent } from "./observations-stats.component";

describe("HistogramHorizontalComponent", () => {
  let component: ObservationsStatsComponent;
  let fixture: ComponentFixture<ObservationsStatsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ObservationsStatsComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ObservationsStatsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
