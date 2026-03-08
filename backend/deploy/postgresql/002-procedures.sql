-- =====================================================
-- Передача роли старосты
-- =====================================================
CREATE OR REPLACE PROCEDURE transfer_headman(
    p_current_headman_id UUID,
    p_to_user_id UUID
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_role user_role;
    v_group_id UUID;
    v_blocked BOOLEAN;
    v_current_group UUID;
BEGIN
    -- Получить данные нового кандидата
    SELECT role, group_id, is_blocked
    INTO v_role, v_group_id, v_blocked
    FROM users
    WHERE id = p_to_user_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'User % not found', p_to_user_id;
    END IF;

    -- Получить группу текущего старосты
    SELECT group_id
    INTO v_current_group
    FROM users
    WHERE id = p_current_headman_id
      AND role = 'headman';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Current headman % not found or not headman', p_current_headman_id;
    END IF;

    -- Проверки
    IF v_role <> 'student' THEN
        RAISE EXCEPTION 'Target user must have role student';
    END IF;

    IF v_group_id <> v_current_group THEN
        RAISE EXCEPTION 'Target user must belong to the same group';
    END IF;

    IF v_blocked THEN
        RAISE EXCEPTION 'Target user is blocked';
    END IF;

    -- Снять роль старосты
    UPDATE users
    SET role = 'student'
    WHERE id = p_current_headman_id;

    -- Назначить нового
    UPDATE users
    SET role = 'headman'
    WHERE id = p_to_user_id;

END;
$$;

-- =====================================================
-- Перенос неуспевших студентов
-- =====================================================
CREATE OR REPLACE PROCEDURE transfer_failed_students(
    p_source_queue_id UUID,
    p_target_queue_id UUID
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_source_status queue_status;
    v_target_status queue_status;
    v_group_id UUID;
    v_subject_id UUID;
BEGIN
    -- Проверить целевую очередь
    SELECT status, group_id, subject_id
    INTO v_target_status, v_group_id, v_subject_id
    FROM queues
    WHERE id = p_target_queue_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Target queue % not found', p_target_queue_id;
    END IF;

    IF v_target_status <> 'draft' THEN
        RAISE EXCEPTION 'Target queue must be in draft status';
    END IF;

    -- Проверить исходную очередь
    SELECT status
    INTO v_source_status
    FROM queues
    WHERE id = p_source_queue_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Source queue % not found', p_source_queue_id;
    END IF;

    IF v_source_status NOT IN ('closed', 'archived') THEN
        RAISE EXCEPTION 'Source queue must be closed or archived';
    END IF;

    -- Перенос студентов
    INSERT INTO queue_slots (id, queue_id, student_id, status, signed_up_at)
    SELECT
        gen_random_uuid(),
        p_target_queue_id,
        s.student_id,
        'waiting',
        now()
    FROM queue_slots s
    WHERE s.queue_id = p_source_queue_id
      AND s.status IN ('failed', 'no_show')
      AND NOT EXISTS (
          SELECT 1
          FROM queue_slots t
          WHERE t.queue_id = p_target_queue_id
            AND t.student_id = s.student_id
      );

END;
$$;