import { TestBed } from '@angular/core/testing';
import { ModronService } from './modron.service';

describe('ModronService', () => {
  let service: ModronService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ModronService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
