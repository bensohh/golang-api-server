-- DROP TABLE STATEMENTS
DROP TABLE IF EXISTS registries;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS teachers;

-- FUNCTION STATEMENTS
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- CREATE TABLE STATEMENTS
CREATE TABLE IF NOT EXISTS teachers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    suspended INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS registries (
    id SERIAL,
    teacher_email VARCHAR(255),
    student_email VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (teacher_email, student_email),
    FOREIGN KEY (teacher_email) REFERENCES teachers(email),
    FOREIGN KEY (student_email) REFERENCES students(email)
);

-- CREATE TRIGGERS STATEMENTS
CREATE TRIGGER set_timestamp_teachers
BEFORE UPDATE ON teachers
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_students
BEFORE UPDATE ON students
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_registries
BEFORE UPDATE ON registries
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- INSERT STATEMENTS
INSERT INTO teachers (name, email) VALUES ('Ken', 'teacherken@gmail.com');
INSERT INTO teachers (name, email) VALUES ('Joe', 'teacherjoe@gmail.com');

INSERT INTO students (name, email) VALUES ('Jon', 'studentjon@gmail.com');
INSERT INTO students (name, email) VALUES ('Hon', 'studenthon@gmail.com');
INSERT INTO students (name, email) VALUES ('Tom', 'studenttom@gmail.com');
INSERT INTO students (name, email) VALUES ('Tom', 'studentunderkenonly@gmail.com');

INSERT INTO registries (teacher_email, student_email) VALUES ('teacherjoe@gmail.com', 'studentjon@gmail.com');
INSERT INTO registries (teacher_email, student_email) VALUES ('teacherjoe@gmail.com', 'studenthon@gmail.com');