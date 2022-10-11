import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HistogramHorizontalComponent } from './histogram-horizontal.component';

describe('HistogramHorizontalComponent', () => {
  let component: HistogramHorizontalComponent;
  let fixture: ComponentFixture<HistogramHorizontalComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [HistogramHorizontalComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(HistogramHorizontalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
