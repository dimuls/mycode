import { Component, OnDestroy, OnInit } from '@angular/core';
import { combineLatest, Subject } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { Class, Exercise, Student, languages } from '../entity';
import { MyCodeService } from '../my-code.service';
import Fuse from 'fuse.js';
import naturalSort from 'typescript-natural-sort';
import { debounceTime, takeUntil } from 'rxjs/operators';

@Component({
  selector: 'app-teacher-exercise-assignments',
  templateUrl: './teacher-exercise-assignments.component.html',
  styleUrls: ['./teacher-exercise-assignments.component.css']
})
export class TeacherExerciseAssignmentsComponent implements OnInit, OnDestroy {

  exerciseID: number;
  exercise: Exercise;

  private stop = new Subject<void>();
  searchQueries = new Subject<string>();

  private studentsFuse: Fuse<Student>;
  private classesFuse: Fuse<Class>;
  foundClasses: Class[];
  foundStudents: Student[];

  classes: Class[] = [];
  classesNames: { [key: number]: string } = {};
  students: { [key: number]: Student[] } = {};

  assignments: { [key: number]: { [key: number]: boolean } } = {};

  languages: { [id: string]: string };

  constructor(
    private activatedRoute: ActivatedRoute,
    private myCodeService: MyCodeService
  ) {
    this.languages = {};
    languages.forEach(i => {
      this.languages[i.id] = i.name;
    });
  }

  ngOnInit(): void {
    this.activatedRoute.params.subscribe(p => {
      this.exerciseID = p.exercise_id;

      const getClasses = this.myCodeService.getClasses();
      const getStudents = this.myCodeService.getStudents();
      const getExercise = this.myCodeService.getExercise({ exercise_id: p.exercise_id });
      const getExerciseAssignments = this.myCodeService.getExerciseAssignments({exercise_id: p.exercise_id});

      combineLatest([getClasses, getStudents, getExercise, getExerciseAssignments]).
        subscribe(([classesResp, studentsResp, exerciseResp, exAssResp]) => {
          this.exercise = exerciseResp.exercise;

          const ss: Student[] = [];
          const sa: { [key: number]: true } = {};

          exAssResp.student_ids.forEach(sid => sa[sid] = true);

          studentsResp.students.forEach(s => {
            ss.push(s);
            if (this.students[s.class_id]) {
              this.students[s.class_id].push(s);
            } else {
              this.students[s.class_id] = [s];
            }
            if (sa[s.id]) {
              if (!this.assignments[s.class_id]) {
                this.assignments[s.class_id] = {};
              }
              this.assignments[s.class_id][s.id] = true;
            }
          });

          this.studentsFuse = new Fuse<Student>(ss, {
            keys: ['name'],
            shouldSort: true,
          });

          classesResp.classes.sort((a, b) => {
            return naturalSort(a.name, b.name);
          }).forEach(c => {
            this.classes.push(c);
            if (this.students[c.id]) {
              this.students[c.id].sort((a, b) => {
                return naturalSort(a.name, b.name);
              });
            }
            this.classesNames[c.id] = c.name;
          });

          this.classesFuse = new Fuse<Class>(this.classes, {
            keys: ['name'],
            shouldSort: true
          });
        });
    });

    this.searchQueries.pipe(
      debounceTime(500),
      takeUntil(this.stop)
    ).subscribe(q => {
      if (q.match(/\d/)) {
        this.foundStudents = null;
        this.foundClasses = this.classesFuse.search(q).map(i => i.item);
      } else {
        if (q.length < 3) {
          this.foundClasses = null;
          this.foundStudents = null;
        } else {
          this.foundClasses = null;
          this.foundStudents = this.studentsFuse.search(q).map(i => i.item);
        }
      }
    });
  }

  ngOnDestroy(): void {
    this.stop.next();
    this.stop.complete();
  }

  classAssigned(c: Class): boolean {
    return !!this.assignments[c.id] && Object.keys(this.assignments[c.id]).length === this.students[c.id].length;
  }

  classWithdrawn(c: Class): boolean {
    return !this.assignments[c.id];
  }

  studentAssigned(s: Student): boolean {
    return !!this.assignments[s.class_id] && !!this.assignments[s.class_id][s.id];
  }

  assignToStudent(s: Student): void {
    this.myCodeService.assignExercise({
      exercise_id: this.exerciseID,
      student_id: s.id
    }).subscribe(() => {
      if (!this.assignments[s.class_id]) {
        this.assignments = { ...this.assignments, [s.class_id]: { [s.id]: true } };
      } else {
        this.assignments[s.class_id] = { ...this.assignments[s.class_id], [s.id]: true };
      }
    });
  }

  assignToClass(c: Class): void {
    this.myCodeService.assignExercise({
      exercise_id: this.exerciseID,
      class_id: c.id
    }).subscribe(() => {
      const newSs = {};
      this.students[c.id].forEach(s => newSs[s.id] = true);
      if (!this.assignments[c.id]) {
        const cid = c.id;
        this.assignments = { ...this.assignments, [cid]: newSs };
      } else {
        this.assignments[c.id] = { ...this.assignments[c.id], ...newSs };
      }
    });
  }

  withdrawFromStudent(s: Student): void {
    this.myCodeService.withdrawExercise({
      exercise_id: this.exerciseID,
      student_id: s.id
    }).subscribe(() => {
      const { [s.id]: _, ...newSs } = this.assignments[s.class_id];
      if (Object.keys(newSs).length === 0) {
        const { [s.class_id]: __, ...newCs } = this.assignments;
        this.assignments = newCs;
      } else {
        this.assignments[s.class_id] = newSs;
      }
    });
  }

  withdrawFromClass(c: Class): void {
    this.myCodeService.withdrawExercise({
      exercise_id: this.exerciseID,
      class_id: c.id
    }).subscribe(() => {
      const { [c.id]: _, ...newCs } = this.assignments;
      this.assignments = newCs;
    });
  }
}
