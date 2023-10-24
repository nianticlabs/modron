import { ComponentFixture, TestBed } from "@angular/core/testing"

import { ModronAppComponent } from "./modron-app.component"

describe("ModronAppComponent", () => {
  let component: ModronAppComponent
  let fixture: ComponentFixture<ModronAppComponent>

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ModronAppComponent],
    }).compileComponents()

    fixture = TestBed.createComponent(ModronAppComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it("should create", () => {
    expect(component).toBeTruthy()
  })
})
