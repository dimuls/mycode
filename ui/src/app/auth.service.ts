import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoginReq, LoginResp, Student, studentUR, Teacher, teacherUR } from './entity';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from '../environments/environment';
import { tap } from 'rxjs/operators';
import { Router } from '@angular/router';

const jwtKey = 'jwt';
const userRoleKey = 'userRole';
const userKey = 'user';

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  private pJWT: string;
  private pUserRole: string;
  private pTeacher: Teacher;
  private pStudent: Student;

  private loggedInSubject: BehaviorSubject<boolean>;
  loggedIn$: Observable<boolean>;

  constructor(
    private httpClient: HttpClient,
    private router: Router
  ) {
    this.loggedInSubject = new BehaviorSubject<boolean>(false);
    this.loggedIn$ = this.loggedInSubject.asObservable();

    this.pJWT = sessionStorage.getItem(jwtKey);
    if (this.pJWT) {
      this.pUserRole = sessionStorage.getItem(userRoleKey);
      if (this.pUserRole === teacherUR) {
        this.pTeacher = JSON.parse(sessionStorage.getItem(userKey));
      } else {
        this.pStudent = JSON.parse(sessionStorage.getItem(userKey));
      }
      this.loggedInSubject.next(true);
    }
  }

  get jwt(): string {
    return this.pJWT;
  }

  get userRole(): string {
    return this.pUserRole;
  }

  get teacher(): Teacher {
    return this.pTeacher;
  }

  get student(): Student {
    return this.pStudent;
  }

  get loggedIn(): boolean {
    return !!this.pJWT;
  }

  login(req: LoginReq): Observable<LoginResp> {
    return this.httpClient.post<LoginResp>(`${environment.apiURL}/Login`, req).pipe(tap(resp => {
      sessionStorage.setItem(jwtKey, resp.jwt);
      this.pJWT = resp.jwt;
      if (resp.teacher) {
        sessionStorage.setItem(userRoleKey, teacherUR);
        this.pUserRole = teacherUR;
        sessionStorage.setItem(userKey, JSON.stringify(resp.teacher));
        this.pTeacher = resp.teacher;
      } else {
        sessionStorage.setItem(userRoleKey, studentUR);
        this.pUserRole = studentUR;
        sessionStorage.setItem(userKey, JSON.stringify(resp.student));
        this.pStudent = resp.student;
      }
      this.loggedInSubject.next(true);
    }));
  }

  logout(withReturnPath?: boolean): void {
    this.pJWT = undefined;
    this.pUserRole = undefined;
    this.pTeacher = undefined;
    this.pStudent = undefined;
    sessionStorage.clear();
    this.loggedInSubject.next(false);
    if (withReturnPath) {
      this.router.navigate(['/login'], {queryParams: {returnPath: this.router.url}});
    } else {
      this.router.navigate(['/login']);
    }
  }
}
