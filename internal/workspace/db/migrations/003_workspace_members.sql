CREATE TABLE workspace.members (
    workspace_id UUID        NOT NULL REFERENCES workspace.workspaces(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES auth.users(id),
    role         TEXT        NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);
