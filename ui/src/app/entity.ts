export const teacherUR = 'teacher';
export const studentUR = 'student';

export interface Teacher {
  id: number;
  user_id: number;
  name: string;
}

export interface Class {
  id: number;
  teacher_id: number;
  name: string;
}

export interface Student {
  id: number;
  user_id: number;
  class_id: number;
  name: string;
}

export const cLang = 'c';
export const cppLang = 'cpp';
export const goLang = 'go';
export const javaLang = 'java';
export const pascalLang = 'pascal';
export const pythonLang = 'python';

export interface Language {
  id: string;
  name: string;
  extension: string;
}

export const languages: Language[] = [
  { id: cLang, name: 'C', extension: 'c' },
  { id: cppLang, name: 'C++', extension: 'cpp' },
  { id: goLang, name: 'Go', extension: 'go' },
  { id: javaLang, name: 'Java', extension: 'java' },
  { id: pascalLang, name: 'Pascal', extension: 'pas' },
  { id: pythonLang, name: 'Python', extension: 'py' }
];

export const linearEstimator = 'linear';
export const exponentialEstimator = 'exponential';
export const logarithmicEstimator = 'logarithmic';

export interface Estimator {
  id: string;
  name: string;
}

export const estimators: Estimator[] = [
  { id: linearEstimator, name: 'Линейный' },
  { id: exponentialEstimator, name: 'Экспоненциальный' },
  { id: logarithmicEstimator, name: 'Логарифмический' },
];

export interface Exercise {
  id: number;
  teacher_id: number;
  title: string;
  description: string;
  language: string;
  estimator: string;
}

export const simpleTestType = 'simple';
export const checkerTestType = 'checker';

export interface TestType {
  id: string;
  name: string;
}

export const testTypes: TestType[] = [
  { id: simpleTestType, name: 'Простой' },
  { id: checkerTestType, name: 'Чекер' },
];

export interface Test {
  id: number;
  exercise_id: number;
  type: string;
  name: string;
  max_duration: string;
  max_memory: string;
  stdin: string;
  expected_stdout: string;
  checker_language: string;
  checker_source: string
}

export interface Solution {
  id: number;
  student_id: number;
  exercise_id: number;
  source: string;
}

export const processingTS = 0;
export const failedTS = 1;
export const succeedTS = 2;

export interface SolutionTestFails {
  wrong_duration: boolean;
  wrong_used_memory: boolean;
  wrong_stdout: boolean;
  wrong_checker: boolean;
}

export interface SolutionTest {
  id: number;
  solution_id: number;
  test_id: number;
  status: string;
  duration: string;
  used_memory: string;
  stdout: string;
  stderr: string;
  checker_stdout?: string;
  checker_stderr?: string;
  fails: SolutionTestFails;
}

export interface LoginReq {
  login: string;
  password: string;
}

export interface LoginResp {
  jwt: string;
  teacher?: Teacher;
  student?: Student;
}

export interface GetClassesResp {
  classes: Class[];
}

export interface GetStudentsResp {
  students: Student[];
}

export interface AddExerciseReq {
  title: string;
  description: string;
  language: string;
  estimator: string;
}

export interface AddExerciseResp {
  exercise_id: number;
}

export interface EditExerciseReq {
  exercise_id: number;
  title?: string;
  description?: string;
  language?: string;
  language_set: boolean;
  estimator?: string;
  estimator_set: boolean;
}

export interface RemoveExerciseReq {
  exercise_id: number;
}

export interface GetExercisesReq {
  student_id?: number;
}

export interface GetExercisesResp {
  exercises: Exercise[];
}

export interface GetExerciseReq {
  exercise_id: number;
}

export interface GetExerciseResp {
  exercise: Exercise;
}

export interface GetExerciseAssignmentsReq {
  exercise_id: number;
}

export interface GetExerciseAssignmentsResp {
  student_ids: number[];
}

export interface AssignExerciseReq {
  exercise_id: number;
  class_id?: number;
  student_id?: number;
}

export interface WithdrawExerciseReq {
  exercise_id: number;
  class_id?: number;
  student_id?: number;
}

export interface AddTestReq {
  exercise_id: number;
  type: string;
  name: string;
  max_duration: string;
  max_memory: string;
  stdin: string;
  expected_stdout?: string;
  checker_language?: string;
  checker_source?: string;
}

export interface AddTestResp {
  test_id: number;
}

export interface EditTestReq {
  test_id: number;
  name?: string;
  max_duration: string;
  max_memory: string;
  stdin?: string;
  stdin_set: boolean;
  expected_stdout?: string;
  expected_stdout_set: boolean;
  checker_language?: string;
  checker_langauge_set: boolean;
  checker_source?: string;
}

export interface RemoveTestReq {
  test_id: number;
}

export interface GetTestsReq {
  exercise_id?: number;
  student_id?: number;
}

export interface GetTestsResp {
  tests: Test[];
}

export interface AddSolutionReq {
  exercise_id: number;
  source: string;
}

export interface AddSolutionResp {
  solution_id: number;
}

export interface GetSolutionsReq {
  exercise_id?: number;
  student_id?: number;
}

export interface GetSolutionsResp {
  solutions: Solution[];
}

export interface GetSolutionTestsReq {
  student_id?: number;
  solution_id?: number;
}

export interface GetSolutionTestsResp {
  solution_tests: SolutionTest[];
}
