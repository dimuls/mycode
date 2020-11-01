package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwtPkg "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/dimuls/mycode"
)

const (
	jwtCtxKey = "jwt"
)

type jwtGetter struct {
	base http.Handler
}

func (a *jwtGetter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jwt := r.Header.Get("Authorization")
	if jwt != "" {
		ctx := context.WithValue(r.Context(), jwtCtxKey, jwt)
		r = r.WithContext(ctx)
	}
	a.base.ServeHTTP(w, r)
}

func WithJWT(base http.Handler) http.Handler {
	auth := &jwtGetter{base: base}
	return auth
}

const (
	jwtTeacher = "teacher"
	jwtStudent = "student"
)

type jwtClaims struct {
	UserID   int64  `json:"user_id"`
	UserRole string `json:"user_role"`
	jwtPkg.StandardClaims
}

func (api *MyCodeAPI) user(ctx context.Context, login string) (
	u *mycode.User, err error) {

	u = &mycode.User{Login: login}

	err = api.db.QueryRowContext(ctx, `
		select id, password_hash from "user" where login = $1
	`, login).Scan(&u.Id, &u.PasswordHash)

	return
}

func (api *MyCodeAPI) Login(ctx context.Context, req *mycode.LoginReq) (
	*mycode.LoginResp, error) {

	if req.Login == "" {
		return nil, fmt.Errorf("login required")
	}

	if req.Password == "" {
		return nil, fmt.Errorf("password required")
	}

	u, err := api.user(ctx, req.Login)
	if err != nil {
		return nil, fmt.Errorf("wrong login or password")
	}

	err = bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("password check: %w", err)
	}

	resp := &mycode.LoginResp{}

	userRole := ctxTeacher
	resp.Teacher, err = api.teacher(ctx, u.Id)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			resp.Student, err = api.student(ctx, u.Id)
			if err != nil {

				if errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf(
						"neither teacher nor student found by user ID")
				}

				return nil, fmt.Errorf("get student from DB: %w", err)
			}

			userRole = ctxStudent

		} else {
			return nil, fmt.Errorf("get teacher from DB: %w", err)
		}

	}

	t := jwtPkg.NewWithClaims(jwtPkg.SigningMethodHS512, jwtClaims{
		UserID:   u.Id,
		UserRole: userRole,
		StandardClaims: jwtPkg.StandardClaims{
			ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
		},
	})

	resp.Jwt, err = t.SignedString([]byte(api.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}

	return resp, nil
}

func (api *MyCodeAPI) jwtKeyFunc(token *jwtPkg.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwtPkg.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(api.jwtSecret), nil
}

func (api *MyCodeAPI) teacher(ctx context.Context, userID int64) (
	*mycode.Teacher, error) {

	t := &mycode.Teacher{UserId: userID}

	err := api.db.QueryRowContext(ctx, `
		select id, name from teacher where user_id = $1
	`, userID).Scan(&t.Id, &t.Name)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (api *MyCodeAPI) student(ctx context.Context, userID int64) (
	*mycode.Student, error) {

	s := &mycode.Student{UserId: userID}

	err := api.db.QueryRowContext(ctx, `
		select id, name, class_id from student where user_id = $1
	`, userID).Scan(&s.Id, &s.Name, &s.ClassId)
	if err != nil {
		return nil, err
	}

	return s, nil
}

const (
	ctxUserRole = "user_role"
	ctxTeacher  = "teacher"
	ctxStudent  = "student"
)

var teacherMethods = map[string]struct{}{
	"GetClasses":             {},
	"GetStudents":            {},
	"GetExercise":            {},
	"AddExercise":            {},
	"EditExercise":           {},
	"RemoveExercise":         {},
	"GetExercises":           {},
	"GetExerciseAssignments": {},
	"AssignExercise":         {},
	"WithdrawExercise":       {},
	"AddTest":                {},
	"EditTest":               {},
	"RemoveTest":             {},
	"GetTests":               {},
	"GetSolutions":           {},
	"GetSolutionTests":       {},
}

var studentMethods = map[string]struct{}{
	"GetExercise":      {},
	"GetExercises":     {},
	"GetTests":         {},
	"AddSolution":      {},
	"GetSolutions":     {},
	"GetSolutionTests": {},
}

func (api *MyCodeAPI) Authorize(ctx context.Context, method string) (
	context.Context, error) {

	if method == "Login" {
		return ctx, nil
	}

	jwt, ok := ctx.Value(jwtCtxKey).(string)
	if !ok {
		return ctx, fmt.Errorf("missing jwt")
	}

	token, err := jwtPkg.ParseWithClaims(jwt, &jwtClaims{}, api.jwtKeyFunc)
	if err != nil {
		return ctx, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return ctx, fmt.Errorf("unexpected claims type: %T", token.Claims)
	}

	switch claims.UserRole {

	case jwtTeacher:
		_, exists := teacherMethods[method]
		if !exists {
			return ctx, fmt.Errorf("method `%s` not allowed for teacher",
				method)
		}

		t, err := api.teacher(ctx, claims.UserID)
		if err != nil {
			return ctx, fmt.Errorf("get teacher: %w", err)
		}

		ctx = context.WithValue(ctx, ctxUserRole, ctxTeacher)
		ctx = context.WithValue(ctx, ctxTeacher, t)

	case jwtStudent:
		_, exists := studentMethods[method]
		if !exists {
			return ctx, fmt.Errorf("method `%s` not allowed for student",
				method)
		}

		t, err := api.student(ctx, claims.UserID)
		if err != nil {
			return ctx, fmt.Errorf("get student: %w", err)
		}

		ctx = context.WithValue(ctx, ctxUserRole, ctxStudent)
		ctx = context.WithValue(ctx, ctxStudent, t)

	default:
		return ctx, fmt.Errorf("unexpected role: %v", claims.UserRole)
	}

	return ctx, nil
}
