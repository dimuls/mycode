package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dimuls/mycode"
)

func (api *MyCodeAPI) HandleRun(ctx context.Context, r *mycode.Run) (err error) {

	log := api.log.WithField("solution_test_id", r.SolutionTestId)

	log.Info("run received")

	defer func() {
		if err != nil {
			log.WithError(err).Error("failed to process run")
		} else {
			log.Info("run processed")
		}
	}()

	t := &mycode.Test{}

	var expectedStdout sql.NullString

	err = api.db.QueryRowContext(ctx, `
		select t.type, t.max_duration, t.max_memory, t.expected_stdout
		from solution_test as st
		join test as t on st.test_id = t.id
		where st.id = $1 
	`, r.SolutionTestId).Scan(&t.Type, &t.MaxDuration, &t.MaxMemory,
		&expectedStdout)
	if err != nil {
		return fmt.Errorf("get test from DB: %w", err)
	}

	t.ExpectedStdout = expectedStdout.String

	runDuration, err := time.ParseDuration(r.Duration)
	if err != nil {
		return fmt.Errorf("parse run duration: %w", err)
	}

	maxDuration, err := time.ParseDuration(t.MaxDuration)
	if err != nil {
		return fmt.Errorf("parse test duration: %w", err)
	}

	runUsedMemory, err := parseBytes(r.UsedMemory)
	if err != nil {
		return fmt.Errorf("parse run used memory: %w", err)
	}

	maxMemory, err := parseBytes(t.MaxMemory)
	if err != nil {
		return fmt.Errorf("parse max memory: %w", err)
	}

	fails := &mycode.SolutionTestFails{
		WrongDuration:   runDuration > maxDuration,
		WrongUsedMemory: runUsedMemory > maxMemory,
		WrongStdout:     t.Type == mycode.TestType_simple && r.Stdout != t.ExpectedStdout,
		WrongChecker:    t.Type == mycode.TestType_checker && r.CheckerStdout != "ok",
	}

	failsJSON, err := json.Marshal(fails)
	if err != nil {
		return fmt.Errorf("JSON marshal fails: %w", err)
	}

	var status mycode.SolutionTestStatus

	failed := fails.WrongDuration || fails.WrongUsedMemory ||
		fails.WrongStdout || fails.WrongChecker || r.Stderr != ""

	var failsJSONStr sql.NullString

	if failed {
		status = mycode.SolutionTestStatus_failed
		failsJSONStr.String = string(failsJSON)
		failsJSONStr.Valid = true
	} else {
		status = mycode.SolutionTestStatus_succeed
		failsJSONStr.Valid = false
	}

	switch t.Type {
	case mycode.TestType_simple:
		_, err = api.db.ExecContext(ctx, `
			update solution_test set status = $1, duration = $2,
				used_memory = $3, stdout = $4, stderr = $5,
				fails = $6
			where id = $7
		`, status, r.Duration, r.UsedMemory, r.Stdout, r.Stderr,
			failsJSONStr, r.SolutionTestId)
	case mycode.TestType_checker:
		_, err = api.db.ExecContext(ctx, `
			update solution_test set status = $1, duration = $2,
				used_memory = $3, stdout = $4, stderr = $5,
				checker_stdout = $6, checker_stderr = $7,
				fails = $8
			where id = $9
		`, status, r.Duration, r.UsedMemory, r.Stdout, r.Stderr,
			r.CheckerStdout, r.CheckerStderr, failsJSONStr,
			r.SolutionTestId)
	}
	if err != nil {
		return fmt.Errorf("add solution test result to DB: %w",
			err)
	}

	return nil
}
