package anthias

import "time"

// Asset is returned by list/get/create/update/replace.
type Asset struct {
	AssetID        string    `json:"asset_id"`
	Name           string    `json:"name"`
	URI            string    `json:"uri"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Duration       int       `json:"duration"`
	Mimetype       string    `json:"mimetype"`
	IsEnabled      bool      `json:"is_enabled"`
	NoCache        bool      `json:"nocache"`
	PlayOrder      int       `json:"play_order"`
	SkipAssetCheck bool      `json:"skip_asset_check"`
	IsActive       bool      `json:"is_active"`
	IsProcessing   bool      `json:"is_processing"`
}

// CreateAssetRequest is the body of POST /api/v2/assets and PUT /api/v2/assets/{id}.
// Required by the server: name, uri, start_date, end_date, duration, mimetype,
// is_enabled.
type CreateAssetRequest struct {
	Name           string    `json:"name"`
	URI            string    `json:"uri"`
	Ext            string    `json:"ext,omitempty"` // write-only; from a file upload response
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Duration       int       `json:"duration"`
	Mimetype       string    `json:"mimetype"`
	IsEnabled      bool      `json:"is_enabled"`
	IsProcessing   *bool     `json:"is_processing,omitempty"`
	NoCache        *bool     `json:"nocache,omitempty"`
	PlayOrder      *int      `json:"play_order,omitempty"`
	SkipAssetCheck *bool     `json:"skip_asset_check,omitempty"`
}

// UpdateAssetRequest is the body of PATCH /api/v2/assets/{id} (partial update).
// All fields are pointers with omitempty.
type UpdateAssetRequest struct {
	Name           *string    `json:"name,omitempty"`
	URI            *string    `json:"uri,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Duration       *int       `json:"duration,omitempty"`
	Mimetype       *string    `json:"mimetype,omitempty"`
	IsEnabled      *bool      `json:"is_enabled,omitempty"`
	IsProcessing   *bool      `json:"is_processing,omitempty"`
	NoCache        *bool      `json:"nocache,omitempty"`
	SkipAssetCheck *bool      `json:"skip_asset_check,omitempty"`
}

// FileUpload is the response of POST /api/v2/file_asset.
type FileUpload struct {
	URI string `json:"uri"`
	Ext string `json:"ext"`
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
	AudioOutput              string `json:"audio_output"`
	DefaultDuration          int    `json:"default_duration"`
	DefaultStreamingDuration int    `json:"default_streaming_duration"`
	DateFormat               string `json:"date_format"`
	AuthBackend              string `json:"auth_backend"`
	ShowSplash               bool   `json:"show_splash"`
	DefaultAssets            bool   `json:"default_assets"`
	ShufflePlaylist          bool   `json:"shuffle_playlist"`
	Use24HourClock           bool   `json:"use_24_hour_clock"`
	DebugLogging             bool   `json:"debug_logging"`
	Username                 string `json:"username"`
}

// UpdateDeviceSettingsRequest is the body of PATCH /api/v2/device_settings.
// All fields are optional.
type UpdateDeviceSettingsRequest struct {
	PlayerName               *string `json:"player_name,omitempty"`
	AudioOutput              *string `json:"audio_output,omitempty"`
	DefaultDuration          *int    `json:"default_duration,omitempty"`
	DefaultStreamingDuration *int    `json:"default_streaming_duration,omitempty"`
	DateFormat               *string `json:"date_format,omitempty"`
	ShowSplash               *bool   `json:"show_splash,omitempty"`
	DefaultAssets            *bool   `json:"default_assets,omitempty"`
	ShufflePlaylist          *bool   `json:"shuffle_playlist,omitempty"`
	Use24HourClock           *bool   `json:"use_24_hour_clock,omitempty"`
	DebugLogging             *bool   `json:"debug_logging,omitempty"`
	Username                 *string `json:"username,omitempty"`
	Password                 *string `json:"password,omitempty"`
	Password2                *string `json:"password_2,omitempty"`
	AuthBackend              *string `json:"auth_backend,omitempty"` // "" or "auth_basic"
	CurrentPassword          *string `json:"current_password,omitempty"`
}

// Uptime is the uptime reported by GET /api/v2/info.
type Uptime struct {
	Days  int     `json:"days"`
	Hours float64 `json:"hours"`
}

// Memory is the memory usage reported by GET /api/v2/info.
type Memory struct {
	Total     int `json:"total"`
	Used      int `json:"used"`
	Free      int `json:"free"`
	Shared    int `json:"shared"`
	Buff      int `json:"buff"`
	Available int `json:"available"`
}

// Info is returned by GET /api/v2/info.
type Info struct {
	Viewlog        string   `json:"viewlog"`
	Loadavg        float64  `json:"loadavg"`
	FreeSpace      string   `json:"free_space"`
	DisplayPower   *string  `json:"display_power"` // nullable
	UpToDate       bool     `json:"up_to_date"`
	AnthiasVersion string   `json:"anthias_version"`
	DeviceModel    string   `json:"device_model"`
	Uptime         Uptime   `json:"uptime"`
	Memory         Memory   `json:"memory"`
	IPAddresses    []string `json:"ip_addresses"`
	MACAddress     string   `json:"mac_address"`
	HostUser       string   `json:"host_user"`
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

// Ptr returns a pointer to v. Useful for populating pointer request fields.
func Ptr[T any](v T) *T {
	return &v
}
