-- 预置演示数据：100 张证书 + 近三届年检 + 续期/费用记录（表须已由 GORM 迁移创建）

WITH RECURSIVE seq(n) AS (
  SELECT 1 UNION ALL SELECT n+1 FROM seq WHERE n < 100
)
INSERT INTO certificates (name, category, issuing_org, cert_no, issue_date, expiry_date, department, responsible_person, is_cancelled, scan_url, front_image_url, back_image_url, notes, created_at, updated_at)
SELECT
  '演示证书-' || printf('%03d', n),
  CASE (n % 7)
    WHEN 0 THEN '营业执照'
    WHEN 1 THEN '行业许可证'
    WHEN 2 THEN '认证证书'
    WHEN 3 THEN '安全生产许可'
    WHEN 4 THEN '特种经营许可'
    ELSE '资质等级证书'
  END,
  '颁证机构-' || printf('%02d', ((n - 1) % 9) + 1),
  'DEMO-' || printf('%06d', 202600 + n),
  date('now', '-' || (380 + (n * 3)) || ' days'),
  date('now', '+' || (((n * 19) % 420) - 90) || ' days'),
  '责任部门-' || printf('%02d', ((n - 1) % 8) + 1),
  '负责人-' || printf('%02d', ((n - 1) % 15) + 1),
  CASE WHEN n = 13 OR n = 27 THEN 1 ELSE 0 END,
  'https://example.invalid/scan/' || n || '.pdf',
  'https://example.invalid/front/' || n || '.png',
  'https://example.invalid/back/' || n || '.png',
  '演示台账数据',
  datetime('now'),
  datetime('now')
FROM seq;

-- 近 3 年年检记录（前 60 本证）
INSERT INTO annual_inspection_records (certificate_id, year, record_url, notes, created_at)
SELECT c.id,
       y.yr,
       'https://example.invalid/inspection/' || c.id || '/' || y.yr || '.pdf',
       printf('年检合格 %d', y.yr),
       datetime('now')
FROM certificates c
CROSS JOIN (
  SELECT 2023 AS yr UNION ALL SELECT 2024 UNION ALL SELECT 2025
) AS y
WHERE c.id <= 60;

-- 历史续期申请（偶数证，模拟三年前至一年前流程）
INSERT INTO renewal_applications (certificate_id, applicant, apply_time, materials_checklist, expected_submit_date, progress, new_cert_notes, created_at, updated_at)
SELECT
  c.id,
  c.responsible_person,
  datetime('now', '-' || (500 + (c.id * 7) % 200) || ' days'),
  '["营业执照副本","法人身份证","申请表","承诺书"]',
  datetime('now', '-' || (480 + (c.id * 7) % 200) || ' days'),
  '已通过',
  '归档：续期完成（演示）',
  datetime('now', '-' || (500 + (c.id * 7) % 200) || ' days'),
  datetime('now')
FROM certificates c
WHERE c.id <= 55 AND c.id % 2 = 0;

-- 费用记录：分布在近三年，便于年度汇总核对
INSERT INTO fee_records (certificate_id, renewal_application_id, admin_fee_cents, agency_fee_cents, fee_date, description, created_at)
SELECT c.id, NULL,
       45000 + (c.id * 173 % 12000),
       15000 + (c.id * 97 % 8000),
       date(printf('%d-06-15', 2024 + (c.id % 3) - 1)),
       '年检/续期行政与代理费用',
       datetime('now')
FROM certificates c
WHERE c.id <= 85;

INSERT INTO fee_records (certificate_id, renewal_application_id, admin_fee_cents, agency_fee_cents, fee_date, description, created_at)
SELECT c.id, NULL,
       30000 + (c.id * 211 % 9000),
       12000 + (c.id * 131 % 6000),
       date(printf('%d-03-10', 2023 + (c.id % 3))),
       '补充费用条目',
       datetime('now')
FROM certificates c
WHERE c.id BETWEEN 20 AND 70 AND c.id % 3 = 0;
