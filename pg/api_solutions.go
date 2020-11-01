package pg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) checkExerciseBelongsToStudent(ctx context.Context,
	exerciseID, studentID int64) error {

	var exist bool

	err := api.db.QueryRowContext(ctx, `
		select exists (
		    select 1 from student_exercise
		    where exercise_id = $1 and student_id = $2
		)
	`, exerciseID, studentID).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("exercise doesn't exists")
		}
		return fmt.Errorf("check exercise belongs to student: %w", err)
	}

	if !exist {
		return fmt.Errorf("exercise doesn't belongs to student")
	}

	return nil
}

func (api *MyCodeAPI) AddSolution(ctx context.Context,
	req *mycode.AddSolutionReq) (resp *mycode.AddSolutionResp, err error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	if req.Source == "" {
		return nil, fmt.Errorf("empty source")
	}

	s, err := api.studentFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get student from context: %w", err)
	}

	err = api.checkExerciseBelongsToStudent(ctx, req.ExerciseId, s.Id)
	if err != nil {
		return nil, err
	}

	tx, err := api.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err2 = fmt.Errorf("%w, failed to rollback: %v", err, err2)
			}
		}
	}()

	var solutionID int64

	err = tx.QueryRowContext(ctx, `
		insert into solution (student_id, exercise_id, source)
		values ($1, $2, $3)
		returning id
	`, s.Id, req.ExerciseId, req.Source).Scan(&solutionID)
	if err != nil {
		return nil, fmt.Errorf("add solution to DB: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		insert into solution_test (solution_id, test_id, status)
		select $1, id, $2 from test where exercise_id = $3
	`, solutionID, mycode.SolutionTestStatus_processing, req.ExerciseId)
	if err != nil {
		return nil, fmt.Errorf("add solution test to DB: %w", err)
	}

	rowsAdded, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get added solution_test count: %w", err)
	}

	if rowsAdded == 0 {
		return nil, fmt.Errorf("exercise do not have tests")
	}

	rows, err := tx.Query(`
				select st.id, e.language, s.source, t.type,
					t.stdin, t.checker_language, t.checker_source
                from solution_test as st
                join test t on st.test_id = t.id
                join solution s on st.solution_id = s.id
                join exercise e on s.exercise_id = e.id
                where st.solution_id = $1
        `, solutionID)
	if err != nil {
		api.log.WithError(err).Error("failed to get solutions from DB")
	}

	for rows.Next() {
		var (
			solutionTestID  int64
			language        mycode.Language
			source          string
			testType        mycode.TestType
			stdin           string
			checkerLanguage sql.NullInt32
			checkerSource   sql.NullString
		)
		err := rows.Scan(&solutionTestID, &language, &source, &testType,
			&stdin, &checkerLanguage, &checkerSource)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get solution row from DB: %w", err)
		}

		err = api.codePublisher.PublishCode(&mycode.Code{
			SolutionTestId:  solutionTestID,
			Language:        language,
			Source:          source,
			Stdin:           stdin,
			CheckerLanguage: mycode.Language(checkerLanguage.Int32),
			CheckerSource:   checkerSource.String,
			WithChecker:     testType == mycode.TestType_checker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to publish code: %w", err)
		}
	}

	if rows.Err() != nil {
		api.log.WithError(rows.Err()).Error("got solutions test rows error")
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit changes to DB: %w", err)
	}

	return &mycode.AddSolutionResp{SolutionId: solutionID}, nil
}

func (api *MyCodeAPI) GetSolutions(ctx context.Context,
	req *mycode.GetSolutionsReq) (*mycode.GetSolutionsResp, error) {

	ur, err := api.userRoleFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user role from context: %w", err)
	}

	switch ur {
	case ctxTeacher:
		if req.StudentId == 0 {
			return nil, fmt.Errorf("empty student_id")
		}

		t, err := api.teacherFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get teacher from context: %w", err)
		}

		err = api.checkStudentBelongsToTeacher(ctx, req.StudentId, t.Id)
		if err != nil {
			return nil, err
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
		}

		req.StudentId = s.Id

	default:
		return nil, fmt.Errorf("unexpected user role: %s", ur)
	}

	var (
		wheres []string
		args   []interface{}
	)

	args = append(args, req.StudentId)
	wheres = append(wheres, fmt.Sprintf("student_id = $%d", len(args)))

	if req.ExerciseId != 0 {
		args = append(args, req.ExerciseId)
		wheres = append(wheres, fmt.Sprintf("exercise_id = $%d", len(args)))
	}

	rows, err := api.db.QueryContext(ctx, fmt.Sprintf(`
			select id, student_id, exercise_id, source 
			from solution
			where %s
		`, strings.Join(wheres, " and ")), args...)
	if err != nil {
		return nil, fmt.Errorf("get solutions error: %w", err)
	}

	var ss []*mycode.Solution

	for rows.Next() {
		s := &mycode.Solution{}
		err = rows.Scan(&s.Id, &s.StudentId, &s.ExerciseId, &s.Source)
		if err != nil {
			return nil, fmt.Errorf(
				"get solution row from DB: %w", err)
		}
		ss = append(ss, s)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("solutions rows error: %w", err)
	}

	return &mycode.GetSolutionsResp{Solutions: ss}, nil
}
