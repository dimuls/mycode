package pg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"

	"github.com/dimuls/mycode"
)

func parseBytes(s string) (datasize.ByteSize, error) {
	var v datasize.ByteSize
	return v, v.UnmarshalText([]byte(s))
}

func (api *MyCodeAPI) AddTest(ctx context.Context,
	req *mycode.AddTestReq) (*mycode.AddTestResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("empty name")
	}

	_, err := time.ParseDuration(req.MaxDuration)
	if err != nil {
		return nil, fmt.Errorf("parse max_duration: %w", err)
	}

	_, err = parseBytes(req.MaxMemory)
	if err != nil {
		return nil, fmt.Errorf("parse max_memory: %w", err)
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	var exerciseTeacherID int64

	err = api.db.QueryRowContext(ctx, `
		select teacher_id from exercise where id = $1
	`, req.ExerciseId).Scan(&exerciseTeacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("test exercise doesn't exists")
		}
	}

	if t.Id != exerciseTeacherID {
		return nil, fmt.Errorf("test exercise doesn't belongs to teacher")
	}

	var id int64

	switch req.Type {
	case mycode.TestType_simple:
		err = api.db.QueryRowContext(ctx, `
			insert into test (
				exercise_id, type, name, max_duration, max_memory, stdin,
				expected_stdout, checker_language, checker_source)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			returning id
		`, req.ExerciseId, req.Type, req.Name, req.MaxDuration, req.MaxMemory,
			req.Stdin, req.ExpectedStdout, nil, nil).
			Scan(&id)
	case mycode.TestType_checker:
		err = api.db.QueryRowContext(ctx, `
			insert into test (
				exercise_id, type, name, max_duration, max_memory, stdin,
				expected_stdout, checker_language, checker_source)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			returning id
		`, req.ExerciseId, req.Type, req.Name, req.MaxDuration, req.MaxMemory,
			req.Stdin, nil, req.CheckerLanguage, req.CheckerSource).
			Scan(&id)
	default:
		return nil, fmt.Errorf("invalid type")
	}
	if err != nil {
		return nil, fmt.Errorf("add test to DB: %w", err)
	}

	return &mycode.AddTestResp{TestId: id}, nil
}

func (api *MyCodeAPI) checkTestBelongsToTeacher(ctx context.Context,
	testID, teacherID int64) error {

	var exerciseTeacherID int64

	err := api.db.QueryRowContext(ctx, `
		select e.teacher_id from test as t
		join exercise as e on t.exercise_id = e.id
		where t.id = $1
	`, testID).Scan(&exerciseTeacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("test exercise doesn't exists")
		}
		return fmt.Errorf("check test belongs to teacher: %w", err)
	}

	if teacherID != exerciseTeacherID {
		return fmt.Errorf("test exercise doesn't belongs to teacher")
	}

	return nil
}

func (api *MyCodeAPI) EditTest(ctx context.Context,
	req *mycode.EditTestReq) (*mycode.EditTestResp, error) {

	if req.TestId == 0 {
		return nil, fmt.Errorf("empty test_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkTestBelongsToTeacher(ctx, req.TestId, t.Id)
	if err != nil {
		return nil, err
	}

	var testType mycode.TestType

	err = api.db.QueryRowContext(ctx, `
		select type from test where id = $1
	`, req.TestId).Scan(&testType)
	if err != nil {
		return nil, fmt.Errorf("geet test type: %w", err)
	}

	var (
		args []interface{}
		sets []string
	)

	if req.Name != "" {
		args = append(args, req.Name)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}

	if req.MaxDuration != "" {
		_, err := time.ParseDuration(req.MaxDuration)
		if err != nil {
			return nil, fmt.Errorf("parse max_duration: %w", err)
		}
		args = append(args, req.MaxDuration)
		sets = append(sets, fmt.Sprintf("max_duration = $%d", len(args)))
	}

	if req.MaxMemory != "" {
		_, err := parseBytes(req.MaxMemory)
		if err != nil {
			return nil, fmt.Errorf("parse max_memory: %w", err)
		}
		args = append(args, req.MaxMemory)
		sets = append(sets, fmt.Sprintf("max_memory = $%d", len(args)))
	}

	if req.StdinSet {
		args = append(args, req.Stdin)
		sets = append(sets, fmt.Sprintf("stdin = $%d", len(args)))
	}

	switch testType {
	case mycode.TestType_simple:
		if req.ExpectedStdoutSet {
			args = append(args, req.ExpectedStdout)
			sets = append(sets, fmt.Sprintf("expected_stdout = $%d", len(args)))
		}
	case mycode.TestType_checker:
		if req.CheckerLanguageSet {
			args = append(args, req.CheckerLanguage)
			sets = append(sets, fmt.Sprintf("checker_language = $%d", len(args)))
		}
		if req.CheckerSource != "" {
			args = append(args, req.CheckerSource)
			sets = append(sets, fmt.Sprintf("checker_source = $%d", len(args)))
		}
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("nothing changed")
	}

	args = append(args, req.TestId)

	_, err = api.db.ExecContext(ctx, fmt.Sprintf(`
		update test set %s where id = $%d
	`, strings.Join(sets, ", "), len(args)), args...)
	if err != nil {
		return nil, fmt.Errorf("update test in DB: %w", err)
	}

	return &mycode.EditTestResp{}, nil
}

func (api *MyCodeAPI) RemoveTest(ctx context.Context,
	req *mycode.RemoveTestReq) (*mycode.RemoveTestResp, error) {

	if req.TestId == 0 {
		return nil, fmt.Errorf("empty test_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkTestBelongsToTeacher(ctx, req.TestId, t.Id)
	if err != nil {
		return nil, err
	}

	_, err = api.db.ExecContext(ctx, `
		delete from test where id = $1
	`, req.TestId)
	if err != nil {
		return nil, fmt.Errorf("delete test from DB: %w", err)
	}

	return &mycode.RemoveTestResp{}, nil
}

func (api *MyCodeAPI) GetTests(ctx context.Context,
	req *mycode.GetTestsReq) (*mycode.GetTestsResp, error) {

	ur, err := api.userRoleFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user role from context: %w", err)
	}

	var rows *sql.Rows

	switch ur {

	case ctxTeacher:
		if req.ExerciseId == 0 && req.StudentId == 0 {
			return nil, fmt.Errorf("both exercise_id and student_id empty")
		}

		t, err := api.teacherFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get teacher from context: %w", err)
		}

		if req.ExerciseId != 0 {
			err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
			if err != nil {
				return nil, err
			}
		} else {
			err = api.checkStudentBelongsToTeacher(ctx, req.StudentId, t.Id)
			if err != nil {
				return nil, err
			}

		}

	case ctxStudent:
		s, err := api.studentFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get student from context: %w", err)
		}

		if req.ExerciseId != 0 {
			err = api.checkExerciseBelongsToStudent(ctx, req.ExerciseId, s.Id)
			if err != nil {
				return nil, err
			}
		} else {
			req.StudentId = s.Id
		}

	default:
		return nil, fmt.Errorf("unexpected user role: %s", ur)
	}

	if req.ExerciseId != 0 {
		rows, err = api.db.QueryContext(ctx, `
			select id, exercise_id, type, name, max_duration, max_memory, stdin,
				expected_stdout, checker_language, checker_source
			from test
			where exercise_id = $1
		`, req.ExerciseId)
	} else {
		rows, err = api.db.QueryContext(ctx, `
			select t.id, t.exercise_id, t.type, t.name, t.max_duration,
				t.max_memory, t.stdin, t.expected_stdout, t.checker_language,
				t.checker_source
			from test as t
			join exercise e on t.exercise_id = e.id
			join student_exercise se on e.id = se.exercise_id
			where se.student_id = $1
		`, req.StudentId)
	}

	if err != nil {
		return nil, fmt.Errorf("get tests from DB: %w", err)
	}

	var ts []*mycode.Test

	for rows.Next() {
		t := &mycode.Test{}

		var (
			expectedStdout  sql.NullString
			checkerLanguage sql.NullInt32
			checkerSource   sql.NullString
		)

		err := rows.Scan(&t.Id, &t.ExerciseId, &t.Type, &t.Name,
			&t.MaxDuration, &t.MaxMemory, &t.Stdin, &expectedStdout,
			&checkerLanguage, &checkerSource)
		if err != nil {
			return nil, fmt.Errorf("get test row from DB: %w", err)
		}

		if expectedStdout.Valid {
			t.ExpectedStdout = expectedStdout.String
		}

		if checkerLanguage.Valid {
			t.CheckerLanguage = mycode.Language_name[checkerLanguage.Int32]
		}

		if checkerSource.Valid {
			t.CheckerSource = checkerSource.String
		}

		ts = append(ts, t)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("tests rows error: %w", rows.Err())
	}

	return &mycode.GetTestsResp{Tests: ts}, nil
}
