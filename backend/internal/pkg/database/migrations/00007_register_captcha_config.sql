-- +goose Up
-- +goose StatementBegin
-- 注册体系扩展：给 register_policies 追加人机验证提供商配置字段。
-- captcha_provider 取值：none | recaptcha | hcaptcha | turnstile
-- captcha_site_key 存放对应提供商的公开 site_key（前端渲染 captcha widget 用）

ALTER TABLE register_policies
    ADD COLUMN IF NOT EXISTS captcha_provider varchar(32)  NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS captcha_site_key varchar(256) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE register_policies
    DROP COLUMN IF EXISTS captcha_site_key,
    DROP COLUMN IF EXISTS captcha_provider;
-- +goose StatementEnd
