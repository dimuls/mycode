import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TeacherExercisesComponent } from './teacher-exercises.component';

describe('TeacherExercisesComponent', () => {
  let component: TeacherExercisesComponent;
  let fixture: ComponentFixture<TeacherExercisesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TeacherExercisesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TeacherExercisesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
