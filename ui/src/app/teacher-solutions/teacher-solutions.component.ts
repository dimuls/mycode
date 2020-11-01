import { Component, OnDestroy, OnInit } from '@angular/core';
import { Class, Exercise, exponentialEstimator, languages, linearEstimator, logarithmicEstimator, Solution, SolutionTest, Student, Test } from '../entity';
import { MyCodeService } from '../my-code.service';
import { combineLatest, Subject } from 'rxjs';
import naturalSort from 'typescript-natural-sort';
import Fuse from 'fuse.js';
import { debounceTime, takeUntil } from 'rxjs/operators';
import { ActivatedRoute } from '@angular/router';
import { formatDurationBack, formatMemoryBack } from '../edit-test-dialog/edit-test-dialog.component';
import chroma from 'chroma-js';

export interface SolutionStats {
  total: number;
  failed: number;
  succeed: number;
  processing: number;
}

export interface SolutionScore {
  score: number;
  color: string;
}

@Component({
  selector: 'app-teacher-solutions',
  templateUrl: './teacher-solutions.component.html',
  styleUrls: ['./teacher-solutions.component.css'],
})
export class TeacherSolutionsComponent implements OnInit, OnDestroy {

  languages = languages.reduce<{[key: string]: string}>((h, l) => {
    return { ...h, [l.id]: l.name };
  }, {});

  private stop = new Subject<void>();
  searchQueries = new Subject<string>();

  loading = true;

  private studentsFuse: Fuse<Student>;
  private classesFuse: Fuse<Class>;
  foundClasses: Class[];
  foundStudents: Student[];

  classes: Class[] = [];
  classesNames: { [key: number]: string } = {};
  classesStudents: { [key: number]: Student[] } = {};

  students: { [key: number]: Student } = {};

  student: Student;
  exercises: Exercise[];
  solutions: { [key: number]: Solution[] };
  solutionTests: { [key: number]: SolutionTest[] };
  tests: { [key: number]: Test };
  solutionsStats: { [key: number]: SolutionStats };

  scores: { [key: number]: SolutionScore } = {};

  constructor(
    private myCodeService: MyCodeService,
    private activatedRoute: ActivatedRoute
  ) { }

  ngOnInit(): void {

    const classesReq = this.myCodeService.getClasses();
    const studentsReq = this.myCodeService.getStudents();

    combineLatest([classesReq, studentsReq]).subscribe(([classesResp, studentsResp]) => {
      const ss: Student[] = [];

      studentsResp.students.forEach(s => {
        ss.push(s);
        if (!this.classesStudents[s.class_id]) {
          this.classesStudents[s.class_id] = [];
        }
        this.classesStudents[s.class_id].push(s);
        this.students[s.id] = s;
      });

      this.studentsFuse = new Fuse<Student>(ss, {
        keys: ['name'],
        shouldSort: true,
      });

      classesResp.classes.sort((a, b) => {
        return naturalSort(a.name, b.name);
      }).forEach(c => {
        this.classes.push(c);
        if (this.classesStudents[c.id]) {
          this.classesStudents[c.id].sort((a, b) => {
            return naturalSort(a.name, b.name);
          });
        }
        this.classesNames[c.id] = c.name;
      });

      this.classesFuse = new Fuse<Class>(this.classes, {
        keys: ['name'],
        shouldSort: true
      });

      this.activatedRoute.queryParams.subscribe(p => {
        if (p.student_id) {
          this.chooseStudent(this.students[p.student_id]);
        } else {
          this.student = null;
          this.exercises = null;
          this.solutions = null;
          this.solutionTests = null;
          this.loading = false;
        }
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

  chooseStudent(s: Student): void {
    const req = { student_id: s.id };
    const getExercises = this.myCodeService.getExercises(req);
    const getSolutions = this.myCodeService.getSolutions(req);
    const getSolutionTests = this.myCodeService.getSolutionTests(req);
    const getTests = this.myCodeService.getTests({ student_id: s.id });

    this.loading = true;

    combineLatest([getExercises, getSolutions, getSolutionTests, getTests]).subscribe(([eResp, sResp, stResp, tResp]) => {

      this.student = s;

      this.exercises = eResp.exercises.sort((a, b) => {
        const ns = naturalSort(a.language, b.language);
        if (ns === 0) {
          if (a.id > b.id) {
            return 1;
          } else if (a.id < b.id) {
            return -1;
          } else {
            return 0;
          }
        }
        return ns;
      });

      const exMap: { [key: number]: Exercise } = {};
      this.exercises.forEach(e => exMap[e.id] = e);

      const sidToEx: { [key: number]: Exercise } = {};

      this.solutions = {};
      sResp.solutions.forEach(ss => {
        if (!this.solutions[ss.exercise_id]) {
          this.solutions[ss.exercise_id] = [];
        }
        this.solutions[ss.exercise_id].push(ss);
        sidToEx[ss.id] = exMap[ss.exercise_id];
      });

      for (const eid in this.solutions) {
        if (!Object.prototype.hasOwnProperty.call(this.solutions, eid)) {
          continue;
        }
        this.solutions[eid].sort((a, b) => {
          return a.id === b.id ? 0 : a.id < b.id ? 1 : -1;
        });
      }

      this.solutionTests = {};
      this.solutionsStats = {};
      stResp.solution_tests.forEach(st => {
        if (!this.solutionTests[st.solution_id]) {
          this.solutionTests[st.solution_id] = [];
        }
        this.solutionTests[st.solution_id].push(st);
        if (!this.solutionsStats[st.solution_id]) {
          this.solutionsStats[st.solution_id] = {
            total: 0,
            failed: 0,
            succeed: 0,
            processing: 0
          };
        }
        this.solutionsStats[st.solution_id].total++;
        this.solutionsStats[st.solution_id][st.status]++;
      });

      const scale = chroma.scale(['red', '#e6e600', 'green']);

      for (const sid in this.solutionsStats) {
        if (Object.prototype.hasOwnProperty.call(this.solutionsStats, sid)) {
          const ss = this.solutionsStats[sid];
          const estimator = sidToEx[sid].estimator;
          let score: number;
          if (!ss.processing) {
            if (estimator === linearEstimator) {
              score = (ss.succeed / ss.total);
            } else if (estimator === logarithmicEstimator) {
              score = (-1.3 / Math.pow(ss.succeed / ss.total + 1, 2.1) + 1.3);
            } else if (estimator === exponentialEstimator) {
              score = ((Math.pow(10, ss.succeed / ss.total) - 1) / (10 - 1));
            }
          }
          if (score !== undefined) {
            this.scores[sid] = {
              score: Math.round(100 * score),
              color: scale(score)
            };
          }
        }
      }

      for (const sid in this.solutionTests) {
        if (!this.solutionTests.hasOwnProperty(sid)) {
          continue;
        }
        this.solutionTests[sid].sort((a, b) => {
          if (a.id === b.id) {
            return 0;
          }
          return a.id < b.id ? -1 : 1;
        });
      }

      this.tests = tResp.tests.reduce((ts, t) => ({ ...ts, [t.id]: t }), {});

      this.loading = false;
    });
  }

  solutionStatus(ss: SolutionStats): string {
    if (ss.processing > 0) {
      return 'в процессе тестирования';
    } else if (ss.failed > 0) {
      if (ss.failed === ss.total) {
        return 'все тесты провалены';
      } else {
        return 'часть тестов провалено';
      }
    }
    return 'все тесты успешно пройдены';
  }

  solutionStatusColor(ss: SolutionStats): string {
    if (ss.processing > 0) {
      return 'blue';
    } else if (ss.failed > 0) {
      return 'red';
    }
    return 'green';
  }

  formatDurationBack(d: string): string {
    return formatDurationBack(d);
  }

  formatMemoryBack(m: string): string {
    return formatMemoryBack(m);
  }
}
