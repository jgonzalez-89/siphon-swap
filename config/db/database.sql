CREATE SCHEMA IF NOT EXISTS cryptoswap;

USE cryptoswap;

CREATE TABLE currency (
    symbol VARCHAR(16) NOT NULL,
    name VARCHAR(100) NOT NULL,
    image VARCHAR(255) NOT NULL,
    available BOOLEAN NOT NULL,
    address_validation VARCHAR(255),
    price DECIMAL(38, 18) NOT NULL,
    popular BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (symbol)
);

CREATE TABLE currencies_networks (
    symbol VARCHAR(16) NOT NULL,
    network VARCHAR(100) NOT NULL,
    PRIMARY KEY (symbol, network),
    FOREIGN KEY (symbol) REFERENCES currency(symbol)
);


CREATE TABLE swap (
    id VARCHAR(50) NOT NULL,
    from_symbol VARCHAR(16) NOT NULL,
    from_network VARCHAR(100) NOT NULL,
    to_symbol VARCHAR(16) NOT NULL,
    to_network VARCHAR(100) NOT NULL,
    exchange_id VARCHAR(255) NOT NULL,
    payin_amount DECIMAL(38, 18) NOT NULL,
    payout_amount DECIMAL(38, 18) NOT NULL,
    payout_address VARCHAR(255) NOT NULL,
    reason TEXT,
    to_address VARCHAR(255) NOT NULL,
    refund_address VARCHAR(255) NOT NULL,
    exchange VARCHAR(100) NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);
