import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { LoginComponent } from './login/login.component';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { TeacherExercisesComponent } from './teacher-exercises/teacher-exercises.component';
import { TeacherSolutionsComponent } from './teacher-solutions/teacher-solutions.component';
import { StudentExercisesComponent } from './student-exercises/student-exercises.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatCardModule } from '@angular/material/card';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { NotFoundComponent } from './not-found/not-found.component';
import { IndexComponent } from './index/index.component';
import { JWTInterceptor } from './jwt.interceptor';
import { ErrorInterceptor } from './error.interceptor';
import { MatListModule } from '@angular/material/list';
import { MatDividerModule } from '@angular/material/divider';
import { MatSelectModule } from '@angular/material/select';
import { NewExerciseDialogComponent } from './new-exercise-dialog/new-exercise-dialog.component';
import { MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { EditExerciseDialogComponent } from './edit-exercise-dialog/edit-exercise-dialog.component';
import { TeacherExerciseTestsComponent } from './teacher-exercise-tests/teacher-exercise-tests.component';
import { NewTestDialogComponent } from './new-test-dialog/new-test-dialog.component';
import { EditTestDialogComponent } from './edit-test-dialog/edit-test-dialog.component';
import { MatChipsModule } from '@angular/material/chips';
import { DragDropModule } from '@angular/cdk/drag-drop';
import { MatIconModule } from '@angular/material/icon';
import { ConfirmDialogComponent } from './confirm-dialog/confirm-dialog.component';
import { TeacherExerciseAssignmentsComponent } from './teacher-exercise-assignments/teacher-exercise-assignments.component';
import { HIGHLIGHT_OPTIONS, HighlightModule } from 'ngx-highlightjs';
import { MarkdownModule } from 'ngx-markdown';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    TeacherExercisesComponent,
    TeacherSolutionsComponent,
    StudentExercisesComponent,
    NotFoundComponent,
    IndexComponent,
    NewExerciseDialogComponent,
    EditExerciseDialogComponent,
    TeacherExerciseTestsComponent,
    NewTestDialogComponent,
    EditTestDialogComponent,
    ConfirmDialogComponent,
    TeacherExerciseAssignmentsComponent
  ],
  imports: [
    FormsModule,
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    HttpClientModule,
    ReactiveFormsModule,
    MatToolbarModule,
    MatCardModule,
    MatInputModule,
    MatButtonModule,
    MatListModule,
    MatDividerModule,
    MatSelectModule,
    MatDialogModule,
    MatExpansionModule,
    MatChipsModule,
    MatIconModule,
    DragDropModule,
    HighlightModule,
    MarkdownModule.forRoot()
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: JWTInterceptor, multi: true },
    { provide: HTTP_INTERCEPTORS, useClass: ErrorInterceptor, multi: true },
    {
      provide: HIGHLIGHT_OPTIONS,
      useValue: {
        lineNumbers: true,
        fullLibraryLoader: () => import('highlight.js'),
      }
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
