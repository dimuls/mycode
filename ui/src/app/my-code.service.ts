import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  AddExerciseReq,
  AddExerciseResp,
  AddSolutionReq,
  AddSolutionResp,
  AddTestReq,
  AddTestResp,
  AssignExerciseReq,
  EditExerciseReq,
  EditTestReq,
  GetClassesResp, GetExerciseAssignmentsReq, GetExerciseAssignmentsResp, GetExerciseReq, GetExerciseResp, GetExercisesReq,
  GetExercisesResp,
  GetSolutionsReq,
  GetSolutionsResp,
  GetSolutionTestsReq,
  GetSolutionTestsResp, GetStudentsResp,
  GetTestsReq,
  GetTestsResp, RemoveExerciseReq,
  RemoveTestReq,
  WithdrawExerciseReq
} from './entity';
import { environment } from '../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class MyCodeService {

  constructor(
    private httpClient: HttpClient
  ) { }

  getClasses(): Observable<GetClassesResp> {
    return this.httpClient.post<GetClassesResp>(`${environment.apiURL}/GetClasses`, {});
  }

  getStudents(): Observable<GetStudentsResp> {
    return this.httpClient.post<GetStudentsResp>(`${environment.apiURL}/GetStudents`, {});
  }

  addExercise(req: AddExerciseReq): Observable<AddExerciseResp> {
    return this.httpClient.post<AddExerciseResp>(`${environment.apiURL}/AddExercise`, req);
  }

  editExercise(req: EditExerciseReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/EditExercise`, req);
  }

  removeExercise(req: RemoveExerciseReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/RemoveExercise`, req);
  }

  getExercises(req: GetExercisesReq): Observable<GetExercisesResp> {
    return this.httpClient.post<GetExercisesResp>(`${environment.apiURL}/GetExercises`, req);
  }

  getExercise(req: GetExerciseReq): Observable<GetExerciseResp> {
    return this.httpClient.post<GetExerciseResp>(`${environment.apiURL}/GetExercise`, req);
  }

  getExerciseAssignments(req: GetExerciseAssignmentsReq): Observable<GetExerciseAssignmentsResp> {
    return this.httpClient.post<GetExerciseAssignmentsResp>(`${environment.apiURL}/GetExerciseAssignments`, req);
  }

  assignExercise(req: AssignExerciseReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/AssignExercise`, req);
  }

  withdrawExercise(req: WithdrawExerciseReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/WithdrawExercise`, req);
  }

  addTest(req: AddTestReq): Observable<AddTestResp> {
    return this.httpClient.post<AddTestResp>(`${environment.apiURL}/AddTest`, req);
  }

  editTest(req: EditTestReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/EditTest`, req);
  }

  removeTest(req: RemoveTestReq): Observable<void> {
    return this.httpClient.post<void>(`${environment.apiURL}/RemoveTest`, req);
  }

  getTests(req: GetTestsReq): Observable<GetTestsResp> {
    return this.httpClient.post<GetTestsResp>(`${environment.apiURL}/GetTests`, req);
  }

  addSolutionReq(req: AddSolutionReq): Observable<AddSolutionResp> {
    return this.httpClient.post<AddSolutionResp>(`${environment.apiURL}/AddSolution`, req);
  }

  getSolutions(req: GetSolutionsReq): Observable<GetSolutionsResp> {
    return this.httpClient.post<GetSolutionsResp>(`${environment.apiURL}/GetSolutions`, req);
  }

  getSolutionTests(req: GetSolutionTestsReq): Observable<GetSolutionTestsResp> {
    return this.httpClient.post<GetSolutionTestsResp>(`${environment.apiURL}/GetSolutionTests`, req);
  }
}
