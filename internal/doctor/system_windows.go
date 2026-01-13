//go:build windows

package doctor

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type SystemProxy struct {
	Enabled  bool
	Server   string
	Override string
	Source   string
}

func DetectSystemProxy() *SystemProxy {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return &SystemProxy{Enabled: false, Source: fmt.Sprintf("registry open error: %v", err)}
	}
	defer key.Close()

	enabledValue, _, err := key.GetIntegerValue("ProxyEnable")
	enabled := err == nil && enabledValue != 0

	server, _, _ := key.GetStringValue("ProxyServer")
	override, _, _ := key.GetStringValue("ProxyOverride")

	return &SystemProxy{
		Enabled:  enabled,
		Server:   server,
		Override: override,
		Source:   "windows registry (WinINET)",
	}
}
