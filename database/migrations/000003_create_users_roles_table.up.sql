CREATE TABLE user_roles (
    user_id uuid    NOT NULL,
    role_id integer NOT NULL,

    PRIMARY KEY (user_id, role_id),

    CONSTRAINT fk_user_id
        FOREIGN KEY (user_id)
            REFERENCES users (id)  -- Corrected to reference 'id' from users table
            ON DELETE CASCADE,

    CONSTRAINT fk_role_id
        FOREIGN KEY (role_id)
            REFERENCES roles (id)
            ON DELETE CASCADE
);
