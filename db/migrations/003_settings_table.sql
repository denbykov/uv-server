CREATE TABLE IF NOT EXISTS settings (
    storage_dir TEXT NOT NULL DEFAULT './storage'
);

INSERT INTO settings (storage_dir) VALUES ('./storage');
