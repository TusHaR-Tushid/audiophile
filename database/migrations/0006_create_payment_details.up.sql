CREATE TABLE IF NOT EXISTS bill_details(
                                      id uuid primary key default gen_random_uuid() not null ,
                                      user_id uuid REFERENCES users(id),
                                      payment_id uuid REFERENCES payment(id),
                                      order_details uuid REFERENCES order_details(id)
);