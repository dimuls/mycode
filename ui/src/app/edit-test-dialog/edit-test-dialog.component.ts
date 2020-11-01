import { Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MyCodeService } from '../my-code.service';
import { formatDuration, formatMemory, swap } from '../new-test-dialog/new-test-dialog.component';
import { checkerTestType, EditTestReq, simpleTestType, Test } from '../entity';
import { languages, testTypes } from '../entity';
import { Subscription } from 'rxjs';

export function formatDurationBack(d: string): string {
  return d.replace(/ms/, 'мс').
    replace(/s/, 'с');
}

export function formatMemoryBack(d: string): string {
  return d.replace(/MB/, 'Мб').
    replace(/KB/, 'Кб');
}

function sameArgs(a: string[], b: string[]): boolean {
  if (a === b) { return true; }
  if (a == null || b == null) { return false; }
  if (a.length !== b.length) { return false; }
  for (let i = 0; i < a.length; ++i) {
    if (a[i] !== b[i]) { return false; }
  }
  return true;
}

@Component({
  selector: 'app-edit-test-dialog',
  templateUrl: './edit-test-dialog.component.html',
  styleUrls: ['./edit-test-dialog.component.css']
})
export class EditTestDialogComponent implements OnInit, OnDestroy {

  types = testTypes;
  languages = languages;

  simpleTestType = simpleTestType;
  checkerTestType = checkerTestType;

  form: FormGroup;

  private typeChanges: Subscription;

  constructor(
    public matDialogRef: MatDialogRef<EditTestDialogComponent>,
    private myCodeService: MyCodeService,
    private formBuilder: FormBuilder,
    @Inject(MAT_DIALOG_DATA) public originalTest: Test
  ) { }

  ngOnInit(): void {
    this.form = this.formBuilder.group({
      type: [this.originalTest.type, Validators.required],
      name: [this.originalTest.name, Validators.required],
      max_duration: [formatDurationBack(this.originalTest.max_duration),
        [Validators.required, Validators.pattern(/^\s*[1-9]\d*\s*(?:с|мс|)\s*$/)]],
      max_memory: [formatMemoryBack(this.originalTest.max_memory),
        [Validators.required, Validators.pattern(/^\s*[1-9]\d*\s*(?:кб|мб|)\s*$/i)]],
      stdin: [this.originalTest.stdin],
      expected_stdout: [this.originalTest.expected_stdout],
      checker_language: [this.originalTest.checker_language],
      checker_source: [this.originalTest.checker_source]
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

  edit(): void {
    if (this.form.valid) {
      const editedTest: Test = {
        ...this.originalTest,
      };
      const req: EditTestReq = {
        test_id: this.originalTest.id,
        max_duration: this.form.controls.max_duration.value,
        max_memory: this.form.controls.max_memory.value,
        stdin_set: false,
        expected_stdout_set: false,
        checker_langauge_set: false,
      };

      if (this.form.controls.name.value !== this.originalTest.name) {
        req.name = this.form.controls.name.value;
        editedTest.name = req.name;
      }

      if (this.form.controls.max_duration.value !== this.originalTest.max_duration) {
        req.max_duration = formatDuration(this.form.controls.max_duration.value);
        editedTest.max_duration = req.max_duration;
      }

      if (this.form.controls.max_memory.value !== this.originalTest.max_memory) {
        req.max_memory = formatMemory(this.form.controls.max_memory.value);
        editedTest.max_memory = req.max_memory;
      }

      if (this.form.controls.type.value === simpleTestType) {
        if (this.form.controls.expected_stdout.value !== this.originalTest.expected_stdout) {
          req.expected_stdout_set = true;
          req.expected_stdout = this.form.controls.expected_stdout.value;
          editedTest.expected_stdout = req.expected_stdout;
          editedTest.checker_language = '',
          editedTest.checker_source = '';
        }
      } else if (this.form.controls.type.value === checkerTestType) {
        editedTest.expected_stdout = '';

        if (this.form.controls.checker_language.value !== this.originalTest.checker_language) {
          req.checker_langauge_set = true;
          req.checker_language = this.form.controls.checker_language.value;
          editedTest.checker_language = req.checker_language;
        }

        if (this.form.controls.checker_source.value !== this.originalTest.checker_source) {
          req.checker_source = this.form.controls.checker_source.value;
          editedTest.checker_source = req.checker_source;
        }
      }

      this.myCodeService.editTest(req).subscribe(() => {
        this.matDialogRef.close(editedTest);
      });
    }
  }

  cancel(): void {
    this.matDialogRef.close();
  }

}
