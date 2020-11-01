import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TeacherExerciseAssignmentsComponent } from './teacher-exercise-assignments.component';

describe('TeacherExerciseAssigmentsComponent', () => {
  let component: TeacherExerciseAssignmentsComponent;
  let fixture: ComponentFixture<TeacherExerciseAssignmentsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TeacherExerciseAssignmentsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TeacherExerciseAssignmentsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
