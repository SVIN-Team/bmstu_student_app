-- =====================================================
-- Расширения
-- =====================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- Типы
-- =====================================================

CREATE TYPE user_role AS ENUM (
    'student',
    'headman',
    'admin'
);

CREATE TYPE lesson_type AS ENUM (
    'lecture',
    'lab',
    'seminar'
);

CREATE TYPE queue_status AS ENUM (
    'draft',
    'open',
    'closed',
    'archived'
);

CREATE TYPE queue_slot_status AS ENUM (
    'waiting',
    'passed',
    'failed',
    'no_show'
);

-- =====================================================
-- Таблицы
-- =====================================================

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    headman_id UUID
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(60),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role user_role NOT NULL DEFAULT 'student',
    group_id UUID,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT fk_users_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE SET NULL,

    CONSTRAINT uq_users_id_group UNIQUE (id, group_id)
);

ALTER TABLE groups
ADD CONSTRAINT fk_groups_headman_same_group
FOREIGN KEY (headman_id, id)
REFERENCES users(id, group_id);

CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL UNIQUE
);

CREATE TABLE teachers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE
);

CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    teacher_id UUID NOT NULL,
    room_id UUID,
    type lesson_type NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_lessons_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_lessons_subject
        FOREIGN KEY (subject_id)
        REFERENCES subjects(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_lessons_teacher
        FOREIGN KEY (teacher_id)
        REFERENCES teachers(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_lessons_room
        FOREIGN KEY (room_id)
        REFERENCES rooms(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_lessons_time
        CHECK (ends_at > starts_at)
);

CREATE TABLE queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    lesson_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    opens_at TIMESTAMP NOT NULL,
    closes_at TIMESTAMP,
    max_size INT,
    status queue_status NOT NULL DEFAULT 'draft',
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT fk_queues_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_queues_subject
        FOREIGN KEY (subject_id)
        REFERENCES subjects(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_queues_lesson
        FOREIGN KEY (lesson_id)
        REFERENCES lessons(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_queues_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE TABLE queue_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL,
    student_id UUID NOT NULL,
    status queue_slot_status NOT NULL DEFAULT 'waiting',
    signed_up_at TIMESTAMP NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT fk_queue_slots_queue
        FOREIGN KEY (queue_id)
        REFERENCES queues(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_queue_slots_student
        FOREIGN KEY (student_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_queue_student
        UNIQUE (queue_id, student_id)
);

-- =====================================================
-- Индексы
-- =====================================================

CREATE INDEX idx_slots_queue_timestamp
ON queue_slots (queue_id, signed_up_at);

CREATE INDEX idx_queues_group_status
ON queues (group_id, status);

CREATE INDEX idx_lessons_group_date
ON lessons (group_id, starts_at);