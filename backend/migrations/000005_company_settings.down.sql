DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'setting:manage') OR (role_id IN (SELECT id FROM roles WHERE code = 'director') AND permission_id IN (SELECT id FROM permissions WHERE code IN ('schedule:manage', 'leave:manage')));
DELETE FROM permissions WHERE code = 'setting:manage';

ALTER TABLE companies 
    DROP CONSTRAINT IF EXISTS chk_company_logo_type,
    DROP CONSTRAINT iF EXISTS chk_company_logo_size,
    DROP COLUMN IF EXISTS logo_updated_at,
    DROP COLUMN IF EXISTS logo_type,
    DROP COLUMN IF EXISTS logo,
    DROP COLUMN IF EXISTS payslip_note,
    DROP COLUMN IF EXISTS email_signature,
    DROP COLUMN IF EXISTS email_sender,
    DROP COLUMN IF EXISTS app_name;

    