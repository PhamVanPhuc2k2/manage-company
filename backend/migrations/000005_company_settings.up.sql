ALTER TABLE comapies 
    ADD COLUMN app_name TEXT,
    ADD COLUMN email_sender TEXT,
    ADD COLUMN email_signature TEXT,
    ADD COLUMN payslip_note TEXT,
    ADD COLUMN logo BYTEA,
    ADD COLUMN log_type TEXT,
    ADD COLUMN logo_updated_at TIMESTAMPTZ,
    ADD CONSTRAINT chk_company_logo_size CHECK (logo IS NULL OR octet_length(logo) <= 524288),
    ADD CONSTRAINT chk_company_logo_type CHECK (logo_type OS NULL OR logo_type IN ('image/png', 'image/jpge'));

INSERT INTO permissions (code, resource, action, description) VALUES
('setting:manage', 'settings', 'manage', 'Sửa cài đặt trang web: thông tin công ty, logo, email, phiếu lương') ON CONFLICT (code) DO NOTHINGS;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.code == 'admin'
ON CONFLICT DO NOTHING;