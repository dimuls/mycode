import { Component, OnInit } from '@angular/core';
import { Exercise, exponentialEstimator, languages, linearEstimator, logarithmicEstimator, Solution, SolutionTest, Test } from '../entity';
import { MyCodeService } from '../my-code.service';
import { combineLatest } from 'rxjs';
import { tap } from 'rxjs/operators';
import { SolutionScore, SolutionStats } from '../teacher-solutions/teacher-solutions.component';
import { formatDurationBack, formatMemoryBack } from '../edit-test-dialog/edit-test-dialog.component';
import chroma from 'chroma-js';

@Component({
  selector: 'app-student-exercises',
  templateUrl: './student-exercises.component.html',
  styleUrls: ['./student-exercises.component.css']
})
export class StudentExercisesComponent implements OnInit {

  loading = true;

  languageNames: { [key: number]: string };
  languageExtensions: { [key: number]: string };
  exercises: { [key: number]: Exercise[] };
  exercisesTests: { [key: number]: Test[] };
  tests: { [key: number]: Test };
  solutions: { [key: number]: Solution[] };
  latestSolutions: { [key: number]: Solution };
  solutionTests: { [key: number]: SolutionTest[] };
  solutionsStats: { [key: number]: SolutionStats };
  sources: { [key: number]: string } = {};

  scores: { [key: number]: SolutionScore } = {};

  constructor(
    private myCodeService: MyCodeService
  ) { }

  ngOnInit(): void {
    this.loading = true;

    const getExercises = this.myCodeService.getExercises({}).pipe(tap(resp => {
      this.exercises = resp.exercises.reduce((es, e) => {
        return { ...es, [e.language]: [...es[e.language] || [], e] };
      }, {});
    }));

    const getTests = this.myCodeService.getTests({}).pipe(tap(resp => {
      this.tests = resp.tests.reduce((ts, t) => {
        return { ...ts, [t.id]: t };
      }, {});
      this.exercisesTests = resp.tests.reduce((ts, t) => {
        return { ...ts, [t.exercise_id]: [...(ts[t.exercise_id] || []), t] };
      }, {});
    }));

    const getSolutions = this.myCodeService.getSolutions({}).pipe(tap(resp => {
      this.latestSolutions = {};
      this.solutions = resp.solutions.reduce((ss, s) => {
        if (!this.latestSolutions[s.exercise_id] || this.latestSolutions[s.exercise_id].id < s.id) {
          this.latestSolutions[s.exercise_id] = s;
        }
        return { ...ss, [s.id]: [...(ss[s.id] || []), s] };
      }, {});
    }));

    const getSolutionsTests = this.myCodeService.getSolutionTests({}).pipe(tap(resp => {
      this.solutionsStats = {};
      this.solutionTests = resp.solution_tests.reduce((sts, st) => {
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
        return { ...sts, [st.solution_id]: [...(sts[st.solution_id] || []), st] };
      }, {});
    }));

    this.languageNames = languages.reduce((ls, l) => {
      return { ...ls, [l.id]: l.name };
    }, {});

    this.languageExtensions = languages.reduce((ls, l) => {
      return { ...ls, [l.id]: l.extension };
    }, {});

    combineLatest([getExercises, getTests, getSolutions, getSolutionsTests]).subscribe(([eResp, tResp, sResp, stResp]) => {
      this.loading = false;

      const exMap: { [key: number]: Exercise } = {};

      eResp.exercises.forEach(e => exMap[e.id] = e);

      const sidToEx: { [key: number]: Exercise } = {};

      sResp.solutions.forEach(ss => {
        sidToEx[ss.id] = exMap[ss.exercise_id];
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
    });
  }

  addSolution(ex: Exercise, f: File): void {
    if (f) {
      const r = new FileReader();
      r.onload = (e) => {
        if (typeof(e.target.result) === 'string') {
          this.myCodeService.addSolutionReq({exercise_id: ex.id, source: e.target.result}).subscribe(() => {});
        }
      };
      r.readAsText(f);
    } else if (this.sources[ex.id]) {
      this.myCodeService.addSolutionReq({exercise_id: ex.id, source: this.sources[ex.id]}).subscribe(() => {
        this.sources[ex.id] = '';
      });
    }
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
