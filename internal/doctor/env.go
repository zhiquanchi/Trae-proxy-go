package doctor

import "os"

func DetectEnvProxy() map[string]string {
	keys := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"ALL_PROXY",
		"NO_PROXY",
		"http_proxy",
		"https_proxy",
		"all_proxy",
		"no_proxy",
	}

	out := map[string]string{}
	for _, key := range keys {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			upper := key
			if len(key) > 0 && key[0] >= 'a' && key[0] <= 'z' {
				switch key {
				case "http_proxy":
					upper = "HTTP_PROXY"
				case "https_proxy":
					upper = "HTTPS_PROXY"
				case "all_proxy":
					upper = "ALL_PROXY"
				case "no_proxy":
					upper = "NO_PROXY"
				}
			}
			if _, exists := out[upper]; !exists {
				out[upper] = v
			}
		}
	}
	return out
}
