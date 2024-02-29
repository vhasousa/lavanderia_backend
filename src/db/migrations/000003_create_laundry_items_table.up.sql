CREATE TABLE IF NOT EXISTS laundry_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    price numeric(10,2)
);