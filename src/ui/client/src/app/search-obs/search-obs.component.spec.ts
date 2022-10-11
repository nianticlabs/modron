import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FilterObsPipe } from '../filter.pipe';

import { SearchObsComponent } from './search-obs.component';

describe('SearchObsComponent', () => {
  let component: SearchObsComponent;
  let fixture: ComponentFixture<SearchObsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [SearchObsComponent, FilterObsPipe],
    }).compileComponents();

    fixture = TestBed.createComponent(SearchObsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
