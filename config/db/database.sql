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
