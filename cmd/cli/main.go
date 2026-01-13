package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"trae-proxy-go/internal/cert"
	"trae-proxy-go/internal/config"
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

