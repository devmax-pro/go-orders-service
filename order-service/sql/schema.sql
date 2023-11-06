CREATE TABLE IF NOT EXISTS order_deliveries
(
    id        SERIAL PRIMARY KEY,
    delivery_name      VARCHAR(255), -- "name"
    phone     VARCHAR(12),
    zip       VARCHAR(10),
    city      VARCHAR(255),
    address   VARCHAR(255),
    region    VARCHAR(255),
    email     VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS order_payments
(
    id            SERIAL PRIMARY KEY,
    payment_transaction   VARCHAR(32), -- "transaction"
    request_id    VARCHAR(32),
    currency      CHAR(3),
    payment_provider VARCHAR(255), -- "provider"
    amount        INTEGER,
    payment_dt    BIGINT,
    bank          VARCHAR(255),
    delivery_cost INTEGER,
    goods_total   INTEGER,
    custom_fee    INTEGER
);

CREATE TABLE IF NOT EXISTS orders
(
    order_uid          VARCHAR(32) PRIMARY KEY,
    track_number       VARCHAR(32),
    entry              VARCHAR(32),
    delivery_id        SERIAL,
    payment_id         SERIAL,
    items_id           SERIAL,
    order_locale       CHAR(2), -- "locale"
    internal_signature VARCHAR(255),
    customer_id        CHAR(32),
    delivery_service   VARCHAR(255),
    shard_key           VARCHAR(32),
    sm_id              INTEGER,
    date_created       TIMESTAMP,
    oof_shard          VARCHAR(32),
    FOREIGN KEY (delivery_id) REFERENCES order_deliveries (id) ON DELETE RESTRICT,
    FOREIGN KEY (payment_id) REFERENCES order_payments (id)  ON DELETE RESTRICT
);


CREATE TABLE IF NOT EXISTS order_items
(
    id           SERIAL PRIMARY KEY,
    order_uid    VARCHAR(32),
    chrt_id      BIGINT,
    track_number VARCHAR(32),
    price        INTEGER,
    rid          VARCHAR(32),
    item_name    VARCHAR(255), -- "name"
    sale         INTEGER,
    size         VARCHAR(255),
    total_price  INTEGER,
    nm_id        INTEGER,
    brand        VARCHAR(255),
    status       INTEGER,
    FOREIGN KEY (order_uid) REFERENCES orders (order_uid) ON DELETE CASCADE
)