package upload

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maben/backend/internal/modules/system/models"
)

func TestNormalizeUploadKeyExtraSchema(t *testing.T) {
	normalized, err := NormalizeUploadKeyExtraSchema(models.MetaJSON{
		"fields": []map[string]any{
			{
				"key":         "callback_scene",
				"description": "业务回调场景",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, extraSchemaVersionV1, normalized["version"])

	fields, ok := normalized["fields"].([]any)
	require.True(t, ok)
	require.Len(t, fields, 1)

	first, ok := fields[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "callback_scene", first["key"])
	require.Equal(t, "callback_scene", first["label"])
	require.Equal(t, extraSchemaFieldTypeString, first["type"])
}

func TestNormalizeUploadKeyExtraSchemaRejectsExplicitFieldConflict(t *testing.T) {
	_, err := NormalizeUploadKeyExtraSchema(models.MetaJSON{
		"version": extraSchemaVersionV1,
		"fields": []map[string]any{
			{
				"key":   "upload_mode",
				"label": "上传方式",
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "conflicts with explicit runtime field")
}

func TestNormalizeUploadRuleExtraSchemaRejectsInvalidSelect(t *testing.T) {
	_, err := NormalizeUploadRuleExtraSchema(models.MetaJSON{
		"version": extraSchemaVersionV1,
		"fields": []map[string]any{
			{
				"key":   "preset",
				"label": "处理预设",
				"type":  extraSchemaFieldTypeSelect,
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires options")
}

func TestDriverExtraDefaultsAliyunOSS(t *testing.T) {
	providerDefaults := DriverExtraDefaults(UploadProviderDriverAliyunOSS, configExtraScopeProvider)
	require.Equal(t, false, providerDefaults[ossProviderExtraUseCNameKey])
	require.Equal(t, defaultOSSSTSRoleSessionName, providerDefaults[ossProviderExtraSTSSessionNameKey])
	require.Equal(t, defaultOSSSTSRoleDurationSeconds, providerDefaults[ossProviderExtraSTSDurationSecondsKey])

	bucketDefaults := DriverExtraDefaults(UploadProviderDriverAliyunOSS, configExtraScopeBucket)
	require.Equal(t, defaultOSSDirectUploadSuccessAction, bucketDefaults[ossBucketExtraSuccessStatusKey])
}
