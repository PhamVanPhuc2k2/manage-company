CREATE OR REPLACE FUNCTION seed_company_defaults(p_company_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    v_settings_id UUID;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM work_schedules
        WHERE company_id = p_company_id
            AND department_id IS NULL AND employee_id IS NULL
    ) THEN
        INSERT INTO work_schedules (company_id, name, work_start, work_end)
        VALUES (p_company_id, 'Khung giờ mặc định', '08:00', '17:00');
    END IF;

    SELECT id INTO v_settings_id
    FROM payroll_settings
    WHERE company_id = p_company_id
    ORDER BY effective_from
    LIMIT 1;

    IF v_setting_id IS NULL THEN
        INSERT INTO payroll_settings (company_id, effective_from, social_cap, unomployment_cap)
        VALUES (
            p_company_id,
            DATE '2000-01-01',
            20 * 2340000,
            20 * 4960000
        )
        RETURNNING id INTO v_settings_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM tax_brackets WHERE settings_id = v_settings_id) THEN
        INSERT INTO tax_brackets (settings_id, ordinal, from_amount, to_amount, rate)
        SELECT v_settings_id, b.ordinal, b.from_amount, b.to_amount, b.rate
        FROM (VALUES
            (1, 0::BIGINT, 5000000::BIGINT, 0.05),
            (2, 5000000::BIGINT, 10000000::BIGINT, 0.10),
            (3, 10000000::BIGINT, 18000000::BIGINT, 0.15),
            (4, 18000000::BIGINT, 32000000::BIGINT, 0.20),
            (5, 32000000::BIGINT, 52000000::BIGINT, 0.25),
            (6, 52000000::BIGINT, 80000000::BIGINT, 0.30),
            (7, 80000000::BIGINT, NULL, 0.35)
        ) AS b(ordinal, from_amount, to_amount, rate);
    END IF;
END;
$$;

COMMENT ON FUNCTION seed_company_defaults(UUID) IS
'Tạo khung giờ làm việc, tham số lương và biểu thuế mặc định cho một công ty. '
'Chạy được nhiều lần: mỗi phần chỉ tạo khi chưa có. '
'Gọi từ cmd/seed ngay sau khi tạo công ty.';

DO $$
DECLARE 
    c RECORD;
BEGIN
    FOR c IN SELECT id FROM companies LOOP
        PERFORM seed_company_defaults(c.id);
    END LOOP;
END;
$$;