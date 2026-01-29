package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"trae-proxy-go/internal/autoconfig"
	"trae-proxy-go/internal/cert"
	"trae-proxy-go/internal/config"
	"trae-proxy-go/internal/doctor"
	"trae-proxy-go/internal/tui"
	"trae-proxy-go/pkg/models"
)

const configFile = "config.yaml"

func main() {
	if len(os.Args) < 2 {
		// 无参数时启动TUI界面
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI运行错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	command := os.Args[1]
	switch command {
	case "list":
		handleList()
	case "add":
		handleAdd()
	case "remove":
		handleRemove()
	case "update":
		handleUpdate()
	case "activate":
		handleActivate()
	case "domain":
		handleDomain()
	case "cert":
		handleCert()
	case "start":
		handleStart()
	case "doctor":
		handleDoctor()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("用法: trae-proxy-cli [command] [options]")
	fmt.Println("\n无参数时启动TUI界面")
	fmt.Println("\n命令:")
	fmt.Println("  list                   列出所有API配置")
	fmt.Println("  add                    添加新API配置")
	fmt.Println("  remove                 删除API配置")
	fmt.Println("  update                 更新API配置")
	fmt.Println("  activate               激活API配置")
	fmt.Println("  domain                 更新代理域名")
	fmt.Println("  cert                   生成证书")
	fmt.Println("  start                  启动代理服务器")
	fmt.Println("  doctor                 检测代理/端口冲突并给出建议")
}

func handleList() {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n当前API配置列表:")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("代理域名: %s\n", cfg.Domain)
	fmt.Println("--------------------------------------------------------------------------------")

	for i, api := range cfg.APIs {
		status := "✗ 未激活"
		if api.Active {
			status = "✓ 激活"
		}
		streamMode := "None"
		if api.StreamMode != "" {
			streamMode = api.StreamMode
		}
		fmt.Printf("%d. %s [%s]\n", i+1, api.Name, status)
		fmt.Printf("   后端API: %s\n", api.Endpoint)
		fmt.Printf("   自定义模型ID: %s\n", api.CustomModelID)
		fmt.Printf("   目标模型ID: %s\n", api.TargetModelID)
		fmt.Printf("   流模式: %s\n", streamMode)
		fmt.Println("--------------------------------------------------------------------------------")
	}
}

func handleAdd() {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "配置名称（必需）")
	endpoint := fs.String("endpoint", "", "后端API URL（必需）")
	customModel := fs.String("custom-model", "", "自定义模型ID（必需）")
	targetModel := fs.String("target-model", "", "目标模型ID（必需）")
	streamMode := fs.String("stream-mode", "none", "流模式 (true/false/none)")
	active := fs.Bool("active", false, "激活此API配置")

	fs.Parse(os.Args[2:])

	if *name == "" || *endpoint == "" || *customModel == "" || *targetModel == "" {
		fmt.Fprintf(os.Stderr, "错误: name, endpoint, custom-model, target-model 都是必需的\n")
		os.Exit(1)
	}

	// 验证URL格式
	if _, err := url.Parse(*endpoint); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的API URL格式: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	// 处理流模式
	streamModeValue := ""
	if *streamMode != "none" {
		streamModeValue = *streamMode
	}

	newAPI := models.API{
		Name:          *name,
		Endpoint:      *endpoint,
		CustomModelID: *customModel,
		TargetModelID: *targetModel,
		StreamMode:    streamModeValue,
		Active:        *active,
	}

	if *active {
		// 如果激活新API，禁用其他所有API
		for i := range cfg.APIs {
			cfg.APIs[i].Active = false
		}
	}

	cfg.APIs = append(cfg.APIs, newAPI)

	if err := config.SaveConfig(cfg, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已添加新API配置: %s\n", *name)
}

func handleRemove() {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	index := fs.Int("index", -1, "API索引（从0开始，必需）")

	fs.Parse(os.Args[2:])

	if *index < 0 {
		fmt.Fprintf(os.Stderr, "错误: index 是必需的且必须 >= 0\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if *index >= len(cfg.APIs) {
		fmt.Fprintf(os.Stderr, "错误: 无效的API索引: %d\n", *index)
		os.Exit(1)
	}

	if len(cfg.APIs) <= 1 {
		fmt.Fprintf(os.Stderr, "错误: 至少需要保留一个API配置\n")
		os.Exit(1)
	}

	removed := cfg.APIs[*index]
	cfg.APIs = append(cfg.APIs[:*index], cfg.APIs[*index+1:]...)

	if err := config.SaveConfig(cfg, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已删除API配置: %s\n", removed.Name)
}

func handleUpdate() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	index := fs.Int("index", -1, "API索引（从0开始，必需）")
	name := fs.String("name", "", "配置名称")
	endpoint := fs.String("endpoint", "", "后端API URL")
	customModel := fs.String("custom-model", "", "自定义模型ID")
	targetModel := fs.String("target-model", "", "目标模型ID")
	streamMode := fs.String("stream-mode", "", "流模式 (true/false/none)")
	active := fs.Bool("active", false, "激活此API配置")
	hasActive := fs.Bool("set-active", false, "设置激活状态（使用-set-active=true/false）")

	fs.Parse(os.Args[2:])

	if *index < 0 {
		fmt.Fprintf(os.Stderr, "错误: index 是必需的且必须 >= 0\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if *index >= len(cfg.APIs) {
		fmt.Fprintf(os.Stderr, "错误: 无效的API索引: %d\n", *index)
		os.Exit(1)
	}

	api := &cfg.APIs[*index]

	if *name != "" {
		api.Name = *name
	}
	if *endpoint != "" {
		if _, err := url.Parse(*endpoint); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的API URL格式: %v\n", err)
			os.Exit(1)
		}
		api.Endpoint = *endpoint
	}
	if *customModel != "" {
		api.CustomModelID = *customModel
	}
	if *targetModel != "" {
		api.TargetModelID = *targetModel
	}
	if *streamMode != "" {
		if *streamMode == "none" {
			api.StreamMode = ""
		} else {
			api.StreamMode = *streamMode
		}
	}
	if *hasActive {
		api.Active = *active
		if *active {
			// 如果激活当前API，禁用其他API
			for i := range cfg.APIs {
				if i != *index {
					cfg.APIs[i].Active = false
				}
			}
		}
	}

	if err := config.SaveConfig(cfg, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已更新API配置: %s\n", api.Name)
}

func handleActivate() {
	fs := flag.NewFlagSet("activate", flag.ExitOnError)
	index := fs.Int("index", -1, "API索引（从0开始，必需）")

	fs.Parse(os.Args[2:])

	if *index < 0 {
		fmt.Fprintf(os.Stderr, "错误: index 是必需的且必须 >= 0\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if *index >= len(cfg.APIs) {
		fmt.Fprintf(os.Stderr, "错误: 无效的API索引: %d\n", *index)
		os.Exit(1)
	}

	// 禁用所有API
	for i := range cfg.APIs {
		cfg.APIs[i].Active = false
	}

	// 激活指定API
	cfg.APIs[*index].Active = true

	if err := config.SaveConfig(cfg, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已激活API配置: %s\n", cfg.APIs[*index].Name)
}

func handleDomain() {
	fs := flag.NewFlagSet("domain", flag.ExitOnError)
	name := fs.String("name", "", "域名（必需）")

	fs.Parse(os.Args[2:])

	if *name == "" {
		fmt.Fprintf(os.Stderr, "错误: name 是必需的\n")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	cfg.Domain = *name

	if err := config.SaveConfig(cfg, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已更新代理域名: %s\n", *name)
}

func handleCert() {
	fs := flag.NewFlagSet("cert", flag.ExitOnError)
	domain := fs.String("domain", "", "域名（默认使用配置中的域名）")
	autoConfig := fs.Bool("auto-config", false, "自动配置系统（安装CA证书和更新hosts文件，需要管理员权限）")
	installCA := fs.Bool("install-ca", false, "仅安装CA证书到系统信任存储")
	updateHosts := fs.Bool("update-hosts", false, "仅更新hosts文件")

	fs.Parse(os.Args[2:])

	var targetDomain string
	if *domain != "" {
		targetDomain = *domain
	} else {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}
		targetDomain = cfg.Domain
	}

	fmt.Printf("为域名 %s 生成证书...\n", targetDomain)

	if err := cert.GenerateCertificates(targetDomain, "ca"); err != nil {
		fmt.Fprintf(os.Stderr, "证书生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("证书生成成功")

	// 处理自动配置
	if *autoConfig || *installCA || *updateHosts {
		// 确定要执行的操作
		shouldInstallCA := *autoConfig || *installCA
		shouldUpdateHosts := *autoConfig || *updateHosts

		if autoconfig.NeedsElevatedPrivileges() {
			fmt.Println("\n注意: 自动配置需要管理员/root权限")
		}

		fmt.Println("\n开始自动配置...")
		if err := autoconfig.AutoConfigure(targetDomain, "ca", shouldInstallCA, shouldUpdateHosts); err != nil {
			fmt.Fprintf(os.Stderr, "\n自动配置失败: %v\n", err)
			fmt.Println("\n请尝试手动配置：")
			fmt.Println(autoconfig.GetInstructions(targetDomain, "ca"))
			os.Exit(1)
		}

		fmt.Println("自动配置成功!")
		if shouldInstallCA {
			fmt.Println("- CA证书已安装到系统信任存储")
		}
		if shouldUpdateHosts {
			fmt.Printf("- hosts文件已更新（%s -> 127.0.0.1）\n", targetDomain)
		}
	} else {
		// 不自动配置时，显示手动配置说明
		fmt.Println("\n如需自动配置，请使用 --auto-config 参数（需要管理员权限）")
		fmt.Println("或查看以下手动配置说明：")
		fmt.Println(autoconfig.GetInstructions(targetDomain, "ca"))
	}
}

func handleStart() {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	debug := fs.Bool("debug", false, "启用调试模式")

	fs.Parse(os.Args[2:])

	// 启动服务器实际上是通过运行proxy程序来实现
	// 这里我们只是打印提示信息，实际的启动应该由proxy程序完成
	fmt.Println("请使用以下命令启动代理服务器:")
	fmt.Printf("  trae-proxy --config=%s", configFile)
	if *debug {
		fmt.Print(" --debug")
	}
	fmt.Println()
	fmt.Println("\n或者直接运行:")
	fmt.Println("  go run cmd/proxy/main.go")
}

func handleDoctor() {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	configPath := fs.String("config", configFile, "配置文件路径")
	domain := fs.String("domain", "", "代理域名（覆盖配置文件）")
	port := fs.Int("port", 0, "监听端口（覆盖配置文件）")

	fs.Parse(os.Args[2:])

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	targetDomain := cfg.Domain
	if *domain != "" {
		targetDomain = *domain
	}
	targetPort := cfg.Server.Port
	if *port != 0 {
		targetPort = *port
	}

	report := doctor.GenerateReport(targetDomain, targetPort)

	fmt.Println("Trae-Proxy Doctor")
	fmt.Printf("OS: %s/%s\n", report.GOOS, report.GOARCH)
	fmt.Printf("Domain: %s\n", report.Domain)
	fmt.Printf("Port: %d (%s)\n", report.Port, report.PortStatus)
	fmt.Println()

	if len(report.Env) == 0 {
		fmt.Println("Env proxy: (none)")
	} else {
		fmt.Println("Env proxy:")
		for _, k := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY"} {
			if v, ok := report.Env[k]; ok && v != "" {
				fmt.Printf("  %s=%s\n", k, v)
			}
		}
	}
	fmt.Println()

	if report.System == nil {
		fmt.Println("System proxy: (unsupported on this OS)")
	} else {
		fmt.Printf("System proxy: enabled=%v\n", report.System.Enabled)
		if report.System.Server != "" {
			fmt.Printf("  server=%s\n", report.System.Server)
		}
		if report.System.Override != "" {
			fmt.Printf("  override=%s\n", report.System.Override)
		}
		fmt.Printf("  source=%s\n", report.System.Source)
	}
	fmt.Println()

	fmt.Println("Suggested settings:")
	fmt.Printf("  NO_PROXY=%s\n", report.SuggestedNoProxy)
	fmt.Println("  PowerShell: $env:NO_PROXY=\"<value above>\"")
	fmt.Println("  CMD: set NO_PROXY=<value above>")
	fmt.Println()

	fmt.Println("Notes:")
	for _, note := range report.Notes {
		fmt.Printf("  - %s\n", note)
	}
}
