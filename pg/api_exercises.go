package pg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) GetExercise(ctx context.Context,
	req *mycode.GetExerciseReq) (*mycode.GetExerciseResp, error) {

	ur, err := api.userRoleFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user role from context: %w", err)
	}

	switch ur {
	case ctxTeacher:
		t, err := api.teacherFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get teacher from context: %w", err)
		}

		err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
		if err != nil {
			return nil, err
		}

	case ctxStudent:
		s, err := api.studentFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get student from context: %w", err)
		}

		err = api.checkExerciseBelongsToStudent(ctx, req.ExerciseId, s.Id)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unexpected user role: %s", ur)
	}

	e := &mycode.Exercise{}

	err = api.db.QueryRowContext(ctx, `
		select id, teacher_id, title, description, language, estimator
		from exercise where id = $1
	`, req.ExerciseId).Scan(&e.Id, &e.TeacherId, &e.Title, &e.Description,
		&e.Language, &e.Estimator)
	if err != nil {
		return nil, fmt.Errorf("get exercise from DB: %w", err)
	}

	return &mycode.GetExerciseResp{Exercise: e}, nil
}

func (api *MyCodeAPI) AddExercise(ctx context.Context,
	req *mycode.AddExerciseReq) (*mycode.AddExerciseResp, error) {

	if req.Title == "" {
		return nil, fmt.Errorf("empty title")
	}

	if req.Description == "" {
		return nil, fmt.Errorf("empty text")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	var id int64

	err = api.db.QueryRowContext(ctx, `
		insert into exercise (
			teacher_id, title, description, language, estimator)
		values ($1, $2, $3, $4, $5)
		returning id
	`, t.Id, req.Title, req.Description, req.Language, req.Estimator).Scan(&id)

	return &mycode.AddExerciseResp{
		ExerciseId: id,
	}, nil
}

func (api *MyCodeAPI) checkExerciseBelongsToTeacher(ctx context.Context,
	exerciseID, teacherID int64) error {

	var exerciseTeacherID int64

	err := api.db.QueryRowContext(ctx, `
		select teacher_id from exercise where id = $1
	`, exerciseID).Scan(&exerciseTeacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("exercise doesn't exists")
		}
		return fmt.Errorf("check exercise belongs to teacher: %w", err)
	}

	if teacherID != exerciseTeacherID {
		return fmt.Errorf("exercise doesn't belongs to teacher")
	}

	return nil
}

func (api *MyCodeAPI) EditExercise(ctx context.Context,
	req *mycode.EditExerciseReq) (*mycode.EditExerciseResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
	if err != nil {
		return nil, err
	}

	var (
		args []interface{}
		sets []string
	)

	if req.Title != "" {
		args = append(args, req.Title)
		sets = append(sets, fmt.Sprintf("title = $%d", len(args)))
	}

	if req.Description != "" {
		args = append(args, req.Description)
		sets = append(sets, fmt.Sprintf("description = $%d", len(args)))
	}

	if req.LanguageSet {
		args = append(args, req.Language)
		sets = append(sets, fmt.Sprintf("language = $%d", len(args)))
	}

	if req.EstimatorSet {
		args = append(args, req.Estimator)
		sets = append(sets, fmt.Sprintf("estimator = $%d", len(args)))
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("nothing changed")
	}

	args = append(args, req.ExerciseId)

	_, err = api.db.ExecContext(ctx, fmt.Sprintf(`
		update exercise set %s where id = $%d 
	`, strings.Join(sets, ", "), len(args)), args...)
	if err != nil {
		return nil, fmt.Errorf("update exercise in DB: %w", err)
	}

	return &mycode.EditExerciseResp{}, nil
}

func (api *MyCodeAPI) RemoveExercise(ctx context.Context,
	req *mycode.RemoveExerciseReq) (*mycode.RemoveExerciseResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
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
			return nil, fmt.Errorf("exercise doesn't exists")
		}
	}

	if t.Id != exerciseTeacherID {
		return nil, fmt.Errorf("exercise doesn't belongs to teacher")
	}

	_, err = api.db.ExecContext(ctx, `
		delete from exercise where id = $1
	`, req.ExerciseId)
	if err != nil {
		return nil, fmt.Errorf("delete exercise from DB: %w", err)
	}

	return &mycode.RemoveExerciseResp{}, nil
}

func (api *MyCodeAPI) GetExercises(ctx context.Context,
	req *mycode.GetExercisesReq) (*mycode.GetExercisesResp, error) {

	ur, err := api.userRoleFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user role from context: %w", err)
	}

	var rows *sql.Rows

	switch ur {
	case ctxTeacher:
		t, err := api.teacherFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get teacher from context: %w", err)
		}

		if req.StudentId != 0 {
			rows, err = api.db.QueryContext(ctx, `
				select e.id, e.teacher_id, e.title, e.description,
					e.language, e.estimator
				from exercise as e
				join student_exercise as se on e.id = se.exercise_id
				where e.teacher_id = $1 and se.student_id = $2
			`, t.Id, req.StudentId)
		} else {
			rows, err = api.db.QueryContext(ctx, `
				select id, teacher_id, title, description, language,
					estimator
				from exercise
				where teacher_id = $1
			`, t.Id)
		}
		if err != nil {
			return nil, fmt.Errorf("get exercises from DB: %w", err)
		}

	case ctxStudent:
		s, err := api.studentFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("get student from context: %w", err)
		}

		rows, err = api.db.QueryContext(ctx, `
			select e.id, e.teacher_id, e.title, e.description, e.language,
				e.estimator
			from student_exercise as se
			join exercise as e on se.exercise_id = e.id
			where se.student_id = $1
		`, s.Id)
		if err != nil {
			return nil, fmt.Errorf("get exercises from DB: %w", err)
		}

	default:
		return nil, fmt.Errorf("unexpected user role: %s", ur)
	}

	var es []*mycode.Exercise

	for rows.Next() {
		e := &mycode.Exercise{}
		err := rows.Scan(&e.Id, &e.TeacherId, &e.Title, &e.Description,
			&e.Language, &e.Estimator)
		if err != nil {
			return nil, fmt.Errorf("get exercise row from DB: %w", err)
		}
		es = append(es, e)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("exercises rows error: %w", rows.Err())
	}

	return &mycode.GetExercisesResp{Exercises: es}, nil
}

func (api *MyCodeAPI) checkClassBelongsToTeacher(
	ctx context.Context, classID, teacherID int64) error {

	var classTeacherID int64

	err := api.db.QueryRowContext(ctx, `
		select teacher_id from class where id = $1
	`, classID).Scan(&classTeacherID)
	if err != nil {
		return fmt.Errorf("get class teacher ID: %w", err)
	}

	if teacherID != classTeacherID {
		return fmt.Errorf("class doesn't belongs to teacher: %w", err)
	}

	return nil
}

func (api *MyCodeAPI) checkStudentBelongsToTeacher(
	ctx context.Context, studentID, teacherID int64) error {

	var classTeacherID int64

	err := api.db.QueryRowContext(ctx, `
		select teacher_id from student as s
		join class as c on s.class_id = c.id
		where s.id = $1
	`, studentID).Scan(&classTeacherID)
	if err != nil {
		return fmt.Errorf("get student teacher ID: %w", err)
	}

	if teacherID != classTeacherID {
		return fmt.Errorf("student doesn't belongs to teacher: %w", err)
	}

	return nil
}

func (api *MyCodeAPI) GetExerciseAssignments(ctx context.Context,
	req *mycode.GetExerciseAssignmentsReq) (
	*mycode.GetExerciseAssignmentsResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
	if err != nil {
		return nil, err
	}

	rows, err := api.db.QueryContext(ctx, `
		select student_id from student_exercise where exercise_id = $1 
	`, req.ExerciseId)
	if err != nil {
		return nil, fmt.Errorf("get student_exercises from DB: %w", err)
	}

	var ids []int64

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf(
				"get student_exercise row from DB: %w", err)
		}
		ids = append(ids, id)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("student_exercises row error: %w", err)
	}

	return &mycode.GetExerciseAssignmentsResp{StudentIds: ids}, nil
}

func (api *MyCodeAPI) AssignExercise(ctx context.Context,
	req *mycode.AssignExerciseReq) (*mycode.AssignExerciseResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
	if err != nil {
		return nil, err
	}

	switch {
	case req.ClassId != 0:
		err = api.checkClassBelongsToTeacher(ctx, req.ClassId, t.Id)
		if err != nil {
			return nil, err
		}

		_, err = api.db.ExecContext(ctx, `
			insert into student_exercise (student_id, exercise_id) 
				select id, $1 from student where class_id = $2
		`, req.ExerciseId, req.ClassId)
		if err != nil {
			return nil, fmt.Errorf(
				"add students exercises to DB: %w", err)
		}

	case req.StudentId != 0:
		err = api.checkStudentBelongsToTeacher(ctx, req.StudentId, t.Id)
		if err != nil {
			return nil, err
		}

		_, err = api.db.ExecContext(ctx, `
			insert into student_exercise (student_id, exercise_id)
			values ($1, $2)
		`, req.StudentId, req.ExerciseId)
		if err != nil {
			return nil, fmt.Errorf("add student exercise to DB: %w", err)
		}

	default:
		return nil, fmt.Errorf("both class_id and student_id are empty")
	}

	return &mycode.AssignExerciseResp{}, nil
}

func (api *MyCodeAPI) WithdrawExercise(ctx context.Context,
	req *mycode.WithdrawExerciseReq) (*mycode.WithdrawExerciseResp, error) {

	if req.ExerciseId == 0 {
		return nil, fmt.Errorf("empty exercise_id")
	}

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	err = api.checkExerciseBelongsToTeacher(ctx, req.ExerciseId, t.Id)
	if err != nil {
		return nil, err
	}

	switch {
	case req.ClassId != 0:
		err = api.checkClassBelongsToTeacher(ctx, req.ClassId, t.Id)
		if err != nil {
			return nil, err
		}

		_, err = api.db.ExecContext(ctx, `
			delete from student_exercise
			where exercise_id = $1 and student_id in (
			    select id from student where class_id = $2)
		`, req.ExerciseId, req.ClassId)
		if err != nil {
			return nil, fmt.Errorf(
				"remove students exercises to DB: %w", err)
		}

	case req.StudentId != 0:
		err = api.checkStudentBelongsToTeacher(ctx, req.StudentId, t.Id)
		if err != nil {
			return nil, err
		}

		_, err = api.db.ExecContext(ctx, `
			delete from student_exercise
			where exercise_id = $1 and student_id = $2 
		`, req.ExerciseId, req.StudentId)
		if err != nil {
			return nil, fmt.Errorf(
				"remove student exercise to DB: %w", err)
		}

	default:
		return nil, fmt.Errorf("both class_id and student_id are empty")
	}

	return &mycode.WithdrawExerciseResp{}, nil
}
