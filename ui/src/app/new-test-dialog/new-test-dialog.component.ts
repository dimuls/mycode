import { Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { AddTestReq, checkerTestType, simpleTestType } from '../entity';
import { MyCodeService } from '../my-code.service';
import { languages, testTypes } from '../entity';
import { Subscription } from 'rxjs';

export function swap(items, firstIndex, secondIndex): string[] {
  const results = items.slice();
  const firstItem = items[firstIndex];
  results[firstIndex] = items[secondIndex];
  results[secondIndex] = firstItem;
  return results;
}

export function formatDuration(d: string): string {
  d = d.replace(/\s+/, '');
  if (d.match(/^\d+$/)) {
    return d + 's';
  }
  return d.replace(/мс/, 'ms').replace(/с/, 's');
}

export function formatMemory(m: string): string {
  m = m.replace(/\s+/, '');
  if (m.match(/^\d+$/)) {
    return m + 'MB';
  }
  return m.replace(/мб/i, 'MB').replace(/кб/i, 'KB');
}

@Component({
  selector: 'app-new-test-dialog',
  templateUrl: './new-test-dialog.component.html',
  styleUrls: ['./new-test-dialog.component.css']
})
export class NewTestDialogComponent implements OnInit, OnDestroy {

  testTypes = testTypes;
  languages = languages;

  simpleTestType = simpleTestType;
  checkerTestType = checkerTestType;

  form: FormGroup;

  private typeChanges: Subscription;

  constructor(
    public matDialogRef: MatDialogRef<NewTestDialogComponent>,
    private myCodeService: MyCodeService,
    private formBuilder: FormBuilder,
    @Inject(MAT_DIALOG_DATA) public exerciseID: number
  ) { }

  ngOnInit(): void {
    this.form = this.formBuilder.group({
      type: [simpleTestType, Validators.required],
      name: ['', Validators.required],
      max_duration: ['1с', [Validators.required, Validators.pattern(/^\s*[1-9]\d*\s*(?:с|мс|)\s*$/)]],
      max_memory: ['1мб', [Validators.required, Validators.pattern(/^\s*[1-9]\d*\s*(?:кб|мб|)\s*$/)]],
      stdin: [''],
      expected_stdout: [''],
      checker_language: [''],
      checker_source: ['']
    });

    this.typeChanges = this.form.controls.type.valueChanges.subscribe(t => {
      if (t === checkerTestType) {
        this.form.controls.checker_language.setValidators(Validators.required);
        this.form.controls.checker_source.setValidators(Validators.required);
      } else {
        this.form.controls.checker_language.clearValidators();
        this.form.controls.checker_source.clearValidators();
      }
    });
  }

  ngOnDestroy(): void {
    this.typeChanges.unsubscribe();
  }

  add(): void {
    if (this.form.valid) {
      const req: AddTestReq = {
        exercise_id: this.exerciseID,
        type: this.form.controls.type.value,
        name: this.form.controls.name.value,
        max_duration: formatDuration(this.form.controls.max_duration.value),
        max_memory: formatMemory(this.form.controls.max_memory.value),
        stdin: this.form.controls.stdin.value,
      };

      if (req.type === simpleTestType) {
        req.expected_stdout = this.form.controls.expected_stdout.value;
      } else if (req.type === checkerTestType) {
        req.checker_language = this.form.controls.checker_language.value;
        req.checker_source = this.form.controls.checker_source.value;
      }

      this.myCodeService.addTest(req).subscribe(resp => {
        this.matDialogRef.close({
          id: resp.test_id,
          ...req
        });
      });
    }
  }

  cancel(): void {
    this.matDialogRef.close();
  }

}
