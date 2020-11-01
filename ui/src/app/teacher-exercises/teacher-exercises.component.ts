import { Component, OnInit } from '@angular/core';
import { Exercise, languages } from '../entity';
import { MyCodeService } from '../my-code.service';
import { MatDialog } from '@angular/material/dialog';
import { NewExerciseDialogComponent } from '../new-exercise-dialog/new-exercise-dialog.component';
import { EditExerciseDialogComponent } from '../edit-exercise-dialog/edit-exercise-dialog.component';
import { ConfirmDialogComponent } from '../confirm-dialog/confirm-dialog.component';

@Component({
  selector: 'app-teacher-exercises',
  templateUrl: './teacher-exercises.component.html',
  styleUrls: ['./teacher-exercises.component.css']
})
export class TeacherExercisesComponent implements OnInit {

  exercises: { [key: string]: Exercise[] } = {};

  languageNames: { [id: number]: string };

  constructor(
    private myCodeService: MyCodeService,
    private dialog: MatDialog
  ) { }

  ngOnInit(): void {
    this.languageNames = {};
    languages.forEach(i => {
      this.languageNames[i.id] = i.name;
    });

    this.myCodeService.getExercises({}).subscribe(resp => {
      resp.exercises.forEach(e => {
        if (this.exercises[e.language]) {
          this.exercises[e.language].push(e);
        } else {
          this.exercises[e.language] = [e];
        }
      });
    });
  }

  newExercise(): void {
    const dialogRef = this.dialog.open(NewExerciseDialogComponent);
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        if (this.exercises[res.language]) {
          this.exercises[res.language] = [...this.exercises[res.language], res];
        } else {
          this.exercises[res.language] = [res];
        }
      }
    });
  }

  edit(e: Exercise): void {
    const dialogRef = this.dialog.open(EditExerciseDialogComponent, { data: e });
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        this.exercises[res.language] = this.exercises[res.language].map(ee => {
          if (ee.id === res.id) {
            return res;
          }
          return ee;
        });
      }
    });
  }

  remove(e: Exercise): void {

    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      data: {
        message: 'Вы действительно хотите удалить задачу?',
        description: 'Это приведёт к удалению всех его тестов, а так-же всех решений, которые прислали ученики. <b>Удаление необратимо!</b>'
      },
    });
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        this.myCodeService.removeExercise({exercise_id: e.id}).subscribe(() => {
          this.exercises[e.language] = this.exercises[e.language].filter((ee => e.id !== ee.id));
          if (!this.exercises[e.language].length) {
            delete this.exercises[e.language];
            this.exercises = { ...this.exercises };
          }
        });
      }
    });
  }

}
