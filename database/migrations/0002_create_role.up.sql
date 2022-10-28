create type user_type as enum('admin', 'user');

CREATE TABLE IF NOT EXISTS roles(
                                    id uuid primary key default gen_random_uuid() not null ,
                                    roles user_type ,
                                    user_id uuid REFERENCES users(id)
);