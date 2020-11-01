import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {

  form: FormGroup;
  loginFailed: boolean;

  private returnPath: string;

  constructor(
    private formBuilder: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private authService: AuthService
  ) { }

  ngOnInit(): void {
    this.returnPath = this.route.snapshot.queryParams.returnPath || '';

    this.form = this.formBuilder.group({
      login: ['', Validators.required],
      password: ['', Validators.required],
    });

    if (this.authService.loggedIn) {
      this.toReturnPath();
    }
  }

  private toReturnPath(): void {
    this.router.navigateByUrl(this.returnPath);
  }

  login(): void {
    if (this.form.valid) {
      this.authService.login({
        login: this.form.controls.login.value,
        password: this.form.controls.password.value
      }).subscribe(() => {
        this.loginFailed = false;
        this.toReturnPath();
      }, () => {
        this.loginFailed = true;
      });
    }
  }

}
