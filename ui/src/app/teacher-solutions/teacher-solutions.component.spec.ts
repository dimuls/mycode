import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TeacherSolutionsComponent } from './teacher-solutions.component';

describe('TeacherStudentsComponent', () => {
  let component: TeacherSolutionsComponent;
  let fixture: ComponentFixture<TeacherSolutionsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TeacherSolutionsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TeacherSolutionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
