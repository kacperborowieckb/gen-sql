CREATE TABLE IF NOT EXISTS generation_projects (
    project_id TEXT PRIMARY KEY,
    ddl_schema TEXT NOT NULL,
    instructions TEXT,
    rows_to_generate INTEGER DEFAULT 100 CHECK (rows_to_generate >= 0),
    temperature REAL DEFAULT 1.0 CHECK (temperature >= 0.0 AND temperature <= 2.0),
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
