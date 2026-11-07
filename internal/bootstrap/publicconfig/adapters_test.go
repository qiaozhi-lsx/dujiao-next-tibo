package publicconfigwiring

import (
	"reflect"
	"testing"

	"github.com/dujiao-next/internal/app/container"
	"github.com/dujiao-next/internal/constants"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/web"
)

// memorySettingStore 是 settingscontract.Store 的内存实现，只为驱动 settings.Service。
type memorySettingStore map[string]jsonmap.JSON

func (s memorySettingStore) GetByKey(key string) (jsonmap.JSON, bool, error) {
	value, ok := s[key]
	return value, ok, nil
}

func (s memorySettingStore) Upsert(key string, value jsonmap.JSON) (jsonmap.JSON, error) {
	s[key] = value
	return value, nil
}

// TestPublicConfigSettingsAdapter_FlagsFollowStoredSettings 适配器按存量设置给出导航配置与注册开关，
// 口径必须与 /api/v1/public/config 一致：nav_config 取设置原文，开关取 registration_config 的字段。
func TestPublicConfigSettingsAdapter_FlagsFollowStoredSettings(t *testing.T) {
	nav := jsonmap.JSON{
		"builtin":      map[string]interface{}{"blog": false, "notice": true, "about": true},
		"custom_items": []interface{}{},
	}
	store := memorySettingStore{
		constants.SettingKeyNavConfig: nav,
		constants.SettingKeyRegistrationConfig: jsonmap.JSON{
			constants.SettingFieldRegistrationEnabled:      true,
			constants.SettingFieldEmailVerificationEnabled: false,
		},
	}
	var src web.InitialConfigFlagsSource = publicConfigSettingsAdapter{settings: settingsapp.NewService(store)}

	if got := src.GetNavConfig(); !reflect.DeepEqual(got, nav) {
		t.Fatalf("GetNavConfig() = %#v, want stored %#v", got, nav)
	}
	registrationEnabled, emailVerificationEnabled := src.GetRegistrationFlags()
	if !registrationEnabled || emailVerificationEnabled {
		t.Fatalf("GetRegistrationFlags() = (%v, %v), want (true, false)", registrationEnabled, emailVerificationEnabled)
	}
}

// TestPublicConfigSettingsAdapter_FlagsDefaultWhenUnset 没有任何设置时：nav_config 为无类型 nil
// （让 web 包补默认结构，而不是把 nil map 塞进 interface 变成 "null"），两个开关默认开启。
func TestPublicConfigSettingsAdapter_FlagsDefaultWhenUnset(t *testing.T) {
	var src web.InitialConfigFlagsSource = publicConfigSettingsAdapter{settings: settingsapp.NewService(memorySettingStore{})}

	if got := src.GetNavConfig(); got != nil {
		t.Fatalf("GetNavConfig() = %#v, want untyped nil when nav_config is unset", got)
	}
	registrationEnabled, emailVerificationEnabled := src.GetRegistrationFlags()
	if !registrationEnabled || !emailVerificationEnabled {
		t.Fatalf("GetRegistrationFlags() = (%v, %v), want defaults (true, true)", registrationEnabled, emailVerificationEnabled)
	}
}

// TestNewInitialConfigSource 入口页的设置来源要同时满足 web.InitialConfigFlagsSource；
// 容器里没有 SettingService 时返回无类型 nil，让 RegisterUser 直接退回不注入。
func TestNewInitialConfigSource(t *testing.T) {
	if got := NewInitialConfigSource(&container.Container{}); got != nil {
		t.Fatalf("NewInitialConfigSource without SettingService = %#v, want nil", got)
	}

	src := NewInitialConfigSource(&container.Container{SettingService: settingsapp.NewService(memorySettingStore{})})
	if src == nil {
		t.Fatal("NewInitialConfigSource with SettingService = nil, want adapter")
	}
	if _, ok := src.(web.InitialConfigFlagsSource); !ok {
		t.Fatalf("NewInitialConfigSource = %T, must implement web.InitialConfigFlagsSource", src)
	}
}
