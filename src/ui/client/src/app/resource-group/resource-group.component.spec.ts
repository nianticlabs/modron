import { ComponentFixture, TestBed } from "@angular/core/testing"
import { ModronStore } from "../state/modron.store"
import { ResourceGroupComponent } from "./resource-group.component"
import { MatSnackBarModule } from "@angular/material/snack-bar"

describe("ResourceGroupComponent", () => {
  let component: ResourceGroupComponent
  let fixture: ComponentFixture<ResourceGroupComponent>

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ResourceGroupComponent],
      imports: [MatSnackBarModule],
      providers: [ModronStore],
    }).compileComponents()

    fixture = TestBed.createComponent(ResourceGroupComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it("should create", () => {
    expect(component).toBeTruthy()
  })
})
