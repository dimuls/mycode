package main

import (
	"database/sql"
	"flag"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/e154/vydumschik"
	"github.com/gosimple/slug"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	classNumbers = "7 8 9 10 11"
	classLetters = "А Б В Г Д"

	// password: password
	passwordHash = "$2a$04$1ssgHXFmtMWAPl2vhc8rse66YR0CTpSpIhVhlaeTBtFHC5hwzZzCG"
)

type student struct {
	Name  string
	Class string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var (
		pgURI         string
		teacherID     int64
		studentsCount int
	)

	flag.StringVar(&pgURI, "p", "", "postgres URI")
	flag.Int64Var(&teacherID, "t", 0, "teacher ID of generating classes")
	flag.IntVar(&studentsCount, "c", 0, "students count to generate")

	flag.Parse()

	if pgURI == "" || studentsCount == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := sql.Open("postgres", pgURI)
	if err != nil {
		logrus.WithError(err).Fatal("failed to open DB connection")
	}

	gs := []string{"male", "female"}

	var allCs []string

	for _, n := range strings.Split(classNumbers, " ") {
		for _, l := range strings.Split(classLetters, " ") {
			allCs = append(allCs, n+l)
		}
	}

	defer db.Close()

	ng := &vydumschik.Name{}

	var ss []student
	cs := map[string]int64{}

	for i := 0; i < studentsCount; i++ {
		gender := gs[rand.Intn(2)]
		name := ng.Full_name(gender)
		class := allCs[rand.Int63n(int64(len(allCs)))]

		cs[class] = 0

		ss = append(ss, student{
			Name:  name,
			Class: class,
		})
	}

	for c := range cs {
		var classID int64
		err := db.QueryRow(`
			insert into class (teacher_id, name) values ($1, $2)
			on conflict do nothing
			returning id
		`, teacherID, c).Scan(&classID)
		if err != nil {
			logrus.WithError(err).Fatal("failed to add class")
		}
		cs[c] = classID
	}

	for _, s := range ss {

		login := slug.Make(s.Name)

		var userID int64

		err := db.QueryRow(`
			insert into "user" (login, password_hash) values ($1, $2)
			on conflict do nothing
			returning id
		`, login, []byte(passwordHash)).Scan(&userID)
		if err != nil {
			logrus.WithError(err).Fatal("failed to add user")
		}

		_, err = db.Exec(`
			insert into student (user_id, class_id, name) values ($1, $2, $3)
			on conflict do nothing;
		`, userID, cs[s.Class], s.Name)
		if err != nil {
			logrus.WithError(err).Fatal("failed to add student")
		}
	}
}
