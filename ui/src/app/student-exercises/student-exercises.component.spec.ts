import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StudentExercisesComponent } from './student-exercises.component';

describe('StudentExercisesComponent', () => {
  let component: StudentExercisesComponent;
  let fixture: ComponentFixture<StudentExercisesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ StudentExercisesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(StudentExercisesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
