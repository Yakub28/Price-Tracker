-- Create stocks table
CREATE TABLE IF NOT EXISTS stocks (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create stock_prices table (time-series data)
CREATE TABLE IF NOT EXISTS stock_prices (
    id BIGSERIAL PRIMARY KEY,
    stock_id INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    price DECIMAL(12, 4) NOT NULL,
    change_percent DECIMAL(8, 4),
    volume BIGINT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_stock FOREIGN KEY (stock_id) REFERENCES stocks(id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_stock_prices_stock_id ON stock_prices(stock_id);
CREATE INDEX IF NOT EXISTS idx_stock_prices_timestamp ON stock_prices(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_stock_prices_stock_timestamp ON stock_prices(stock_id, timestamp DESC);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    stock_id INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    threshold DECIMAL(8, 4) NOT NULL,
    message TEXT,
    triggered_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_alert_stock FOREIGN KEY (stock_id) REFERENCES stocks(id)
);

CREATE INDEX IF NOT EXISTS idx_alerts_stock_id ON alerts(stock_id);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts(triggered_at DESC);
