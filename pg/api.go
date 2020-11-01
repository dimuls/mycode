package pg

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	rice "github.com/GeertJohan/go.rice"
	migraterice "github.com/atrox/go-migrate-rice"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"

	"github.com/dimuls/mycode"
)

type CodePublisher interface {
	PublishCode(c *mycode.Code) error
}

type MyCodeAPI struct {
	jwtSecret     string
	db            *sql.DB
	codePublisher CodePublisher
	stop          chan struct{}
	wg            sync.WaitGroup
	log           *logrus.Entry
}

func NewMyCodeAPI(pgURI, jwtSecret string, cp CodePublisher) (*MyCodeAPI, error) {
	db, err := sql.Open("postgres", pgURI)
	if err != nil {
		return nil, err
	}

	return &MyCodeAPI{
		jwtSecret:     jwtSecret,
		db:            db,
		codePublisher: cp,
		stop:          make(chan struct{}),
		log:           logrus.WithField("subsystem", "pg_my_code_api"),
	}, nil
}

func (api *MyCodeAPI) Close() error {
	close(api.stop)
	api.wg.Wait()
	return api.db.Close()
}

//go:generate rice embed-go

func (api *MyCodeAPI) Migrate() error {
	migrationsBox := rice.MustFindBox("migrations")

	sourceDriver, err := migraterice.WithInstance(migrationsBox)
	if err != nil {
		return fmt.Errorf("create rice source instance: %w", err)
	}

	dbDriver, err := postgres.WithInstance(api.db, &postgres.Config{
		MigrationsTable: "migration",
	})
	if err != nil {
		return fmt.Errorf("create database instance: %w", err)
	}

	m, err := migrate.NewWithInstance("rice", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (api *MyCodeAPI) userRoleFromContext(ctx context.Context) (string, error) {
	ri := ctx.Value(ctxUserRole)
	if ri == nil {
		return "", fmt.Errorf("not found")
	}

	r, ok := ri.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type: %T", ri)
	}

	return r, nil
}

func (api *MyCodeAPI) teacherFromContext(ctx context.Context) (
	*mycode.Teacher, error) {

	ti := ctx.Value(ctxTeacher)
	if ti == nil {
		return nil, fmt.Errorf("not found")
	}

	t, ok := ti.(*mycode.Teacher)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", ti)
	}

	return t, nil
}

func (api *MyCodeAPI) studentFromContext(ctx context.Context) (
	*mycode.Student, error) {

	si := ctx.Value(ctxStudent)
	if si == nil {
		return nil, fmt.Errorf("not found")
	}

	s, ok := si.(*mycode.Student)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", si)
	}

	return s, nil
}
