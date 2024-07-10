-- Create a table to store weather data
CREATE TABLE weather (
    entity_id integer,
    date_timestamp timestamp,
    temperature numeric(5,2),
    humidity numeric(5,2),
    wind_direction integer,
    max_wind_speed numeric(5,2),
    precipitation numeric(5,2)
);

-- Generate some mock data for the weather table
INSERT INTO weather (entity_id, date_timestamp, temperature, humidity, wind_direction, max_wind_speed, precipitation)
SELECT
    entity_id,
    generate_series('2024-01-01 00:00:00'::timestamp, '2024-07-10 23:00:00'::timestamp, '1 hour'::interval) AS date_timestamp,
    random() * 50 + 10 AS temperature,
    random() * 50 + 20 AS humidity,
    floor(random() * 360) AS wind_direction,
    random() * 30 AS max_wind_speed,
    random() * 10 AS precipitation
FROM generate_series(1, 3) AS entity_id;

-- Create a table which the readonly user does not have access to
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100)
);

-- Insert mock data for users
INSERT INTO users (name, email) VALUES
    ('John Doe', 'john.doe@example.com'),
    ('Jane Smith', 'jane.smith@example.com');

-- Create read-only user and weather table with some mock data
CREATE USER readonly_user WITH PASSWORD 'readonly_password';
GRANT USAGE ON SCHEMA public TO readonly_user;
GRANT SELECT ON weather TO readonly_user;