import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatDialogRef } from '@angular/material/dialog';
import { AuthService } from '../auth.service';
import { estimators, languages } from '../entity';
import { MyCodeService } from '../my-code.service';

@Component({
  selector: 'app-new-exercise-dialog',
  templateUrl: './new-exercise-dialog.component.html',
  styleUrls: ['./new-exercise-dialog.component.css']
})
export class NewExerciseDialogComponent implements OnInit {

  languages = languages;
  estimators = estimators;

  form: FormGroup;

  constructor(
    public matDialogRef: MatDialogRef<NewExerciseDialogComponent>,
    private authService: AuthService,
    private myCodeService: MyCodeService,
    private formBuilder: FormBuilder
  ) { }

  ngOnInit(): void {
    this.form = this.formBuilder.group({
      title: ['', Validators.required],
      description: ['', Validators.required],
      language: ['', Validators.required],
      estimator: ['', Validators.required]
    });
  }

  add(): void {
    if (this.form.valid) {
      const req = {
        title: this.form.controls.title.value,
        description: this.form.controls.description.value,
        language: this.form.controls.language.value,
        estimator: this.form.controls.estimator.value,
      };
      this.myCodeService.addExercise(req).subscribe(resp => {
        this.matDialogRef.close({
          id: resp.exercise_id,
          teacher_id: this.authService.teacher.id,
          ...req
        });
      });
    }
  }

  cancel(): void {
    this.matDialogRef.close();
  }

}
