package unifi

import (
	"context"
	"errors"
	"fmt"
)

// SysInfo represents detailed system information from the UniFi controller.
type SysInfo struct {
	Timezone        string   `json:"timezone"`
	Version         string   `json:"version"`
	PreviousVersion string   `json:"previous_version"`
	Build           string   `json:"build"`
	Name            string   `json:"name"`
	Hostname        string   `json:"hostname"`
	IPAddrs         []string `json:"ip_addrs"`
	Uptime          int64    `json:"uptime"`
	UBNTDeviceType  string   `json:"ubnt_device_type"`
	UDMVersion      string   `json:"udm_version"`

	/*

	   {
	       "Meta": {
	           "rc": "ok"
	       },
	       "data": [
	           {
	               "timezone": "America/New_York",
	               "autobackup": false,
	               "build": "atag_6.0.43_14348",
	               "version": "6.0.43",
	               "previous_version": "5.12.60",
	               "debug_mgmt": "warn",
	               "debug_system": "warn",
	               "debug_device": "warn",
	               "debug_sdn": "warn",
	               "data_retention_days": 90,
	               "data_retention_time_in_hours_for_5minutes_scale": 24,
	               "data_retention_time_in_hours_for_hourly_scale": 720,
	               "data_retention_time_in_hours_for_daily_scale": 2160,
	               "data_retention_time_in_hours_for_monthly_scale": 8760,
	               "data_retention_time_in_hours_for_others": 2160,
	               "update_available": false,
	               "update_downloaded": false,
	               "live_chat": "super-only",
	               "store_enabled": "super-only",
	               "hostname": "example-domain.ui.com",
	               "name": "Dream Machine",
	               "ip_addrs": [
	                   "1.2.3.4"
	               ],
	               "inform_port": 8080,
	               "https_port": 8443,
	               "override_inform_host": false,
	               "image_maps_use_google_engine": false,
	               "radius_disconnect_running": false,
	               "facebook_wifi_registered": false,
	               "sso_app_id": "",
	               "sso_app_sec": "",
	               "uptime": 2541796,
	               "anonymous_controller_id": "",
	               "ubnt_device_type": "UDMB",
	               "udm_version": "1.8.6.2969",
	               "unsupported_device_count": 0,
	               "unsupported_device_list": [],
	               "unifi_go_enabled": false
	           }
	       ]
	   }

	*/
}

// GetSystemInfo retrieves system info using the new API.
func (c *client) GetSystemInfo(ctx context.Context, id string) (*SysInfo, error) {
	var respBody struct {
		Meta Meta      `json:"Meta"`
		Data []SysInfo `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/stat/sysinfo", id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	return &respBody.Data[0], nil
}

// serverInfo represents basic server info from old API .
type serverInfo struct {
	Up            bool   `json:"up"`
	ServerVersion string `json:"server_version"`
	UUID          string `json:"uuid"`
}

// getOldSysInfo retrieves system information using the old API style.
func (c *client) getOldSysInfo(ctx context.Context) (*SysInfo, error) {
	var response struct {
		Data serverInfo `json:"Meta"`
	}

	err := c.Get(ctx, c.apiPaths.StatusPath, nil, &response)
	if err != nil {
		return nil, err
	}
	d := response.Data
	return &SysInfo{
		Version: d.ServerVersion,
	}, nil
}

// GetSystemInformation retrieves system information, trying the new API first and falling back to the old API if necessary.
func (c *client) GetSystemInformation() (*SysInfo, error) {
	c.Trace("Reading system information")
	ctx, cancel := c.newRequestContext()
	defer cancel()

	var resultingError error
	info, err := c.GetSystemInfo(ctx, "default")
	if err != nil {
		resultingError = err
	} else if info == nil || info.Version == "" {
		resultingError = errors.New("new API returned empty server info")
	}

	if resultingError != nil {
		info, err = c.getOldSysInfo(ctx)
		if err != nil {
			resultingError = errors.Join(resultingError, err)
		} else if info == nil || info.Version == "" {
			resultingError = errors.Join(resultingError, errors.New("old API returned empty server info"))
		} else {
			resultingError = nil
		}
	}

	if resultingError != nil {
		return nil, resultingError
	}
	return info, nil
}
