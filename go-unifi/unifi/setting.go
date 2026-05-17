package unifi

import (
	"context"
	"encoding/json"
	"fmt"
)

type Setting struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`
	Key    string `json:"key"`
}

func (s *Setting) newFields() (interface{}, error) {
	switch s.Key {
	case SettingAutoSpeedtestKey:
		return &SettingAutoSpeedtest{}, nil
	case SettingBaresipKey:
		return &SettingBaresip{}, nil
	case SettingBroadcastKey:
		return &SettingBroadcast{}, nil
	case SettingConnectivityKey:
		return &SettingConnectivity{}, nil
	case SettingCountryKey:
		return &SettingCountry{}, nil
	case SettingDashboardKey:
		return &SettingDashboard{}, nil
	case SettingDohKey:
		return &SettingDoh{}, nil
	case SettingDpiKey:
		return &SettingDpi{}, nil
	case SettingElementAdoptKey:
		return &SettingElementAdopt{}, nil
	case SettingEtherLightingKey:
		return &SettingEtherLighting{}, nil
	case SettingEvaluationScoreKey:
		return &SettingEvaluationScore{}, nil
	case SettingGlobalApKey:
		return &SettingGlobalAp{}, nil
	case SettingGlobalNatKey:
		return &SettingGlobalNat{}, nil
	case SettingGlobalSwitchKey:
		return &SettingGlobalSwitch{}, nil
	case SettingGuestAccessKey:
		return &SettingGuestAccess{}, nil
	case SettingIpsKey:
		return &SettingIps{}, nil
	case SettingLcmKey:
		return &SettingLcm{}, nil
	case SettingLocaleKey:
		return &SettingLocale{}, nil
	case SettingMagicSiteToSiteVpnKey:
		return &SettingMagicSiteToSiteVpn{}, nil
	case SettingMgmtKey:
		return &SettingMgmt{}, nil
	case SettingNetflowKey:
		return &SettingNetflow{}, nil
	case SettingNetworkOptimizationKey:
		return &SettingNetworkOptimization{}, nil
	case SettingNtpKey:
		return &SettingNtp{}, nil
	case SettingPortaKey:
		return &SettingPorta{}, nil
	case SettingRadioAiKey:
		return &SettingRadioAi{}, nil
	case SettingRadiusKey:
		return &SettingRadius{}, nil
	case SettingRsyslogdKey:
		return &SettingRsyslogd{}, nil
	case SettingSnmpKey:
		return &SettingSnmp{}, nil
	case SettingSslInspectionKey:
		return &SettingSslInspection{}, nil
	case SettingSuperCloudaccessKey:
		return &SettingSuperCloudaccess{}, nil
	case SettingSuperEventsKey:
		return &SettingSuperEvents{}, nil
	case SettingSuperFwupdateKey:
		return &SettingSuperFwupdate{}, nil
	case SettingSuperIdentityKey:
		return &SettingSuperIdentity{}, nil
	case SettingSuperMailKey:
		return &SettingSuperMail{}, nil
	case SettingSuperMgmtKey:
		return &SettingSuperMgmt{}, nil
	case SettingSuperSdnKey:
		return &SettingSuperSdn{}, nil
	case SettingSuperSmtpKey:
		return &SettingSuperSmtp{}, nil
	case SettingTeleportKey:
		return &SettingTeleport{}, nil
	case SettingUsgKey:
		return &SettingUsg{}, nil
	case SettingUswKey:
		return &SettingUsw{}, nil
	}

	return nil, fmt.Errorf("unexpected key %q", s.Key)
}

func (c *client) SetSetting(ctx context.Context, site, key string, reqBody interface{}) (interface{}, error) {
	var respBody struct {
		Meta Meta              `json:"meta"`
		Data []json.RawMessage `json:"data"`
	}
	err := c.Put(ctx, fmt.Sprintf("s/%s/set/setting/%s", site, key), reqBody, &respBody)
	if err != nil {
		return nil, err
	}
	var raw json.RawMessage
	var setting *Setting
	for _, d := range respBody.Data {
		err = json.Unmarshal(d, &setting)
		if err != nil {
			return nil, err
		}
		if setting.Key == key {
			raw = d
			break
		}
	}
	if setting == nil || setting.Key != key {
		return nil, ErrNotFound
	}
	fields, err := setting.newFields()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, &fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func (c *client) GetSetting(ctx context.Context, site, key string) (*Setting, interface{}, error) {
	var respBody struct {
		Meta Meta              `json:"Meta"`
		Data []json.RawMessage `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/get/setting", site), nil, &respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get setting %s: %w", key, err)
	}

	var raw json.RawMessage
	var setting *Setting
	for _, d := range respBody.Data {
		err = json.Unmarshal(d, &setting)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to decode get setting %s: %w", key, err)
		}
		if setting.Key == key {
			raw = d
			break
		}
	}
	if setting == nil || setting.Key != key {
		return nil, nil, ErrNotFound
	}

	fields, err := setting.newFields()
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(raw, &fields)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode get setting fields %s: %w", key, err)
	}

	return setting, fields, nil
}
