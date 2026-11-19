-- Money is whole cents. Stock is single units. Times are unix milliseconds.
-- `day` is the business day (06:00 to 06:00), stored so reports never recompute it.

CREATE TABLE settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE categories (
  id     INTEGER PRIMARY KEY,
  name   TEXT NOT NULL,
  color  TEXT NOT NULL,
  sort   INTEGER NOT NULL DEFAULT 0,
  active INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE products (
  id            INTEGER PRIMARY KEY,
  category_id   INTEGER NOT NULL REFERENCES categories(id),
  name          TEXT NOT NULL,
  price_cents   INTEGER NOT NULL,
  cost_cents    INTEGER NOT NULL DEFAULT 0,
  barcode       TEXT,
  crate_barcode TEXT,
  crate_size    INTEGER NOT NULL DEFAULT 1 CHECK (crate_size >= 1),
  track_stock   INTEGER NOT NULL DEFAULT 1,
  stock         INTEGER NOT NULL DEFAULT 0,
  min_stock     INTEGER NOT NULL DEFAULT 0,
  location      TEXT NOT NULL DEFAULT '',
  quick_count   INTEGER NOT NULL DEFAULT 0,
  sort          INTEGER NOT NULL DEFAULT 0,
  active        INTEGER NOT NULL DEFAULT 1
);
CREATE UNIQUE INDEX products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE UNIQUE INDEX products_crate_barcode ON products(crate_barcode) WHERE crate_barcode IS NOT NULL;

CREATE TABLE shifts (
  id                   INTEGER PRIMARY KEY,
  worker_room          TEXT NOT NULL,
  worker_name          TEXT NOT NULL,
  opened_at            INTEGER NOT NULL,
  closed_at            INTEGER,
  day                  TEXT NOT NULL,
  float_cents          INTEGER NOT NULL,
  expected_float_cents INTEGER NOT NULL,
  expected_cents       INTEGER,
  counted_cents        INTEGER,
  note                 TEXT NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX one_open_shift ON shifts((1)) WHERE closed_at IS NULL;

CREATE TABLE sales (
  id            INTEGER PRIMARY KEY,
  shift_id      INTEGER REFERENCES shifts(id),
  created_at    INTEGER NOT NULL,
  day           TEXT NOT NULL,
  total_cents   INTEGER NOT NULL,
  payment       TEXT NOT NULL CHECK (payment IN ('cash', 'other')),
  voids_sale_id INTEGER REFERENCES sales(id),
  reason        TEXT NOT NULL DEFAULT ''
);
CREATE INDEX sales_shift ON sales(shift_id);
CREATE INDEX sales_day ON sales(day);
CREATE UNIQUE INDEX one_void_per_sale ON sales(voids_sale_id) WHERE voids_sale_id IS NOT NULL;

CREATE TABLE sale_lines (
  id               INTEGER PRIMARY KEY,
  sale_id          INTEGER NOT NULL REFERENCES sales(id),
  product_id       INTEGER REFERENCES products(id),
  name             TEXT NOT NULL,
  category         TEXT NOT NULL DEFAULT '',
  qty              INTEGER NOT NULL,
  unit_price_cents INTEGER NOT NULL,
  unit_cost_cents  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX sale_lines_sale ON sale_lines(sale_id);

CREATE TABLE stock_moves (
  id         INTEGER PRIMARY KEY,
  product_id INTEGER NOT NULL REFERENCES products(id),
  delta      INTEGER NOT NULL,
  kind       TEXT NOT NULL CHECK (kind IN ('sale', 'void', 'restock', 'count')),
  ref_id     INTEGER,
  created_at INTEGER NOT NULL
);
CREATE INDEX stock_moves_product ON stock_moves(product_id);

-- shift_id is NULL when no shift was open, for example the OBK emptying the box.
CREATE TABLE cash_moves (
  id           INTEGER PRIMARY KEY,
  shift_id     INTEGER REFERENCES shifts(id),
  kind         TEXT NOT NULL CHECK (kind IN ('withdrawal', 'topup', 'rental_deposit', 'rental_refund')),
  amount_cents INTEGER NOT NULL,
  note         TEXT NOT NULL DEFAULT '',
  created_at   INTEGER NOT NULL
);
CREATE INDEX cash_moves_shift ON cash_moves(shift_id);

CREATE TABLE stocktakes (
  id         INTEGER PRIMARY KEY,
  shift_id   INTEGER REFERENCES shifts(id),
  kind       TEXT NOT NULL CHECK (kind IN ('quick', 'full')),
  by_room    TEXT NOT NULL DEFAULT '',
  by_name    TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL
);

CREATE TABLE stocktake_lines (
  stocktake_id INTEGER NOT NULL REFERENCES stocktakes(id),
  product_id   INTEGER NOT NULL REFERENCES products(id),
  expected     INTEGER NOT NULL,
  counted      INTEGER NOT NULL,
  unit_cost_cents INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (stocktake_id, product_id)
);

CREATE TABLE restocks (
  id               INTEGER PRIMARY KEY,
  created_at       INTEGER NOT NULL,
  supplier         TEXT NOT NULL DEFAULT '',
  note             TEXT NOT NULL DEFAULT '',
  total_cost_cents INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE restock_lines (
  id              INTEGER PRIMARY KEY,
  restock_id      INTEGER NOT NULL REFERENCES restocks(id),
  product_id      INTEGER NOT NULL REFERENCES products(id),
  qty             INTEGER NOT NULL,
  unit_cost_cents INTEGER NOT NULL
);

CREATE TABLE rental_items (
  id            INTEGER PRIMARY KEY,
  name          TEXT NOT NULL,
  code          TEXT,
  fee_cents     INTEGER NOT NULL DEFAULT 0,
  fee_per       TEXT NOT NULL DEFAULT 'flat' CHECK (fee_per IN ('flat', 'hour')),
  deposit_cents INTEGER NOT NULL DEFAULT 0,
  active        INTEGER NOT NULL DEFAULT 1
);
CREATE UNIQUE INDEX rental_items_code ON rental_items(code) WHERE code IS NOT NULL;

CREATE TABLE rentals (
  id              INTEGER PRIMARY KEY,
  item_id         INTEGER NOT NULL REFERENCES rental_items(id),
  shift_id        INTEGER NOT NULL REFERENCES shifts(id),
  renter_name     TEXT NOT NULL,
  renter_room     TEXT NOT NULL,
  deposit_cents   INTEGER NOT NULL DEFAULT 0,
  deposit_note    TEXT NOT NULL DEFAULT '',
  started_at      INTEGER NOT NULL,
  returned_at     INTEGER,
  return_shift_id INTEGER REFERENCES shifts(id),
  fee_cents       INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX one_rental_per_item ON rentals(item_id) WHERE returned_at IS NULL;
