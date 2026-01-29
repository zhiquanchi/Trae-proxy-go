package doctor

import (
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
)

type Report struct {
	Domain     string
	Port       int
	GOOS       string
	GOARCH     string
	Env        map[string]string
	System     *SystemProxy
	PortStatus string

	SuggestedNoProxy string
	Notes            []string
}

func GenerateReport(domain string, port int) Report {
	env := DetectEnvProxy()
	system := DetectSystemProxy()

	portStatus := checkPort(port)

	suggestedNoProxy := buildSuggestedNoProxy(domain, env)

	notes := buildNotes(domain, env, system, portStatus)

	return Report{
		Domain:           domain,
		Port:             port,
		GOOS:             runtime.GOOS,
		GOARCH:           runtime.GOARCH,
		Env:              env,
		System:           system,
		PortStatus:       portStatus,
		SuggestedNoProxy: suggestedNoProxy,
		Notes:            notes,
	}
}

func buildSuggestedNoProxy(domain string, env map[string]string) string {
	candidates := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		strings.TrimSpace(domain),
	}

	existing := envValue(env, "NO_PROXY")
	if existing != "" {
		for _, token := range splitProxyList(existing) {
			if token != "" && !containsToken(candidates, token) {
				candidates = append(candidates, token)
			}
		}
	}

	normalized := make([]string, 0, len(candidates))
	seen := map[string]bool{}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		key := strings.ToLower(c)
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, c)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		ai, aj := strings.ToLower(normalized[i]), strings.ToLower(normalized[j])
		return ai < aj
	})

	return strings.Join(normalized, ",")
}

func buildNotes(domain string, env map[string]string, system *SystemProxy, portStatus string) []string {
	var notes []string

	hasEnvProxy := envValue(env, "HTTP_PROXY") != "" || envValue(env, "HTTPS_PROXY") != "" || envValue(env, "ALL_PROXY") != ""
	if hasEnvProxy {
		if !listContains(envValue(env, "NO_PROXY"), domain) && !listContains(envValue(env, "NO_PROXY"), "127.0.0.1") {
			notes = append(notes, "检测到环境变量代理，建议配置 NO_PROXY 以避免请求被其他代理接管（尤其是本机 hosts 指向 127.0.0.1 的场景）")
		}
	}

	if system != nil && system.Enabled {
		notes = append(notes, "检测到系统代理开启，若客户端走系统代理，hosts 可能不生效；建议在代理软件中将域名设为直连/绕过")
	}

	if strings.Contains(strings.ToLower(portStatus), "in use") {
		notes = append(notes, "监听端口可能被占用：可以关闭占用程序，或修改 server.port 并用端口转发/客户端可配置 base_url 的方式接入")
	}

	if len(notes) == 0 {
		notes = append(notes, "未发现明显的代理冲突信号")
	}

	return notes
}

func checkPort(port int) string {
	if port <= 0 || port > 65535 {
		return "invalid port"
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		msg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(msg, "address already in use"):
			return "in use"
		case strings.Contains(msg, "permission denied") || strings.Contains(msg, "access is denied"):
			return "permission denied"
		default:
			return "error: " + err.Error()
		}
	}
	_ = ln.Close()
	return "available"
}

func envValue(env map[string]string, key string) string {
	if env == nil {
		return ""
	}
	if v, ok := env[key]; ok {
		return v
	}
	return ""
}

func listContains(list string, token string) bool {
	if list == "" || token == "" {
		return false
	}
	for _, item := range splitProxyList(list) {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(token)) {
			return true
		}
	}
	return false
}

func splitProxyList(list string) []string {
	list = strings.ReplaceAll(list, ";", ",")
	parts := strings.Split(list, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func containsToken(list []string, token string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(token)) {
			return true
		}
	}
	return false
}
