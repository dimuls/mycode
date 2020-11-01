import { Component, OnInit } from '@angular/core';
import { AuthService } from '../auth.service';
import { Router } from '@angular/router';
import { teacherUR } from '../entity';

@Component({
  selector: 'app-index',
  templateUrl: './index.component.html',
  styleUrls: ['./index.component.css']
})
export class IndexComponent implements OnInit {

  constructor(
    private authService: AuthService,
    private router: Router
  ) { }

  ngOnInit(): void {
    if (this.authService.loggedIn) {
      if (this.authService.userRole === teacherUR) {
        this.router.navigate(['/teacher/solutions']);
      } else {
        this.router.navigate(['/student/exercises']);
      }
    } else {
      this.router.navigate(['/login']);
    }
  }

}
