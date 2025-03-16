CREATE TYPE INVOICE_SERVICE_NAME as ENUM('DMP', 'SSP');

CREATE TYPE INVOICE_STATUS AS ENUM('PAID', 'PENDING', 'UNPAID');

CREATE TABLE invoices (
    id UUID PRIMARY KEY UNIQUE NOT NULL,
    service_name INVOICE_SERVICE_NAME NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    status INVOICE_STATUS NOT NULL,
    date TIMESTAMP NOT NULL
);

INSERT INTO invoices (id, service_name, amount, status, date) VALUES
    ('dda97bce-ac2a-4431-9c7b-3b6bcdfe8a23', 'DMP', 120.30, 'PAID', '2025-03-18 12:34:56'),
    ('dc874c3f-2773-413e-a3c8-e9f24b04079c', 'SSP', 230.50, 'PENDING', '2024-03-18 12:34:56'),
    ('550e8400-e29b-41d4-a716-446655440000', 'DMP', 150.75, 'UNPAID', '2024-04-01 09:00:00'),
    ('6ba7b810-9dad-11d1-80b4-00c04fd430c8', 'SSP', 300.25, 'PAID', '2024-03-15 15:30:00'),
    ('6ba7b811-9dad-11d1-80b4-00c04fd430c8', 'DMP', 175.90, 'PENDING', '2024-03-20 11:45:00'),
    ('6ba7b812-9dad-11d1-80b4-00c04fd430c8', 'SSP', 450.00, 'PAID', '2024-03-25 14:20:00'),
    ('6ba7b813-9dad-11d1-80b4-00c04fd430c8', 'DMP', 200.80, 'UNPAID', '2024-04-05 10:15:00'),
    ('6ba7b814-9dad-11d1-80b4-00c04fd430c8', 'SSP', 275.60, 'PENDING', '2024-03-28 16:40:00'),
    ('6ba7b815-9dad-11d1-80b4-00c04fd430c8', 'DMP', 180.45, 'PAID', '2024-03-22 13:50:00'),
    ('6ba7b816-9dad-11d1-80b4-00c04fd430c8', 'SSP', 325.90, 'UNPAID', '2024-04-02 08:30:00');
