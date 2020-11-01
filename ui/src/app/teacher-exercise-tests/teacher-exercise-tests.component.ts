import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Exercise, Test, languages, testTypes, simpleTestType, checkerTestType } from '../entity';
import { MyCodeService } from '../my-code.service';
import { MatDialog } from '@angular/material/dialog';
import { NewTestDialogComponent } from '../new-test-dialog/new-test-dialog.component';
import { EditTestDialogComponent, formatDurationBack, formatMemoryBack } from '../edit-test-dialog/edit-test-dialog.component';
import { combineLatest } from 'rxjs';
import { ConfirmDialogComponent } from '../confirm-dialog/confirm-dialog.component';

@Component({
  selector: 'app-teacher-exercise-tests',
  templateUrl: './teacher-exercise-tests.component.html',
  styleUrls: ['./teacher-exercise-tests.component.css']
})
export class TeacherExerciseTestsComponent implements OnInit {

  exerciseID: number;
  exercise: Exercise;

  simpleTestType = simpleTestType;
  checkerTestType = checkerTestType;

  tests: Test[] = [];

  testTypesNames: { [id: string]: string }; 
  languageNames: { [id: string]: string };

  constructor(
    private activatedRoute: ActivatedRoute,
    private dialog: MatDialog,
    private myCodeService: MyCodeService
  ) {
    this.testTypesNames = {};
    testTypes.forEach(i => {
      this.testTypesNames[i.id] = i.name;
    });
    this.languageNames = {};
    languages.forEach(i => {
      this.languageNames[i.id] = i.name;
    });
  }

  ngOnInit(): void {
    this.activatedRoute.params.subscribe(p => {
      this.exerciseID = p.exercise_id;

      const getTests = this.myCodeService.getTests({ exercise_id: p.exercise_id });
      const getExercise = this.myCodeService.getExercise({ exercise_id: p.exercise_id });

      combineLatest([getTests, getExercise]).subscribe(([testsResp, exerciseResp]) => {
        this.tests = testsResp.tests;
        this.exercise = exerciseResp.exercise;
      });
    });
  }

  newTest(): void {
    const dialogRef = this.dialog.open(NewTestDialogComponent, {data: this.exerciseID});
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        this.tests = [...this.tests, res];
      }
    });
  }

  edit(t: Test): void {
    const dialogRef = this.dialog.open(EditTestDialogComponent, { data: t });
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        this.tests = this.tests.map(tt => {
          if (tt.id === res.id) {
            return res;
          }
          return tt;
        });
      }
    });
  }

  remove(t: Test): void {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      data: {
        message: 'Вы действительно хотите удалить тест?',
        description: 'Это приведёт к удалению всех проверок, которые были сделаны данным тестов. <b>Удаление необратимо!</b>'
      }
    });
    dialogRef.afterClosed().subscribe(res => {
      if (res) {
        this.myCodeService.removeTest({test_id: t.id}).subscribe(() => {
          this.tests = this.tests.filter((ee => t.id !== ee.id));
        });
      }
    });
  }

  formatDurationBack(s: string): string {
    return formatDurationBack(s);
  }

  formatMemoryBack(s: string): string {
    return formatMemoryBack(s);
  }
}
