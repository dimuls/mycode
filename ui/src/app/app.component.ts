import { Component, OnInit } from '@angular/core';
import { AuthService } from './auth.service';
import { studentUR, teacherUR } from './entity';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
})
export class AppComponent implements OnInit {

  teacherUR = teacherUR;
  studentUR = studentUR;

  userRole: string;

  constructor(
    private authService: AuthService,
  ) { }

  ngOnInit(): void {
    this.authService.loggedIn$.subscribe(loggedIn => {
      if (loggedIn) {
        this.userRole = this.authService.userRole;
      } else {
        this.userRole = '';
      }
    });
  }

  logout(): void {
    this.authService.logout(false);
  }
}
