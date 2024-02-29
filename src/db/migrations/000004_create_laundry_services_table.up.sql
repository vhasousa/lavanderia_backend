CREATE TABLE IF NOT EXISTS laundry_services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status VARCHAR(15) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    estimated_completion_date TIMESTAMP,
    total_price numeric(10,2),
    weight numeric(10,2),
    is_piece boolean,
    is_weight boolean,
    client_id UUID,
    is_paid boolean,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);