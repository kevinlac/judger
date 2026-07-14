CREATE TABLE IF NOT EXISTS problems (
    id TEXT NOT NULL,
    time_limit_ms INT NOT NULL,
    memory_limit_mb INT NOT NULL,
    problem_type TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT pk_problem_id PRIMARY KEY(id),
    CONSTRAINT timelimit_nonzero CHECK (time_limit_ms > 0),
    CONSTRAINT memory_limit_nonzero CHECK (memory_limit_mb > 0),
    CONSTRAINT chk_problem_type CHECK (problem_type IN ('standard', 'custom'))
);

CREATE TABLE IF NOT EXISTS submissions (
    id UUID NOT NULL, -- supplied for path to user's submitted code
    problem_id TEXT NOT NULL,
    lang TEXT NOT NULL,
    processing_status TEXT NOT NULL,
    verdict TEXT,
    submitted_at TIMESTAMP DEFAULT NOW(),
    judged_at TIMESTAMP,
    
    CONSTRAINT pk_submission_id PRIMARY KEY(id),
    CONSTRAINT fk_problem_id FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE RESTRICT,
    CONSTRAINT chk_language CHECK (lang IN ('C++', 'C', 'Java', 'Python')),
    CONSTRAINT chk_sub_status CHECK (processing_status IN ('queued', 'running', 'judged')),
    CONSTRAINT chk_sub_verdict CHECK (verdict IN ('AC', 'WA', 'TLE', 'MLE', 'RTE', 'IE', 'CE'))
);