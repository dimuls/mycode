import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { IndexComponent } from './index/index.component';
import { LoginComponent } from './login/login.component';
import { TeacherSolutionsComponent } from './teacher-solutions/teacher-solutions.component';
import { TeacherExercisesComponent } from './teacher-exercises/teacher-exercises.component';
import { StudentExercisesComponent } from './student-exercises/student-exercises.component';
import { studentUR, teacherUR } from './entity';
import { AuthGuard } from './auth.guard';
import { NotFoundComponent } from './not-found/not-found.component';
import { TeacherExerciseTestsComponent } from './teacher-exercise-tests/teacher-exercise-tests.component';
import { TeacherExerciseAssignmentsComponent } from './teacher-exercise-assignments/teacher-exercise-assignments.component';

const routes: Routes = [
  { path: '', component: IndexComponent },
  { path: 'login', component: LoginComponent },
  { path: 'teacher', canActivate: [AuthGuard], data: { roles: [teacherUR] }, children: [
    { path: 'solutions', component: TeacherSolutionsComponent },
    { path: 'exercises', component: TeacherExercisesComponent },
    { path: 'exercise/:exercise_id/tests', component: TeacherExerciseTestsComponent },
    { path: 'exercise/:exercise_id/assignments', component: TeacherExerciseAssignmentsComponent }
  ] },
  { path: 'student', canActivate: [AuthGuard], data: { roles: [studentUR] }, children: [
    { path: 'exercises', component: StudentExercisesComponent }
  ] },
  { path: 'not-found', component: NotFoundComponent },
  { path: '**', redirectTo: 'not-found' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
