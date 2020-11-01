import { TestBed } from '@angular/core/testing';

import { MyCodeService } from './my-code.service';

describe('MyCodeService', () => {
  let service: MyCodeService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MyCodeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
