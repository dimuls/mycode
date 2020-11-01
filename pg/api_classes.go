package pg

import (
	"context"
	"fmt"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) GetClasses(ctx context.Context,
	req *mycode.GetClassesReq) (*mycode.GetClassesResp, error) {

	t, err := api.teacherFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teacher from context: %w", err)
	}

	rows, err := api.db.QueryContext(ctx, `
		select id, teacher_id, name from class where teacher_id = $1
	`, t.Id)
	if err != nil {
		return nil, fmt.Errorf("get classes from DB: %w", err)
	}

	var cs []*mycode.Class

	for rows.Next() {
		c := &mycode.Class{}
		err := rows.Scan(&c.Id, &c.TeacherId, &c.Name)
		if err != nil {
			return nil, fmt.Errorf("get class row from DB: %w", err)
		}
		cs = append(cs, c)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("classes rows error: %w", rows.Err())
	}

	return &mycode.GetClassesResp{Classes: cs}, nil
}
