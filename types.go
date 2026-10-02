package anthias

import (
	"encoding/json"
	"errors"
	"time"
)

// Asset is returned by list/get/create/update/replace.
type Asset struct {
	AssetID string `json:"asset_id"`
	Name    string `json:"name"`
	URI     string `json:"uri"`
	// StartDate and EndDate are the zero time when the player stores no date.
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Duration       int       `json:"duration"`
	Mimetype       string    `json:"mimetype"`
	IsEnabled      bool      `json:"is_enabled"`
	NoCache        bool      `json:"nocache"`
	PlayOrder      int       `json:"play_order"`
	SkipAssetCheck bool      `json:"skip_asset_check"`
	SkipSSLVerify  bool      `json:"skip_ssl_verify"`
	IsActive       bool      `json:"is_active"`
	IsProcessing   bool      `json:"is_processing"`
	// PlayDays lists the ISO weekdays the asset plays on (1 = Monday … 7 = Sunday).
	PlayDays []int `json:"play_days"`
	// PlayTimeFrom and PlayTimeTo bound the daily play window ("HH:MM:SS");
	// both are nil when the asset has no time-of-day window.
	PlayTimeFrom          *string           `json:"play_time_from"`
	PlayTimeTo            *string           `json:"play_time_to"`
	IsReachable           bool              `json:"is_reachable"`
	LastReachabilityCheck *time.Time        `json:"last_reachability_check"`
	Metadata              map[string]any    `json:"metadata"`
	RefreshIntervalS      int               `json:"refresh_interval_s"`
	CustomHeaders         map[string]string `json:"custom_headers"`
}

// CreateAssetRequest is the body of POST /api/v2/assets.
// Required by the server: name, uri, start_date, end_date, duration, mimetype,
// is_enabled.
type CreateAssetRequest struct {
	Name           string    `json:"name"`
	URI            string    `json:"uri"`
	Ext            string    `json:"ext,omitempty"` // write-only; from a file upload response
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Duration       int       `json:"duration"` // seconds, 0–31536000
	Mimetype       string    `json:"mimetype"`
	IsEnabled      bool      `json:"is_enabled"`
	IsProcessing   *bool     `json:"is_processing,omitempty"`
	NoCache        *bool     `json:"nocache,omitempty"`
	PlayOrder      *int      `json:"play_order,omitempty"`
	SkipAssetCheck *bool     `json:"skip_asset_check,omitempty"`
	SkipSSLVerify  *bool     `json:"skip_ssl_verify,omitempty"`
	// PlayDays: ISO weekdays 1 (Monday) – 7 (Sunday); must not be empty if set.
	PlayDays []int `json:"play_days,omitempty"`
	// PlayTimeFrom and PlayTimeTo ("HH:MM" or "HH:MM:SS") must be set together.
	PlayTimeFrom     *string           `json:"play_time_from,omitempty"`
	PlayTimeTo       *string           `json:"play_time_to,omitempty"`
	RefreshIntervalS *int              `json:"refresh_interval_s,omitempty"` // webpage auto-refresh, 0–86400
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`     // request headers for webpage assets
}

// ReplaceAssetRequest is the body of PUT /api/v2/assets/{id}. An asset's
// uri and mimetype cannot be changed after creation; create a new asset
// instead. Required by the server: name, start_date, end_date, duration,
// is_enabled.
type ReplaceAssetRequest struct {
	Name             string            `json:"name"`
	StartDate        time.Time         `json:"start_date"`
	EndDate          time.Time         `json:"end_date"`
	Duration         int               `json:"duration"` // seconds, 0–31536000
	IsEnabled        bool              `json:"is_enabled"`
	IsProcessing     *bool             `json:"is_processing,omitempty"`
	NoCache          *bool             `json:"nocache,omitempty"`
	PlayOrder        *int              `json:"play_order,omitempty"`
	SkipAssetCheck   *bool             `json:"skip_asset_check,omitempty"`
	SkipSSLVerify    *bool             `json:"skip_ssl_verify,omitempty"`
	PlayDays         []int             `json:"play_days,omitempty"`
	PlayTimeFrom     *string           `json:"play_time_from,omitempty"`
	PlayTimeTo       *string           `json:"play_time_to,omitempty"`
	RefreshIntervalS *int              `json:"refresh_interval_s,omitempty"`
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`
}

// UpdateAssetRequest is the body of PATCH /api/v2/assets/{id} (partial
// update). Nil fields are left unchanged. An asset's uri and mimetype cannot
// be changed after creation; create a new asset instead.
type UpdateAssetRequest struct {
	Name             *string           `json:"name,omitempty"`
	StartDate        *time.Time        `json:"start_date,omitempty"`
	EndDate          *time.Time        `json:"end_date,omitempty"`
	Duration         *int              `json:"duration,omitempty"`
	IsEnabled        *bool             `json:"is_enabled,omitempty"`
	IsProcessing     *bool             `json:"is_processing,omitempty"`
	NoCache          *bool             `json:"nocache,omitempty"`
	PlayOrder        *int              `json:"play_order,omitempty"`
	SkipAssetCheck   *bool             `json:"skip_asset_check,omitempty"`
	SkipSSLVerify    *bool             `json:"skip_ssl_verify,omitempty"`
	PlayDays         []int             `json:"play_days,omitempty"`
	PlayTimeFrom     *string           `json:"play_time_from,omitempty"`
	PlayTimeTo       *string           `json:"play_time_to,omitempty"`
	RefreshIntervalS *int              `json:"refresh_interval_s,omitempty"`
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`

	// ClearPlayTimeWindow removes the daily play window (sends null for
	// play_time_from and play_time_to). Cannot be combined with
	// PlayTimeFrom/PlayTimeTo.
	ClearPlayTimeWindow bool `json:"-"`
	// ClearCustomHeaders removes all custom headers. Cannot be combined
	// with CustomHeaders.
	ClearCustomHeaders bool `json:"-"`
}

// MarshalJSON encodes the request, expressing the Clear* flags as the
// explicit null / empty-object values the player expects.
func (r UpdateAssetRequest) MarshalJSON() ([]byte, error) {
	if r.ClearPlayTimeWindow && (r.PlayTimeFrom != nil || r.PlayTimeTo != nil) {
		return nil, errors.New("anthias: ClearPlayTimeWindow conflicts with PlayTimeFrom/PlayTimeTo")
	}
	if r.ClearCustomHeaders && len(r.CustomHeaders) > 0 {
		return nil, errors.New("anthias: ClearCustomHeaders conflicts with CustomHeaders")
	}
	type plain UpdateAssetRequest
	b, err := json.Marshal(plain(r))
	if err != nil || (!r.ClearPlayTimeWindow && !r.ClearCustomHeaders) {
		return b, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	if r.ClearPlayTimeWindow {
		fields["play_time_from"] = json.RawMessage("null")
		fields["play_time_to"] = json.RawMessage("null")
	}
	if r.ClearCustomHeaders {
		fields["custom_headers"] = json.RawMessage("{}")
	}
	return json.Marshal(fields)
}

// FileUpload is the response of POST /api/v2/file_asset.
type FileUpload struct {
	URI      string `json:"uri"`
	Ext      string `json:"ext"`
	UploadID string `json:"upload_id"`
}

// AssetContent is returned by GET /api/v2/assets/{id}/content.
// Type is "file" or "url". For "file", Filename, Mimetype and Content
// (base64) are set. For "url", URL is set.
type AssetContent struct {
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Filename string `json:"filename,omitempty"`
	Mimetype string `json:"mimetype,omitempty"`
	Content  string `json:"content,omitempty"`
}

// DeviceSettings is returned by GET /api/v2/device_settings.
type DeviceSettings struct {
	PlayerName               string `json:"player_name"`
	AudioOutput              string `json:"audio_output"` // "hdmi" or "local"
	DefaultDuration          int    `json:"default_duration"`
	DefaultStreamingDuration int    `json:"default_streaming_duration"`
	DateFormat               string `json:"date_format"`
	// Timezone is an IANA zone name; empty means the host's zone.
	Timezone        string `json:"timezone"`
	AuthBackend     string `json:"auth_backend"`
	ShowSplash      bool   `json:"show_splash"`
	DefaultAssets   bool   `json:"default_assets"`
	ShufflePlaylist bool   `json:"shuffle_playlist"`
	Use24HourClock  bool   `json:"use_24_hour_clock"`
	DebugLogging    bool   `json:"debug_logging"`
	PreferDarkMode  bool   `json:"prefer_dark_mode"`
	VerifySSL       bool   `json:"verify_ssl"`
	ScreenRotation  int    `json:"screen_rotation"` // 0, 90, 180 or 270
	Username        string `json:"username"`
	// Scheduled display power. Times are "HH:MM" in the device timezone;
	// DisplayPowerDays is a comma-separated list of weekdays (Monday = 0)
	// on which an on-period begins.
	DisplayPowerScheduleEnabled bool   `json:"display_power_schedule_enabled"`
	DisplayPowerOnTime          string `json:"display_power_on_time"`
	DisplayPowerOffTime         string `json:"display_power_off_time"`
	DisplayPowerDays            string `json:"display_power_days"`
}

// UpdateDeviceSettingsRequest is the body of PATCH /api/v2/device_settings.
// All fields are optional; nil fields are left unchanged.
type UpdateDeviceSettingsRequest struct {
	PlayerName                  *string `json:"player_name,omitempty"`
	AudioOutput                 *string `json:"audio_output,omitempty"`
	DefaultDuration             *int    `json:"default_duration,omitempty"`
	DefaultStreamingDuration    *int    `json:"default_streaming_duration,omitempty"`
	DateFormat                  *string `json:"date_format,omitempty"`
	Timezone                    *string `json:"timezone,omitempty"` // Ptr("") resets to the host zone
	ShowSplash                  *bool   `json:"show_splash,omitempty"`
	DefaultAssets               *bool   `json:"default_assets,omitempty"`
	ShufflePlaylist             *bool   `json:"shuffle_playlist,omitempty"`
	Use24HourClock              *bool   `json:"use_24_hour_clock,omitempty"`
	DebugLogging                *bool   `json:"debug_logging,omitempty"`
	PreferDarkMode              *bool   `json:"prefer_dark_mode,omitempty"`
	VerifySSL                   *bool   `json:"verify_ssl,omitempty"`
	ScreenRotation              *int    `json:"screen_rotation,omitempty"` // 0, 90, 180 or 270
	Username                    *string `json:"username,omitempty"`
	Password                    *string `json:"password,omitempty"`
	Password2                   *string `json:"password_2,omitempty"`
	AuthBackend                 *string `json:"auth_backend,omitempty"` // "" or "auth_basic"
	CurrentPassword             *string `json:"current_password,omitempty"`
	DisplayPowerScheduleEnabled *bool   `json:"display_power_schedule_enabled,omitempty"`
	DisplayPowerOnTime          *string `json:"display_power_on_time,omitempty"`  // "HH:MM"
	DisplayPowerOffTime         *string `json:"display_power_off_time,omitempty"` // "HH:MM"
	DisplayPowerDays            *string `json:"display_power_days,omitempty"`     // e.g. "0,1,2,3,4"; Ptr("") = every day
}

// Uptime is the uptime reported by GET /api/v2/info.
type Uptime struct {
	Days  int     `json:"days"`
	Hours float64 `json:"hours"`
}

// Memory is the memory usage reported by GET /api/v2/info.
type Memory struct {
	Total     int  `json:"total"`
	Used      int  `json:"used"`
	Free      int  `json:"free"`
	Shared    int  `json:"shared"`
	Buff      int  `json:"buff"`
	Available int  `json:"available"`
	LowRAM    bool `json:"low_ram"`
}

// StorageSMART is SMART detail for a SATA/NVMe device.
type StorageSMART struct {
	Supported   bool      `json:"supported"`
	Device      string    `json:"device"`
	CheckedAt   time.Time `json:"checked_at"`
	Model       *string   `json:"model"`
	Firmware    *string   `json:"firmware"`
	Passed      *bool     `json:"passed"`
	WearPct     *int      `json:"wear_pct"`
	WearIsExact bool      `json:"wear_is_exact"`
	// WearIsAdvisory is true when WearPct came from an ATA vendor
	// attribute, whose direction is convention rather than spec: show it,
	// but don't alert on it alone.
	WearIsAdvisory     bool    `json:"wear_is_advisory"`
	PreEOL             *string `json:"pre_eol"`
	TemperatureC       *int    `json:"temperature_c"`
	PowerOnHours       *int    `json:"power_on_hours"`
	MediaErrors        *int    `json:"media_errors"`
	ReallocatedSectors *int    `json:"reallocated_sectors"`
	PendingSectors     *int    `json:"pending_sectors"`
}

// StorageMedia describes the storage device. Kind is "sd", "emmc", "disk"
// or "unknown"; wear fields are populated on eMMC and SMART-capable disks.
type StorageMedia struct {
	Kind           string        `json:"kind"`
	Name           *string       `json:"name"`
	Manufacturer   *string       `json:"manufacturer"`
	ManufacturerID *int          `json:"manufacturer_id"`
	Manufactured   *string       `json:"manufactured"`
	WearPct        *int          `json:"wear_pct"`
	WearIsExact    bool          `json:"wear_is_exact"`
	WearIsAdvisory bool          `json:"wear_is_advisory"` // see StorageSMART.WearIsAdvisory
	PreEOL         *string       `json:"pre_eol"`
	SMART          *StorageSMART `json:"smart"`
}

// Storage is the health of the filesystem the player runs from. Check
// Supported first; when false no other field carries information. Branch
// on Status: "ok", "wear", "errors", "full", "failing" or "unknown".
type Storage struct {
	Supported            bool         `json:"supported"`
	Status               string       `json:"status"`
	Device               *string      `json:"device"`
	Disk                 *string      `json:"disk"`
	MountPoint           *string      `json:"mount_point"`
	FSType               *string      `json:"fstype"`
	ReadOnly             bool         `json:"read_only"`
	ErrorStatsSupported  bool         `json:"error_stats_supported"`
	ErrorsCount          int          `json:"errors_count"`
	ErrorsNew            int          `json:"errors_new"`
	ErrorsThisBoot       bool         `json:"errors_this_boot"`
	FirstError           *time.Time   `json:"first_error"`
	LastError            *time.Time   `json:"last_error"`
	LastErrorFunction    *string      `json:"last_error_function"`
	LastCheck            *time.Time   `json:"last_check"`
	WriteOK              *bool        `json:"write_ok"`
	WriteReason          *string      `json:"write_reason"`
	WriteFailCount       int          `json:"write_fail_count"`
	WriteFailedSinceBoot bool         `json:"write_failed_since_boot"`
	FirstWriteFail       *time.Time   `json:"first_write_fail"`
	LastWriteFail        *time.Time   `json:"last_write_fail"`
	FsyncMS              *float64     `json:"fsync_ms"`
	LifetimeWrittenKB    *int64       `json:"lifetime_written_kb"`
	Media                StorageMedia `json:"media"`
}

// DeviceTime is the player's clock as reported by GET /api/v2/info.
type DeviceTime struct {
	ISO      string `json:"iso"`
	Timezone string `json:"timezone"`
	Offset   string `json:"offset"` // e.g. "UTC+03:00"
}

// UnderVoltage is power-supply health (Raspberry Pi only). Check Supported
// first; counters reset on reboot.
type UnderVoltage struct {
	Supported     bool       `json:"supported"`
	Active        bool       `json:"active"`
	SeenSinceBoot bool       `json:"seen_since_boot"`
	FirstSeen     *time.Time `json:"first_seen"`
	LastSeen      *time.Time `json:"last_seen"`
	Count         int        `json:"count"`
}

// Info is returned by GET /api/v2/info.
type Info struct {
	Viewlog        string       `json:"viewlog"`
	Loadavg        float64      `json:"loadavg"`
	FreeSpace      string       `json:"free_space"`
	DisplayPower   *string      `json:"display_power"` // nullable
	UpToDate       bool         `json:"up_to_date"`
	AnthiasVersion string       `json:"anthias_version"`
	DeviceModel    string       `json:"device_model"`
	Uptime         Uptime       `json:"uptime"`
	Memory         Memory       `json:"memory"`
	IPAddresses    []string     `json:"ip_addresses"` // URLs, e.g. "http://10.0.0.108"
	MACAddress     string       `json:"mac_address"`
	HostUser       string       `json:"host_user"`
	Storage        Storage      `json:"storage"`
	Time           DeviceTime   `json:"time"`
	UnderVoltage   UnderVoltage `json:"under_voltage"`
}

// Integrations is returned by GET /api/v2/integrations.
type Integrations struct {
	IsBalena                bool    `json:"is_balena"`
	BalenaDeviceID          *string `json:"balena_device_id,omitempty"`
	BalenaAppID             *string `json:"balena_app_id,omitempty"`
	BalenaAppName           *string `json:"balena_app_name,omitempty"`
	BalenaSupervisorVersion *string `json:"balena_supervisor_version,omitempty"`
	BalenaHostOSVersion     *string `json:"balena_host_os_version,omitempty"`
	BalenaDeviceNameAtInit  *string `json:"balena_device_name_at_init,omitempty"`
}

// ViewerPlaylist is returned by GET /api/v2/viewer/playlist.
type ViewerPlaylist struct {
	// Assets are the currently active assets in play order.
	Assets []Asset `json:"assets"`
	// Deadline is when the playlist next changes and should be fetched
	// again; nil when nothing is scheduled to change.
	Deadline *time.Time `json:"deadline"`
	// Now is the player's clock at evaluation time.
	Now time.Time `json:"now"`
}

// ViewerSettings is returned by GET /api/v2/viewer/settings.
type ViewerSettings struct {
	ShufflePlaylist bool   `json:"shuffle_playlist"`
	ShowSplash      bool   `json:"show_splash"`
	ScreenRotation  int    `json:"screen_rotation"` // 0, 90, 180 or 270
	AudioOutput     string `json:"audio_output"`
	DebugLogging    bool   `json:"debug_logging"`
}

// ScreenlyValidation is returned by [Client.ValidateScreenlyToken].
// Valid means the token works and the migration asset group is ready.
type ScreenlyValidation struct {
	Valid           bool    `json:"valid"`
	AssetGroupID    *string `json:"asset_group_id"`
	AssetGroupTitle *string `json:"asset_group_title"`
	Error           *string `json:"error"`
}

// ScreenlyMigrateRequest is the body of POST /api/v2/integrations/screenly/migrate.
type ScreenlyMigrateRequest struct {
	Token   string `json:"token"`
	AssetID string `json:"asset_id"`
	// AssetGroupID places the asset in a Screenly asset group, typically
	// [ScreenlyValidation.AssetGroupID].
	AssetGroupID string `json:"asset_group_id,omitempty"`
}

// ScreenlyMigration is returned by [Client.MigrateAssetToScreenly].
type ScreenlyMigration struct {
	Success         bool    `json:"success"`
	ScreenlyAssetID *string `json:"screenly_asset_id"`
	Error           *string `json:"error"`
}

// ImportMediaItem is one media item on an import provider.
type ImportMediaItem struct {
	RemoteID  string `json:"remote_id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	// Importable is false for media Anthias cannot play (audio,
	// documents); SkipReason then explains why.
	Importable bool    `json:"importable"`
	SkipReason *string `json:"skip_reason"`
}

// ImportValidation is returned by [Client.ValidateImportToken].
type ImportValidation struct {
	Valid bool              `json:"valid"`
	Items []ImportMediaItem `json:"items"`
	Error *string           `json:"error"`
}

// ImportItemRequest is the body of POST /api/v2/integrations/import/{provider}/item.
type ImportItemRequest struct {
	Token    string `json:"token"`
	RemoteID string `json:"remote_id"`
	// Enable sets whether the new asset is enabled; the player defaults to true.
	Enable *bool `json:"enable,omitempty"`
}

// ImportResult is returned by [Client.ImportItem]. A re-import of an item
// the player already holds is Success with AssetID pointing at the
// existing asset and Skipped set.
type ImportResult struct {
	Success bool    `json:"success"`
	AssetID *string `json:"asset_id"`
	Skipped bool    `json:"skipped"`
	Reason  *string `json:"reason"`
	Error   *string `json:"error"`
}

// Ptr returns a pointer to v. Useful for populating pointer request fields.
func Ptr[T any](v T) *T {
	return &v
}
