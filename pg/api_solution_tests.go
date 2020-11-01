package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) checkSolutionBelongsToTeacher(ctx context.Context,
	solutionID, teacherID int64) error {

	var solutionTeacherID int64

	err := api.db.QueryRowContext(ctx, `
			select c.teacher_id from solution as s
			join student as st on s.student_id = st.id
			join class as c on st.class_id = c.id
			where s.id = $1
		`, solutionID).Scan(&solutionTeacherID)
	if err != nil {
		return fmt.Errorf(
			"check solution belongs to teacher: %w", err)
	}

	if teacherID != solutionTeacherID {
		return fmt.Errorf("solution doesn't belongs to teacher")
	}

	return nil
}

func (api *MyCodeAPI) checkSolutionBelongsToStudent(ctx context.Context,
	solutionID, studentID int64) error {

	var solutionStudentID int64

	err := api.db.QueryRowContext(ctx, `
			select student_id from solution
			where id = $1
		`, solutionID).Scan(&solutionStudentID)
	if err != nil {
		return fmt.Errorf(
			"check solution belongs to student: %w", err)
	}

	if studentID != solutionStudentID {
		return fmt.Errorf("solution doesn't belongs to student")
	}

	return nil
}

func (api *MyCodeAPI) GetSolutionTests(ctx context.Context,
	req *mycode.GetSolutionTestsReq) (*mycode.GetSolutionTestsResp, error) {

	ur, err := api.userRoleFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user role from context: %w", err)
	}

	switch ur {
	case ctxTeacher:
		if req.StudentId == 0 && req.SolutionId == 0 {
			return nil, fmt.Errorf(
				"neither student_id not solution_id defined")
		}

		t, err := api.teacherFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get teacher from context: %w", err)
		}

		if req.StudentId == 0 {
			err = api.checkSolutionBelongsToTeacher(ctx, req.SolutionId, t.Id)
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

		if req.SolutionId != 0 {
			err = api.checkSolutionBelongsToStudent(ctx, req.SolutionId, s.Id)
			if err != nil {
				return nil, err
			}
		}

		req.StudentId = s.Id

	default:
		return nil, fmt.Errorf("unexpected user role: %s", ur)
	}

	var sts []*mycode.SolutionTest

	var rows *sql.Rows

	if req.SolutionId != 0 {
		rows, err = api.db.QueryContext(ctx, `
			select id, solution_id, test_id, status, duration, used_memory,
				   stdout, stderr, checker_stdout, checker_stderr, fails
			from solution_test
			where solution_id = $1
		`, req.SolutionId)
		if err != nil {
			return nil, fmt.Errorf("get solution tests from DB: %w", err)
		}
	} else {
		rows, err = api.db.QueryContext(ctx, `
			select st.id, st.solution_id, st.test_id, st.status, st.duration,
				   st.used_memory, st.stdout, st.stderr,
				   st.checker_stdout, st.checker_stderr, st.fails
			from solution_test as st
			join solution s on s.id = st.solution_id
			where s.student_id = $1
		`, req.StudentId)
		if err != nil {
			return nil, fmt.Errorf("get solution tests from DB: %w", err)
		}
	}

	for rows.Next() {
		var (
			st                                   = &mycode.SolutionTest{}
			duration, usedMemory, stdout, stderr sql.NullString
			checkerStdout, checkerStderr         sql.NullString
			failsJSON                            []byte
		)
		err = rows.Scan(&st.Id, &st.SolutionId, &st.TestId, &st.Status,
			&duration, &usedMemory, &stdout, &stderr,
			&checkerStdout, &checkerStderr, &failsJSON)

		if duration.Valid {
			st.Duration = duration.String
		}

		if usedMemory.Valid {
			st.UsedMemory = usedMemory.String
		}

		if stdout.Valid {
			st.Stdout = stdout.String
		}

		if stderr.Valid {
			st.Stderr = stderr.String
		}

		if checkerStdout.Valid {
			st.CheckerStdout = checkerStdout.String
		}

		if checkerStderr.Valid {
			st.CheckerStderr = checkerStderr.String
		}

		st.Fails = &mycode.SolutionTestFails{}

		if failsJSON != nil {
			err := json.Unmarshal(failsJSON, st.Fails)
			if err != nil {
				return nil, fmt.Errorf("JSON unmarshal fails: %w", err)
			}

		}

		sts = append(sts, st)
	}

	return &mycode.GetSolutionTestsResp{SolutionTests: sts}, nil
}
