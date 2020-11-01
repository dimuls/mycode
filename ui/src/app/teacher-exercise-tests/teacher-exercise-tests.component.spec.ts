import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TeacherExerciseTestsComponent } from './teacher-exercise-tests.component';

describe('TeacherExerciseTestsComponent', () => {
  let component: TeacherExerciseTestsComponent;
  let fixture: ComponentFixture<TeacherExerciseTestsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TeacherExerciseTestsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TeacherExerciseTestsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
