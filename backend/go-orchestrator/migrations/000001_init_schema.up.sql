CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone_number VARCHAR(20) UNIQUE,
    pass_hash VARCHAR(255),
    phone_verified BOOLEAN NOT NULL DEFAULT false,
    verification_code VARCHAR(10),
    code_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    qr_issued_at TIMESTAMP,
    qr_expires_at TIMESTAMP
);
CREATE TABLE IF NOT EXISTS user_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS goods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    weight DECIMAL(10, 2) NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    quantity_available INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS drones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'idle',
    battery_level DECIMAL(5, 2) DEFAULT 100.0,
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    altitude DECIMAL(10, 2),
    speed DECIMAL(10, 2),
    current_delivery_id UUID,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS parcel_automats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    city VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    number_of_cells INTEGER NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    coordinates VARCHAR(255) NOT NULL,
    aruco_id INTEGER NOT NULL,
    is_working BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE IF NOT EXISTS locker_cells_out (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    cell_number INTEGER
);
CREATE TABLE IF NOT EXISTS locker_cells_internal (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    cell_number INTEGER
);
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    good_id UUID NOT NULL,
    parcel_automat_id UUID NOT NULL,
    locker_cell_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    drone_id UUID,
    parcel_automat_id UUID NOT NULL,
    internal_locker_cell_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
ALTER TABLE user_devices
ADD CONSTRAINT fk_user_devices_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE locker_cells_out
ADD CONSTRAINT fk_locker_cells_out_post_id FOREIGN KEY (post_id) REFERENCES parcel_automats(id) ON DELETE CASCADE;
ALTER TABLE locker_cells_internal
ADD CONSTRAINT fk_locker_cells_internal_post_id FOREIGN KEY (post_id) REFERENCES parcel_automats(id) ON DELETE CASCADE;
ALTER TABLE orders
ADD CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE orders
ADD CONSTRAINT fk_orders_good_id FOREIGN KEY (good_id) REFERENCES goods(id) ON DELETE CASCADE;
ALTER TABLE orders
ADD CONSTRAINT fk_orders_parcel_automat_id FOREIGN KEY (parcel_automat_id) REFERENCES parcel_automats(id) ON DELETE CASCADE;
ALTER TABLE orders
ADD CONSTRAINT fk_orders_locker_cell_id FOREIGN KEY (locker_cell_id) REFERENCES locker_cells_out(id) ON DELETE
SET NULL;
ALTER TABLE deliveries
ADD CONSTRAINT fk_deliveries_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;
ALTER TABLE deliveries
ADD CONSTRAINT fk_deliveries_drone_id FOREIGN KEY (drone_id) REFERENCES drones(id) ON DELETE
SET NULL;
ALTER TABLE deliveries
ADD CONSTRAINT fk_deliveries_parcel_automat_id FOREIGN KEY (parcel_automat_id) REFERENCES parcel_automats(id) ON DELETE CASCADE;
ALTER TABLE deliveries
ADD CONSTRAINT fk_deliveries_internal_locker_cell_id FOREIGN KEY (internal_locker_cell_id) REFERENCES locker_cells_internal(id);
ALTER TABLE drones
ADD CONSTRAINT fk_drones_current_delivery_id FOREIGN KEY (current_delivery_id) REFERENCES deliveries(id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries(status);
CREATE INDEX IF NOT EXISTS idx_deliveries_drone_id ON deliveries(drone_id);
CREATE INDEX IF NOT EXISTS idx_locker_cells_out_status ON locker_cells_out(status);
CREATE INDEX IF NOT EXISTS idx_locker_cells_internal_status ON locker_cells_internal(status);
CREATE INDEX IF NOT EXISTS idx_locker_cells_out_post_number ON locker_cells_out(post_id, cell_number);
CREATE INDEX IF NOT EXISTS idx_locker_cells_internal_post_number ON locker_cells_internal(post_id, cell_number);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_parcel_automats_ip_address ON parcel_automats(ip_address);
CREATE INDEX IF NOT EXISTS idx_drones_ip_address ON drones(ip_address);
CREATE INDEX IF NOT EXISTS idx_drones_status ON drones(status);
CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
CREATE OR REPLACE FUNCTION update_drone_battery(
        p_drone_id UUID,
        p_battery_level DECIMAL(5, 2)
    ) RETURNS void AS $$ BEGIN
UPDATE drones
SET battery_level = p_battery_level,
    updated_at = CURRENT_TIMESTAMP
WHERE id = p_drone_id;
END;
$$ LANGUAGE plpgsql;