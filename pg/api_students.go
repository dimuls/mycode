package pg

import (
	"context"
	"fmt"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) GetStudents(ctx context.Context,
	req *mycode.GetStudentsReq) (*mycode.GetStudentsResp, error) {

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	rows, err := api.db.QueryContext(ctx, `
		select s.id, s.user_id, s.class_id, s.name from student as s
		join class as c on s.class_id = c.id
		where c.teacher_id = $1
	`, t.Id)
	if err != nil {
		return nil, fmt.Errorf("get students from DB: %w", err)
	}

	var ss []*mycode.Student

	for rows.Next() {
		s := &mycode.Student{}
		err := rows.Scan(&s.Id, &s.UserId, &s.ClassId, &s.Name)
		if err != nil {
			return nil, fmt.Errorf("get student row from DB: %w", err)
		}
		ss = append(ss, s)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("students rows error: %w", rows.Err())
	}

	return &mycode.GetStudentsResp{Students: ss}, nil
}
