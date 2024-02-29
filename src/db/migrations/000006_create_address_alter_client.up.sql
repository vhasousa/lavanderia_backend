CREATE TABLE address (
    address_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state CHAR(2) NOT NULL,
    postal_code VARCHAR(20),
    number VARCHAR(5) NOT NULL,
    complement VARCHAR(255),
    landmark VARCHAR(255)
);

ALTER TABLE clients
ADD COLUMN address_id UUID;

ALTER TABLE clients
ADD CONSTRAINT fk_address
FOREIGN KEY (address_id) REFERENCES address(address_id)
ON DELETE CASCADE;
