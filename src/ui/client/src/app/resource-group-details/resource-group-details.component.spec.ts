import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { reverseSortPipe } from '../filter.pipe';
import { ObservationsPipe } from '../resource-groups/resource-groups.pipe';
import { AuthenticationStore } from '../state/authentication.store';
import { ModronStore } from '../state/modron.store';
import { NotificationStore } from '../state/notification.store';

import { ResourceGroupDetailsComponent } from './resource-group-details.component';

describe('ResourceGroupDetailsComponent', () => {
  let component: ResourceGroupDetailsComponent;
  let fixture: ComponentFixture<ResourceGroupDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [
        ResourceGroupDetailsComponent,
        ObservationsPipe,
        reverseSortPipe,
      ],
      providers: [
        ModronStore,
        NotificationStore,
        AuthenticationStore,
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get(): string {
                  return 'mock-observation-id';
                },
              },
            },
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ResourceGroupDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
