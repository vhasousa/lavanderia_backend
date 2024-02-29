CREATE TABLE IF NOT EXISTS laundry_items_services (
    laundry_service_id UUID,
    laundry_item_id UUID,
    item_quantity INT,
    observation TEXT,
    PRIMARY KEY (laundry_service_id, laundry_item_id),
    FOREIGN KEY (laundry_service_id) REFERENCES laundry_services(id) ON DELETE CASCADE,
    FOREIGN KEY (laundry_item_id) REFERENCES laundry_items(id)
);
