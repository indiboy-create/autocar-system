package store

const SchemaVersion = 1

const schemaSQL = `
CREATE TABLE IF NOT EXISTS brand (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    firstletter TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    logo_local_path TEXT NOT NULL DEFAULT '',
    state INTEGER NOT NULL DEFAULT 0,
    havenewenergy INTEGER NOT NULL DEFAULT 0,
    api_sort INTEGER NOT NULL DEFAULT 0,
    collected_at TEXT NOT NULL DEFAULT '',
    user_updated_at TEXT NOT NULL DEFAULT '',
    media_updated_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS firm (
    id INTEGER PRIMARY KEY,
    brand_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    py TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    logo_local_path TEXT NOT NULL DEFAULT '',
    collected_at TEXT NOT NULL DEFAULT '',
    user_updated_at TEXT NOT NULL DEFAULT '',
    media_updated_at TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (brand_id) REFERENCES brand(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS series (
    id INTEGER PRIMARY KEY,
    brand_id INTEGER NOT NULL,
    firm_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL DEFAULT 0,
    group_name TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    state INTEGER NOT NULL DEFAULT 0,
    price_min REAL NOT NULL DEFAULT 0,
    price_max REAL NOT NULL DEFAULT 0,
    place TEXT NOT NULL DEFAULT '',
    py TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    logo_local_path TEXT NOT NULL DEFAULT '',
    level_id INTEGER NOT NULL DEFAULT 0,
    level_name TEXT NOT NULL DEFAULT '',
    newenergy INTEGER NOT NULL DEFAULT 0,
    is_foreign_car INTEGER NOT NULL DEFAULT 0,
    is_top INTEGER NOT NULL DEFAULT 0,
    fuel_types TEXT NOT NULL DEFAULT '',
    cartypename TEXT NOT NULL DEFAULT '',
    tag1 TEXT NOT NULL DEFAULT '',
    rank INTEGER NOT NULL DEFAULT 0,
    uv_rank INTEGER NOT NULL DEFAULT 0,
    hot_rank INTEGER NOT NULL DEFAULT 0,
    collected_at TEXT NOT NULL DEFAULT '',
    user_updated_at TEXT NOT NULL DEFAULT '',
    media_updated_at TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (brand_id) REFERENCES brand(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    FOREIGN KEY (firm_id) REFERENCES firm(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_series_identity
ON series(id, brand_id, firm_id);

CREATE TABLE IF NOT EXISTS spec (
    id INTEGER PRIMARY KEY,
    brand_id INTEGER NOT NULL,
    firm_id INTEGER NOT NULL,
    series_id INTEGER NOT NULL,
    group_name TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    series_name TEXT NOT NULL DEFAULT '',
    model_year TEXT NOT NULL DEFAULT '',
    body_form TEXT NOT NULL DEFAULT '',
    price_min REAL NOT NULL DEFAULT 0,
    price_max REAL NOT NULL DEFAULT 0,
    news_price REAL NOT NULL DEFAULT 0,
    news_id INTEGER NOT NULL DEFAULT 0,
    dealer_id INTEGER NOT NULL DEFAULT 0,
    sale_type TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    image_local_path TEXT NOT NULL DEFAULT '',
    image_sha256 TEXT NOT NULL DEFAULT '',
    white_image_url TEXT NOT NULL DEFAULT '',
    white_image_local_path TEXT NOT NULL DEFAULT '',
    white_image_sha256 TEXT NOT NULL DEFAULT '',
    collected_at TEXT NOT NULL DEFAULT '',
    user_updated_at TEXT NOT NULL DEFAULT '',
    media_updated_at TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (series_id, brand_id, firm_id)
        REFERENCES series(id, brand_id, firm_id)
        ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS category (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    sort INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS youvision (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    brand_id INTEGER NOT NULL,
    firm_id INTEGER NOT NULL,
    series_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    cover_image TEXT NOT NULL DEFAULT '',
    images TEXT NOT NULL DEFAULT '[]',
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT '',
    UNIQUE (series_id, category_id, name),
    FOREIGN KEY (series_id, brand_id, firm_id)
        REFERENCES series(id, brand_id, firm_id)
        ON DELETE RESTRICT ON UPDATE RESTRICT,
    FOREIGN KEY (category_id) REFERENCES category(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_youvision_identity
ON youvision(id, brand_id, firm_id, series_id, category_id);

CREATE TABLE IF NOT EXISTS gallery (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    youvision_id INTEGER NOT NULL,
    brand_id INTEGER NOT NULL,
    firm_id INTEGER NOT NULL,
    series_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL CHECK(length(trim(title)) > 0),
    cover_image TEXT NOT NULL DEFAULT '',
    images TEXT NOT NULL DEFAULT '[]',
    image_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT '',
    UNIQUE (youvision_id, title),
    FOREIGN KEY (youvision_id, brand_id, firm_id, series_id, category_id)
        REFERENCES youvision(id, brand_id, firm_id, series_id, category_id)
        ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE INDEX IF NOT EXISTS ix_firm_brand_id ON firm(brand_id);
CREATE INDEX IF NOT EXISTS ix_firm_name ON firm(name);
CREATE INDEX IF NOT EXISTS ix_series_brand_id ON series(brand_id);
CREATE INDEX IF NOT EXISTS ix_series_firm_id ON series(firm_id);
CREATE INDEX IF NOT EXISTS ix_series_name ON series(name);
CREATE INDEX IF NOT EXISTS ix_spec_series_id ON spec(series_id);
CREATE INDEX IF NOT EXISTS ix_spec_model_year ON spec(model_year);
CREATE INDEX IF NOT EXISTS ix_youvision_series_category ON youvision(series_id, category_id);
CREATE INDEX IF NOT EXISTS ix_youvision_name ON youvision(name);
CREATE INDEX IF NOT EXISTS ix_gallery_youvision_id ON gallery(youvision_id);
CREATE INDEX IF NOT EXISTS ix_gallery_title ON gallery(title);
`

func schemaStatements() []string {
	raw := splitSQL(schemaSQL)
	statements := make([]string, 0, len(raw))
	for _, statement := range raw {
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
