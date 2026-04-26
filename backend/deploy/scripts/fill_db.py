#!/usr/bin/env python3
"""
Скрипт наполнения БД учебного приложения реалистичными данными.
Адаптирован для psycopg 3 (не требует компиляции).
"""

import os
import random
import uuid
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional

import bcrypt
import psycopg
from psycopg import sql
from psycopg.rows import dict_row
from faker import Faker

# =====================================================
# Конфигурация
# =====================================================

DB_CONFIG = {
    'dbname': os.environ.get('STUDHUB_POSTGRES_DB', 'studhub_db'),
    'user': os.environ.get('STUDHUB_POSTGRES_USER', 'studhub_admin'),
    'password': os.environ.get('STUDHUB_POSTGRES_PASSWORD', 'SuperSecurePassword1'),
    'host': os.environ.get('DB_HOST', 'localhost'),
    'port': os.environ.get('DB_PORT', '5432'),
}

GROUPS_COUNT = 4
STUDENTS_PER_GROUP_MIN = 15
STUDENTS_PER_GROUP_MAX = 25
SUBJECTS_COUNT = 12
TEACHERS_COUNT = 10
ROOMS_COUNT = 15

WEEKDAYS = [0, 1, 2, 3, 4]
LESSON_DURATION = timedelta(hours=1, minutes=30)
DAY_START = datetime.strptime("09:00", "%H:%M").time()
DAY_END = datetime.strptime("18:00", "%H:%M").time()

WEEKS_AHEAD = 3

PASSED_PROB = 0.7
FAILED_PROB = 0.2
NO_SHOW_PROB = 0.1

DEFAULT_PASSWORD = "password"

# =====================================================
# Вспомогательные функции
# =====================================================

fake = Faker("ru_RU")

def hash_password(password: str) -> str:
    salt = bcrypt.gensalt()
    return bcrypt.hashpw(password.encode('utf-8'), salt).decode('utf-8')

def transliterate(text: str) -> str:
    mapping = {
        'а': 'a', 'б': 'b', 'в': 'v', 'г': 'g', 'д': 'd', 'е': 'e', 'ё': 'e',
        'ж': 'zh', 'з': 'z', 'и': 'i', 'й': 'y', 'к': 'k', 'л': 'l', 'м': 'm',
        'н': 'n', 'о': 'o', 'п': 'p', 'р': 'r', 'с': 's', 'т': 't', 'у': 'u',
        'ф': 'f', 'х': 'h', 'ц': 'ts', 'ч': 'ch', 'ш': 'sh', 'щ': 'sch',
        'ъ': '', 'ы': 'y', 'ь': '', 'э': 'e', 'ю': 'yu', 'я': 'ya'
    }
    return ''.join(mapping.get(c, c) for c in text.lower())

def generate_lessons_for_group(
        group_id: uuid.UUID,
        subject_ids: List[uuid.UUID],
        teacher_by_subject: Dict[uuid.UUID, Optional[uuid.UUID]],
        room_ids: List[uuid.UUID],
        start_date: datetime,
        weeks_ahead: int
) -> List[Dict[str, Any]]:
    lessons = []
    current_week = start_date
    end_date = start_date + timedelta(weeks=weeks_ahead)

    while current_week < end_date:
        if current_week.weekday() in WEEKDAYS:
            num_lessons = random.randint(2, 4)
            possible_starts = []
            slot_start = datetime.combine(current_week.date(), DAY_START)
            slot_end = datetime.combine(current_week.date(), DAY_END)
            while slot_start + LESSON_DURATION <= slot_end:
                possible_starts.append(slot_start)
                slot_start += LESSON_DURATION
            if len(possible_starts) < num_lessons:
                num_lessons = len(possible_starts)
            selected_starts = random.sample(possible_starts, num_lessons)
            selected_starts.sort()

            for start_time in selected_starts:
                end_time = start_time + LESSON_DURATION
                subject_id = random.choice(subject_ids)
                teacher_id = teacher_by_subject.get(subject_id)
                if teacher_id and random.random() < 0.7:
                    teacher_id = None
                room_id = random.choice(room_ids) if room_ids else None
                lesson_type = random.choice(['lecture', 'lab', 'seminar'])
                lessons.append({
                    'group_id': group_id,
                    'subject_id': subject_id,
                    'teacher_id': teacher_id,
                    'room_id': room_id,
                    'type': lesson_type,
                    'starts_at': start_time,
                    'ends_at': end_time,
                })
        current_week += timedelta(days=1)
    return lessons

def get_queue_status(opens_at: datetime, closes_at: Optional[datetime], now: datetime) -> str:
    if closes_at and now > closes_at:
        return 'closed'
    if now >= opens_at:
        return 'open'
    return 'draft'

# =====================================================
# Основная функция наполнения
# =====================================================

def seed_database():
    conn = None
    try:
        conn = psycopg.connect(**DB_CONFIG, autocommit=False)
        cur = conn.cursor()

        print("Очистка только сгенерированных скриптом данных (дефолтные пользователи и данные останутся)...")
        # 1. Сначала удаляем слоты очереди, связанные со студентами-ботами или группами
        cur.execute("DELETE FROM queue_slots WHERE student_id IN (SELECT id FROM users WHERE email LIKE '%@student.edu') OR queue_id IN (SELECT id FROM queues WHERE group_id IN (SELECT id FROM groups WHERE name SIMILAR TO '(ИВТ|ПИ|РТ|БИ)-%'));")

        # 2. Удаляем очереди, которые скрипт создал для сгенерированных групп
        cur.execute("DELETE FROM queues WHERE created_by IN (SELECT id FROM users WHERE email LIKE '%@student.edu' OR email LIKE 'admin.%@edu.ru') OR group_id IN (SELECT id FROM groups WHERE name SIMILAR TO '(ИВТ|ПИ|РТ|БИ)-%');")

        # 3. Отвязываем старост, чтобы пользователи не блокировались (FK constrain)
        cur.execute("UPDATE groups SET headman_id = NULL WHERE name SIMILAR TO '(ИВТ|ПИ|РТ|БИ)-%';")

        # 4. Удаляем сгенерированные уроки
        cur.execute("DELETE FROM lessons WHERE group_id IN (SELECT id FROM groups WHERE name SIMILAR TO '(ИВТ|ПИ|РТ|БИ)-%');")

        # 5. Удаляем сгенерированных пользователей (студентов и дополнительных админов)
        cur.execute("DELETE FROM users WHERE email LIKE '%@student.edu' OR email LIKE 'admin.%@edu.ru';")

        # 6. Удаляем сгенерированные группы
        cur.execute("DELETE FROM groups WHERE name SIMILAR TO '(ИВТ|ПИ|РТ|БИ)-%';")

        # 7. Удаляем сгенерированные аудитории
        cur.execute("DELETE FROM rooms WHERE name SIMILAR TO '(A|B|C|ГК)-%';")

        # 8. Удаляем сгенерированные предметы
        subject_names = [
            "Базы данных", "Web-разработка", "Операционные системы", "Алгоритмы и структуры данных",
            "Математический анализ", "Физика", "Английский язык", "Информационная безопасность",
            "Машинное обучение", "Программирование на Python", "Компьютерные сети", "Технологии Java"
        ]
        cur.execute("DELETE FROM subjects WHERE name = ANY(%s);", (subject_names,))

        # 9. Удаляем учителей, у которых нет уроков (вычистим ранее сгенерированных, не трогая тех, у кого остались какие-то пары)
        cur.execute("DELETE FROM teachers WHERE id NOT IN (SELECT teacher_id FROM lessons WHERE teacher_id IS NOT NULL);")

        conn.commit()

        now = datetime.now()

        print("Генерация групп...")
        group_names = [f"ИВТ-{random.randint(21, 24)}", f"ПИ-{random.randint(21, 24)}",
                       f"РТ-{random.randint(21, 24)}", f"БИ-{random.randint(21, 24)}"]
        group_ids = []
        for name in group_names[:GROUPS_COUNT]:
            cur.execute("INSERT INTO groups (name) VALUES (%s) RETURNING id", (name,))
            group_ids.append(cur.fetchone()[0])

        print("Генерация предметов...")
        subject_names = [
            "Базы данных", "Web-разработка", "Операционные системы", "Алгоритмы и структуры данных",
            "Математический анализ", "Физика", "Английский язык", "Информационная безопасность",
            "Машинное обучение", "Программирование на Python", "Компьютерные сети", "Технологии Java"
        ]
        random.shuffle(subject_names)
        subject_ids = []
        for name in subject_names[:SUBJECTS_COUNT]:
            cur.execute("INSERT INTO subjects (name) VALUES (%s) RETURNING id", (name,))
            subject_ids.append(cur.fetchone()[0])

        print("Генерация преподавателей...")
        teacher_ids = []
        for _ in range(TEACHERS_COUNT):
            first_name = fake.first_name()
            last_name = fake.last_name()
            patronymic = fake.middle_name()
            cur.execute(
                "INSERT INTO teachers (last_name, first_name, patronymic) VALUES (%s, %s, %s) RETURNING id",
                (last_name, first_name, patronymic)
            )
            teacher_ids.append(cur.fetchone()[0])

        teacher_by_subject = {}
        for subj_id in subject_ids:
            if random.random() < 0.8:
                teacher_by_subject[subj_id] = random.choice(teacher_ids)
            else:
                teacher_by_subject[subj_id] = None

        print("Генерация аудиторий...")
        room_ids = []
        buildings = ["A", "B", "C", "ГК"]
        for i in range(ROOMS_COUNT):
            bld = random.choice(buildings)
            num = random.randint(101, 500)
            name = f"{bld}-{num}"
            try:
                cur.execute("INSERT INTO rooms (name) VALUES (%s) RETURNING id", (name,))
                room_ids.append(cur.fetchone()[0])
            except psycopg.errors.UniqueViolation:
                conn.rollback()  # откатываем только эту вставку
                name = f"{bld}-{num}_{i}"
                cur.execute("INSERT INTO rooms (name) VALUES (%s) RETURNING id", (name,))
                room_ids.append(cur.fetchone()[0])

        print("Генерация пользователей (студенты и старосты)...")
        user_ids_by_group = {}
        headman_ids = {}
        hashed_pw = hash_password(DEFAULT_PASSWORD)

        for group_id in group_ids:
            users_in_group = []
            count = random.randint(STUDENTS_PER_GROUP_MIN, STUDENTS_PER_GROUP_MAX)
            for _ in range(count):
                first_name = fake.first_name()
                last_name = fake.last_name()
                patronymic = fake.middle_name()
                email = f"{transliterate(last_name)}.{transliterate(first_name)}.{random.randint(1,999)}@student.edu"
                cur.execute("""
                            INSERT INTO users (email, password_hash, first_name, last_name, patronymic, role, group_id)
                            VALUES (%s, %s, %s, %s, %s, 'student', %s) RETURNING id
                            """, (email, hashed_pw, first_name, last_name, patronymic, group_id))
                users_in_group.append(cur.fetchone()[0])
            user_ids_by_group[group_id] = users_in_group
            headman_id = random.choice(users_in_group)
            headman_ids[group_id] = headman_id

        for group_id, headman_id in headman_ids.items():
            cur.execute("UPDATE groups SET headman_id = %s WHERE id = %s", (headman_id, group_id))

        admin_ids = []
        for _ in range(random.randint(2, 3)):
            first_name = fake.first_name()
            last_name = fake.last_name()
            patronymic = fake.middle_name()
            email = f"admin.{transliterate(last_name)}@edu.ru"
            cur.execute("""
                        INSERT INTO users (email, password_hash, first_name, last_name, patronymic, role, group_id)
                        VALUES (%s, %s, %s, %s, %s, 'admin', NULL) RETURNING id
                        """, (email, hashed_pw, first_name, last_name, patronymic))
            admin_ids.append(cur.fetchone()[0])

        print("Генерация расписания уроков...")
        all_lessons = []
        start_date = now.replace(hour=0, minute=0, second=0, microsecond=0)
        if start_date.weekday() >= 5:
            start_date += timedelta(days=(7 - start_date.weekday()))

        for group_id in group_ids:
            lessons = generate_lessons_for_group(
                group_id, subject_ids, teacher_by_subject, room_ids,
                start_date, WEEKS_AHEAD
            )
            all_lessons.extend(lessons)

        # Вставка уроков пакетно через executemany
        lesson_insert_sql = """
                            INSERT INTO lessons (group_id, subject_id, teacher_id, room_id, type, starts_at, ends_at)
                            VALUES (%(group_id)s, %(subject_id)s, %(teacher_id)s, %(room_id)s, %(type)s, %(starts_at)s, %(ends_at)s) \
                            """
        with conn.cursor() as cur2:
            cur2.executemany(lesson_insert_sql, all_lessons)

        cur.execute("SELECT id, group_id, starts_at FROM lessons ORDER BY starts_at;")
        lessons_data = cur.fetchall()  # список кортежей

        print("Генерация очередей и слотов...")
        for lesson_id, group_id, starts_at in lessons_data:
            opens_at = starts_at - timedelta(days=random.randint(1, 2))
            opens_at = opens_at.replace(hour=random.randint(8, 12), minute=0, second=0)
            closes_at = starts_at - timedelta(minutes=random.choice([0, 15, 30]))
            max_size = random.randint(10, len(user_ids_by_group.get(group_id, [])))
            created_by = random.choice([headman_ids[group_id], random.choice(admin_ids)])
            status = get_queue_status(opens_at, closes_at, now)

            cur.execute("""
                        INSERT INTO queues (group_id, subject_id, lesson_id, created_by, created_at, opens_at, closes_at, max_size, status)
                        VALUES (%s, (SELECT subject_id FROM lessons WHERE id=%s), %s, %s, %s, %s, %s, %s, %s)
                        RETURNING id
                        """, (group_id, lesson_id, lesson_id, created_by, now, opens_at, closes_at, max_size, status))
            queue_id = cur.fetchone()[0]

            if status != 'draft':
                students = user_ids_by_group.get(group_id, [])
                num_slots = random.randint(0, min(max_size, len(students)))
                selected_students = random.sample(students, num_slots)
                slot_records = []
                for student_id in selected_students:
                    if closes_at and now > closes_at:
                        signup_limit = closes_at
                    else:
                        signup_limit = now
                    signed_up_at = fake.date_time_between(start_date=opens_at, end_date=signup_limit)
                    if status == 'closed':
                        if now > starts_at + LESSON_DURATION:
                            r = random.random()
                            if r < PASSED_PROB:
                                slot_status = 'passed'
                            elif r < PASSED_PROB + FAILED_PROB:
                                slot_status = 'failed'
                            else:
                                slot_status = 'no_show'
                        else:
                            slot_status = 'waiting'
                    else:
                        slot_status = 'waiting'
                    slot_records.append((queue_id, student_id, slot_status, signed_up_at))

                if slot_records:
                    cur.executemany("""
                                    INSERT INTO queue_slots (queue_id, student_id, status, signed_up_at)
                                    VALUES (%s, %s, %s, %s)
                                    ON CONFLICT (queue_id, student_id) DO NOTHING
                                    """, slot_records)

        print("Генерация дополнительных очередей (без привязки к уроку)...")
        for _ in range(10):
            group_id = random.choice(group_ids)
            subject_id = random.choice(subject_ids)
            created_by = random.choice([headman_ids[group_id], random.choice(admin_ids)])
            opens_at = now + timedelta(days=random.randint(-3, 7))
            opens_at = opens_at.replace(hour=random.randint(8, 18), minute=0)
            closes_at = opens_at + timedelta(days=random.randint(1, 3))
            max_size = random.randint(10, 40)
            status = get_queue_status(opens_at, closes_at, now)
            cur.execute("""
                        INSERT INTO queues (group_id, subject_id, created_by, created_at, opens_at, closes_at, max_size, status)
                        VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                        RETURNING id
                        """, (group_id, subject_id, created_by, now, opens_at, closes_at, max_size, status))
            queue_id = cur.fetchone()[0]

            if status != 'draft':
                students = user_ids_by_group.get(group_id, [])
                num_slots = random.randint(0, min(max_size, len(students)))
                if num_slots > 0:
                    selected_students = random.sample(students, num_slots)
                    slot_records = []
                    for student_id in selected_students:
                        if closes_at and now > closes_at:
                            signup_limit = closes_at
                        else:
                            signup_limit = now
                        signed_up_at = fake.date_time_between(start_date=opens_at, end_date=signup_limit)
                        slot_records.append((queue_id, student_id, signed_up_at))
                    if slot_records:
                        cur.executemany("""
                                        INSERT INTO queue_slots (queue_id, student_id, signed_up_at)
                                        VALUES (%s, %s, %s)
                                        ON CONFLICT DO NOTHING
                                        """, slot_records)

        conn.commit()
        total_users = len(admin_ids) + sum(len(u) for u in user_ids_by_group.values())
        print(f"Наполнение завершено. Создано групп: {len(group_ids)}, пользователей: {total_users}, уроков: {len(all_lessons)}")

    except Exception as e:
        if conn:
            conn.rollback()
        print(f"Ошибка при наполнении БД: {e}")
        raise
    finally:
        if conn:
            conn.close()

if __name__ == "__main__":
    seed_database()