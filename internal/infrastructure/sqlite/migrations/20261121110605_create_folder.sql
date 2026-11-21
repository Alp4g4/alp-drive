-- +goose Up
CREATE TABLE folder (
    id         TEXT      NOT NULL,
    name       TEXT      NOT NULL,
    parent_id  TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT pk_folder PRIMARY KEY (id),
    CONSTRAINT fk_folder_parent_id
        FOREIGN KEY (parent_id) REFERENCES folder (id) ON DELETE CASCADE
);

CREATE INDEX idx_folder_parent_id_name_id ON folder (parent_id, name, id);

-- +goose Down
DROP TABLE folder;
