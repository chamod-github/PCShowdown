-- +goose Up
-- SQL in this section is executed when the migration is applied.
-- link to a pc
CREATE TABLE IF NOT EXISTS link (
    link_id TEXT PRIMARY KEY DEFAULT (SUBSTR(LOWER(HEX(RANDOMBLOB(16))), 0, 6)),
    permission TEXT NOT NULL CHECK(permission IN("edit", "view")),
    pc_id INTEGER NOT NULL REFERENCES pc(pc_id) ON DELETE CASCADE
);
-- info about the pc
CREATE TABLE IF NOT EXISTS pc (
    pc_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    info TEXT NOT NULL
);
-- parts of a pc
CREATE TABLE IF NOT EXISTS part(
    part_id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(
        type IN(
            "cpu",
            "cooler",
            "mobo",
            "gpu",
            "psu",
            "ram",
            "case",
            "disk",
            "other"
        )
    ),
    brand TEXT NOT NULL,
    model TEXT NOT NULL,
    qty INTEGER NOT NULL,
    pc_id INTEGER NOT NULL REFERENCES pc(pc_id) ON DELETE CASCADE
);
-- images of a PC
CREATE TABLE IF NOT EXISTS image(
    image_id INTEGER PRIMARY KEY AUTOINCREMENT,
    pc_id INTEGER NOT NULL REFERENCES pc(pc_id) ON DELETE CASCADE,
    link TEXT NOT NULL
);
-- Creating indices
CREATE INDEX idx_link_id ON link (link_id);
CREATE INDEX idx_pc_id ON pc (pc_id);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_link_id;
DROP INDEX IF EXISTS idx_pc_id;
DROP TABLE IF EXISTS link;
DROP TABLE IF EXISTS part;
DROP TABLE IF EXISTS pc;