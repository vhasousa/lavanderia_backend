CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone CHAR(12),
    is_mensal boolean,
    monthly_date DATE
) INHERITS (users);