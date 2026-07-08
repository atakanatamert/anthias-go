package anthias

import (
	"context"
	"net/http"
	"testing"
)

func TestInfoAndIntegrations(t *testing.T) {
	displayPower := "on"
	deviceID := "balena-device"
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		switch calls {
		case 0:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/info")
			writeJSON(t, w, http.StatusOK, Info{
				Viewlog:        "ok",
				Loadavg:        0.42,
				FreeSpace:      "10G",
				DisplayPower:   &displayPower,
				UpToDate:       true,
				AnthiasVersion: "v0.19.0",
				DeviceModel:    "raspberry-pi",
				Uptime:         Uptime{Days: 1, Hours: 2.5},
				Memory:         Memory{Total: 100, Used: 50, Free: 50, Shared: 1, Buff: 2, Available: 80},
				IPAddresses:    []string{"192.0.2.10"},
				MACAddress:     "00:11:22:33:44:55",
				HostUser:       "pi",
			})
		case 1:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/integrations")
			writeJSON(t, w, http.StatusOK, Integrations{
				IsBalena:       true,
				BalenaDeviceID: &deviceID,
			})
		default:
			t.Fatalf("unexpected call %d", calls)
		}
		calls++
	}))

	info, err := c.GetInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.AnthiasVersion != "v0.19.0" || info.DisplayPower == nil || *info.DisplayPower != "on" {
		t.Fatalf("info = %#v", info)
	}
	integrations, err := c.GetIntegrations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !integrations.IsBalena || integrations.BalenaDeviceID == nil || *integrations.BalenaDeviceID != "balena-device" {
		t.Fatalf("integrations = %#v", integrations)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
