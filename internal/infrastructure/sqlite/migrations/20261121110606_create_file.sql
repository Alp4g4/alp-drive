-- +goose Up
CREATE TABLE file (
    id         TEXT      NOT NULL,
    name       TEXT      NOT NULL,
    folder_id  TEXT      NOT NULL,
    size       INTEGER   NOT NULL,
    mime_type  TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT pk_file PRIMARY KEY (id),
    CONSTRAINT fk_file_folder_id
        FOREIGN KEY (folder_id) REFERENCES folder (id) ON DELETE CASCADE
);

CREATE INDEX idx_file_folder_id_name_id ON file (folder_id, name, id);

-- +goose Down
DROP TABLE file;
