create table "user" (
    id bigserial primary key,
    login text not null unique,
    password_hash bytea not null
);

create table teacher (
    id bigserial primary key,
    user_id bigint not null references "user" (id) on delete cascade,
    name text not null
);

create index on teacher (user_id);

create table class (
    id bigserial primary key,
    teacher_id bigint not null references teacher (id) on delete cascade,
    name text not null
);

create index on class (teacher_id);

create table student (
    id bigserial primary key,
    user_id bigint not null references "user" (id) on delete cascade,
    class_id bigint not null references class (id) on delete cascade,
    name text not null
);

create index on student (user_id);
create index on student (class_id);

create table exercise (
    id bigserial primary key,
    teacher_id bigint not null references teacher (id) on delete cascade,
    title text not null,
    description text not null,
    language int not null,
    estimator int not null
);

create index on exercise (teacher_id);

create table test (
    id bigserial primary key,
    exercise_id bigint not null references exercise(id) on delete cascade,
    type text not null,
    name text not null,
    max_duration text not null,
    max_memory text not null,
    stdin text not null,
    expected_stdout text,
    checker_language int,
    checker_source text
);

create index on test (exercise_id);

create table student_exercise (
    student_id bigint not null references student (id) on delete cascade,
    exercise_id bigint not null references exercise (id) on delete cascade,

    primary key (student_id, exercise_id)
);

create index on student_exercise (exercise_id, student_id);

create table solution (
    id bigserial primary key,
    student_id bigint not null references student (id) on delete cascade,
    exercise_id bigint not null references exercise (id) on delete cascade,
    source text not null,

    foreign key (student_id, exercise_id) references
        student_exercise (student_id, exercise_id) on delete cascade
);

create index on solution (student_id);
create index on solution (exercise_id);

create table solution_test (
    id bigserial primary key,
    solution_id bigint not null references solution (id) on delete cascade,
    test_id bigint not null references test (id) on delete cascade,
    status text not null,
    duration text,
    used_memory text,
    stdout text,
    stderr text,
    checker_stdout text,
    checker_stderr text,
    fails jsonb
);

create index on solution_test (solution_id);
create index on solution_test (test_id);
