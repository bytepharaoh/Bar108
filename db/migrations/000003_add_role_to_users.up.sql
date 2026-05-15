ALTER TABLE users
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'customer'
CHECK (role IN ('customer', 'admin'));