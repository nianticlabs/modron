import {ComponentFixture, TestBed} from "@angular/core/testing";
import {SidenavComponent} from "./sidenav.component";
import {Router} from "@angular/router";
import {RouterTestingModule} from "@angular/router/testing";
import {Component} from "@angular/core";

@Component({
  template: ""
})
class DummyComponent {
}

describe("SidenavComponent", () => {
  let component: SidenavComponent;
  let fixture: ComponentFixture<SidenavComponent>;
  let router: Router;

  beforeEach(async () => {
    const testingModule = TestBed.configureTestingModule({
      imports: [SidenavComponent, RouterTestingModule.withRoutes(
        [{path: "modron/resourcegroups", component: DummyComponent}]
      )],
      providers: []
    })
    await testingModule.compileComponents();

    fixture = TestBed.createComponent(SidenavComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should mark the icon as active when the route matches", async () => {
    await router.navigateByUrl("/modron/resourcegroups");
    fixture.detectChanges();
    const {debugElement} = fixture;
    const navItems = debugElement.nativeElement.querySelectorAll("div.nav-items div.nav-item")
    expect(navItems.length).toBe(3);
    expect(navItems[0].classList).toContain("active");
    expect(navItems[1].classList).not.toContain("active");
    expect(navItems[2].classList).not.toContain("active");
  });
});
