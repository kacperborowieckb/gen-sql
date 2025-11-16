CREATE TABLE IF NOT EXISTS generation_projects (
    project_id TEXT PRIMARY KEY,
    ddl_schema TEXT NOT NULL,
    generation_instructions TEXT,
    max_rows INTEGER DEFAULT 100 CHECK (max_rows >= 0),
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);