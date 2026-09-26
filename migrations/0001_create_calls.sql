CREATE TABLE calls (
    id                    integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    filename              text NOT NULL,
    transcript            text,
    flags                 jsonb DEFAULT '[]'::jsonb,
    flag_count            integer DEFAULT 0,
    is_pushy              boolean DEFAULT false,
    score                 integer DEFAULT 100,
    created_at            timestamptz DEFAULT CURRENT_TIMESTAMP,
    trackdrive_url        text,
    agent_name            text,
    disposition           varchar(100) DEFAULT '',
    offer_name            varchar(255) DEFAULT '',
    agent_talk_time       integer DEFAULT 0,
    forward_duration      integer DEFAULT 0,
    status                text NOT NULL DEFAULT 'pending',
    attempts              integer NOT NULL DEFAULT 0,
    processing_started_at timestamptz,
    last_error            text,

    CONSTRAINT calls_status_check
        CHECK (status IN ('pending', 'processing', 'complete', 'failed'))
);

CREATE INDEX idx_calls_pending
    ON calls (status)
    WHERE status = 'pending';

CREATE INDEX idx_calls_processing
    ON calls (status, processing_started_at)
    WHERE status = 'processing';
