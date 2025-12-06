CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP
);

---- create above / drop below ----

DROP TABLE subscriptions;
