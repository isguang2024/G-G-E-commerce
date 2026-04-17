-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'apps'
          AND column_name = 'default_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE apps RENAME COLUMN default_space_key TO default_menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'app_host_bindings'
          AND column_name = 'default_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE app_host_bindings RENAME COLUMN default_space_key TO default_menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_spaces'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_spaces RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_space_host_bindings'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_space_host_bindings RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_space_entry_bindings'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_space_entry_bindings RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menus'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menus RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'space_menu_placements'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE space_menu_placements RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'ui_pages'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE ui_pages RENAME COLUMN space_key TO menu_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'page_space_bindings'
          AND column_name = 'space_key'
    ) THEN
        EXECUTE 'ALTER TABLE page_space_bindings RENAME COLUMN space_key TO menu_space_key';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_menu_spaces_space_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_spaces_menu_space_key
    ON menu_spaces (app_key, menu_space_key)
    WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_menu_space_host_bindings_space_key;
CREATE INDEX IF NOT EXISTS idx_menu_space_host_bindings_menu_space_key
    ON menu_space_host_bindings (menu_space_key);

DROP INDEX IF EXISTS idx_menu_space_entry_bindings_space_key;
CREATE INDEX IF NOT EXISTS idx_menu_space_entry_bindings_menu_space_key
    ON menu_space_entry_bindings (menu_space_key)
    WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_menus_space_key;
CREATE INDEX IF NOT EXISTS idx_menus_menu_space_key
    ON menus (app_key, menu_space_key);

DROP INDEX IF EXISTS idx_space_menu_placements_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_space_menu_placements_menu_space_unique
    ON space_menu_placements (app_key, menu_space_key, menu_key)
    WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_space_menu_placements_parent;
CREATE INDEX IF NOT EXISTS idx_space_menu_placements_menu_space_parent
    ON space_menu_placements (app_key, menu_space_key, parent_menu_key, sort_order);

DROP INDEX IF EXISTS idx_ui_pages_space_key;
CREATE INDEX IF NOT EXISTS idx_ui_pages_menu_space_key
    ON ui_pages (menu_space_key);

DROP INDEX IF EXISTS idx_page_space_bindings_page_space_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_page_space_bindings_page_menu_space_unique
    ON page_space_bindings (app_key, page_id, menu_space_key)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_page_space_bindings_page_menu_space_unique;
DROP INDEX IF EXISTS idx_ui_pages_menu_space_key;
DROP INDEX IF EXISTS idx_space_menu_placements_menu_space_parent;
DROP INDEX IF EXISTS idx_space_menu_placements_menu_space_unique;
DROP INDEX IF EXISTS idx_menus_menu_space_key;
DROP INDEX IF EXISTS idx_menu_space_entry_bindings_menu_space_key;
DROP INDEX IF EXISTS idx_menu_space_host_bindings_menu_space_key;
DROP INDEX IF EXISTS idx_menu_spaces_menu_space_key;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'page_space_bindings'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE page_space_bindings RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'ui_pages'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE ui_pages RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'space_menu_placements'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE space_menu_placements RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menus'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menus RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_space_entry_bindings'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_space_entry_bindings RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_space_host_bindings'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_space_host_bindings RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'menu_spaces'
          AND column_name = 'menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE menu_spaces RENAME COLUMN menu_space_key TO space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'app_host_bindings'
          AND column_name = 'default_menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE app_host_bindings RENAME COLUMN default_menu_space_key TO default_space_key';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'apps'
          AND column_name = 'default_menu_space_key'
    ) THEN
        EXECUTE 'ALTER TABLE apps RENAME COLUMN default_menu_space_key TO default_space_key';
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_spaces_space_key
    ON menu_spaces (app_key, space_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_menu_space_host_bindings_space_key
    ON menu_space_host_bindings (space_key);

CREATE INDEX IF NOT EXISTS idx_menu_space_entry_bindings_space_key
    ON menu_space_entry_bindings (space_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_menus_space_key
    ON menus (app_key, space_key);

CREATE UNIQUE INDEX IF NOT EXISTS idx_space_menu_placements_unique
    ON space_menu_placements (app_key, space_key, menu_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_space_menu_placements_parent
    ON space_menu_placements (app_key, space_key, parent_menu_key, sort_order);

CREATE INDEX IF NOT EXISTS idx_ui_pages_space_key
    ON ui_pages (space_key);

CREATE UNIQUE INDEX IF NOT EXISTS idx_page_space_bindings_page_space_unique
    ON page_space_bindings (app_key, page_id, space_key)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd
