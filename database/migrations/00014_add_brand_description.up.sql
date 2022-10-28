CREATE TABLE IF NOT EXISTS brands(
                                    id SERIAL PRIMARY KEY not null ,
                                    brand_name TEXT ,
                                    brand_description TEXT
);

ALTER TABLE inventory ADD COLUMN brand_id INTEGER REFERENCES brands(id);
ALTER TABLE inventory ADD COLUMN product_description TEXT;