CREATE TABLE IF NOT EXISTS users(
    id uuid primary key default gen_random_uuid() not null ,
    name TEXT NOT NULL ,
    email TEXT UNIQUE CHECK (email <> '') NOT NULL ,
    password TEXT NOT NULL ,
    phone_no TEXT NOT NULL ,
    age INTEGER  NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS user_address(
    id uuid primary key default gen_random_uuid() not null ,
    user_id uuid REFERENCES users(id),
    address TEXT NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS category(
    id uuid primary key default gen_random_uuid() not null ,
    name TEXT NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS inventory(
    id uuid primary key default gen_random_uuid() not null,
    category_id uuid REFERENCES category(id),
    name TEXT NOT NULL,
    price FLOAT NOT NULL ,
    Quantity INTEGER NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS images(
    id uuid primary key default gen_random_uuid() not null ,
    url TEXT ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS images_per_product(
    id uuid primary key default gen_random_uuid() not null ,
    product_id uuid REFERENCES inventory(id),
    image_id   uuid REFERENCES images(id)
);

CREATE TABLE IF NOT EXISTS user_cart_products(
    id uuid primary key default gen_random_uuid() not null ,
    product_id uuid REFERENCES inventory(id),
    quantity INTEGER NOT NULL ,
    total_amount FLOAT NOT NULL ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

create type payment_type as enum('cod', 'debit_card', 'credit_card');

CREATE TABLE IF NOT EXISTS payment(
    id uuid primary key default gen_random_uuid() not null ,
    user_id uuid REFERENCES users(id),
    payment_type payment_type NOT NULL ,
    name TEXT NOT NULL ,
    account_number INTEGER NOT NULL
);

create type status_type as enum('created', 'processing', 'completed');

CREATE TABLE IF NOT EXISTS orders(
    id uuid primary key default gen_random_uuid() not null ,
    user_id uuid REFERENCES users(id),
    payment_id uuid REFERENCES payment(id),
    address_id uuid REFERENCES user_address(id),
    total_amount FLOAT NOT NULL,
    status status_type NOT NULL DEFAULT 'created',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS sessions(
        id uuid primary key default gen_random_uuid() not null ,
        user_id uuid REFERENCES users(id) NOT NULL ,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
        expires_at TIMESTAMP WITH TIME ZONE
);