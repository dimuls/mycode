import { Component, Inject, OnInit } from '@angular/core';
import { EditExerciseReq, estimators, Exercise, languages } from '../entity';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MyCodeService } from '../my-code.service';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

@Component({
  selector: 'app-edit-exercise-dialog',
  templateUrl: './edit-exercise-dialog.component.html',
  styleUrls: ['./edit-exercise-dialog.component.css']
})
export class EditExerciseDialogComponent implements OnInit {

  languages = languages;
  estimators = estimators;

  form: FormGroup;

  constructor(
    public matDialogRef: MatDialogRef<EditExerciseDialogComponent>,
    private myCodeService: MyCodeService,
    private formBuilder: FormBuilder,
    @Inject(MAT_DIALOG_DATA) public originalExercise: Exercise
  ) {
  }

  ngOnInit(): void {
    this.form = this.formBuilder.group({
      title: [this.originalExercise.title, Validators.required],
      description: [this.originalExercise.description, Validators.required],
      language: [this.originalExercise.language, Validators.required],
      estimator: [this.originalExercise.estimator, Validators.required]
    });
  }

  edit(): void {
    if (this.form.valid) {
      const req: EditExerciseReq = {
        exercise_id: this.originalExercise.id,
        language_set: false,
        estimator_set: false
      };

      let changed = false;

      if (this.form.controls.title.value !== this.originalExercise.title) {
        req.title = this.form.controls.title.value;
        changed = true;
      }

      if (this.form.controls.description.value !== this.originalExercise.description) {
        req.description = this.form.controls.description.value;
        changed = true;
      }

      // if (this.form.controls.language.value !== this.originalExercise.language) {
      //   req.new_language = this.form.controls.language.value;
      //   changed = true;
      // }

      if (this.form.controls.estimator.value !== this.originalExercise.estimator) {
        req.estimator = this.form.controls.estimator.value;
        req.estimator_set = true;
        changed = true;
      }

      if (!changed) {
        this.matDialogRef.close();
        return;
      }

      this.myCodeService.editExercise(req).subscribe(_ => {
        this.matDialogRef.close({
          ...this.originalExercise,
          title: req.title || this.originalExercise.title,
          description: req.description || this.originalExercise.description,
          language: this.originalExercise.language,
          estimator: req.estimator || this.originalExercise.estimator
        });
      });
    }
  }

  cancel(): void {
    this.matDialogRef.close();
  }
}
