package tui

import (
	"fmt"
	"net/url"
	"strings"
	"trae-proxy-go/internal/autoconfig"
	"trae-proxy-go/internal/cert"
	"trae-proxy-go/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type actionType int

const (
	actionNone actionType = iota
	actionAdd
	actionEdit
	actionDelete
	actionActivate
	actionDomain
	actionCert
)

type action struct {
	action actionType
	index  int
}

// 列表视图
type listViewModel struct {
	config  *models.Config
	selected int
	action  *action
}

func newListView(cfg *models.Config) listViewModel {
	return listViewModel{
		config:  cfg,
		selected: 0,
	}
}

func (m listViewModel) update(msg tea.Msg, cfg *models.Config, mode vimMode) (listViewModel, tea.Cmd) {
	m.config = cfg
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only handle navigation in normal mode
		if mode != vimModeNormal {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.config.APIs)-1 {
				m.selected++
			}
		case "g":
			// Go to top (simplified from vim's gg for single-key convenience)
			m.selected = 0
		case "G":
			// Go to bottom
			if len(m.config.APIs) > 0 {
				m.selected = len(m.config.APIs) - 1
			}
		case "a":
			m.action = &action{action: actionAdd}
		case "e":
			if len(m.config.APIs) > 0 {
				m.action = &action{action: actionEdit, index: m.selected}
			}
		case "d":
			if len(m.config.APIs) > 0 && len(m.config.APIs) > 1 {
				m.action = &action{action: actionDelete, index: m.selected}
			}
		case " ", "x":
			if len(m.config.APIs) > 0 {
				m.action = &action{action: actionActivate, index: m.selected}
			}
		case "D":
			m.action = &action{action: actionDomain}
		case "C":
			m.action = &action{action: actionCert}
		}
	}
	return m, nil
}

func (m listViewModel) view(mode vimMode) string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Trae-Proxy 配置管理"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("代理域名: %s\n\n", m.config.Domain))
	s.WriteString(borderStyle.Render("API 配置列表:\n\n"))

	if len(m.config.APIs) == 0 {
		s.WriteString("暂无API配置\n")
	} else {
		for i, api := range m.config.APIs {
			prefix := "  "
			style := itemStyle
			if i == m.selected {
				prefix = "> "
				style = selectedItemStyle
			}

			status := inactiveStyle.Render("✗ 未激活")
			if api.Active {
				status = activeStyle.Render("✓ 激活")
			}

			streamMode := "None"
			if api.StreamMode != "" {
				streamMode = api.StreamMode
			}

			s.WriteString(style.Render(fmt.Sprintf("%s%d. %s [%s]", prefix, i+1, api.Name, status)))
			s.WriteString("\n")
			s.WriteString(style.Render(fmt.Sprintf("   后端API: %s", api.Endpoint)))
			s.WriteString("\n")
			s.WriteString(style.Render(fmt.Sprintf("   自定义模型ID: %s", api.CustomModelID)))
			s.WriteString("\n")
			s.WriteString(style.Render(fmt.Sprintf("   目标模型ID: %s", api.TargetModelID)))
			s.WriteString("\n")
			s.WriteString(style.Render(fmt.Sprintf("   流模式: %s", streamMode)))
			s.WriteString("\n\n")
		}
	}

	// Update help text to show vim keybindings
	s.WriteString(helpStyle.Render("Vim: [j/k]上下 [g/G]首尾 [a]添加 [e]编辑 [d]删除 [x]激活 [D]域名 [C]证书 [:q]退出"))
	return s.String()
}

// 添加视图
type addViewModel struct {
	name       textinput.Model
	endpoint   textinput.Model
	customModel textinput.Model
	targetModel textinput.Model
	streamMode textinput.Model
	active     bool
	focused    int
	done       bool
	err        error
}

func newAddView() addViewModel {
	name := textinput.New()
	name.Placeholder = "配置名称"
	name.Focus()

	endpoint := textinput.New()
	endpoint.Placeholder = "https://api.example.com"

	customModel := textinput.New()
	customModel.Placeholder = "custom-model-id"

	targetModel := textinput.New()
	targetModel.Placeholder = "target-model-id"

	streamMode := textinput.New()
	streamMode.Placeholder = "none/true/false"

	return addViewModel{
		name:        name,
		endpoint:    endpoint,
		customModel: customModel,
		targetModel: targetModel,
		streamMode:  streamMode,
		active:      false,
		focused:     0,
	}
}

func (m addViewModel) update(msg tea.Msg) (addViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" {
				if m.focused == 5 {
					// 保存
					m.done = true
					m.err = m.validate()
					return m, nil
				}
				if m.focused == 6 {
					// 取消
					m.done = true
					return m, nil
				}
				if m.focused == 4 {
					m.active = !m.active
					return m, nil
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > 6 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = 6
			}

			m.name.Blur()
			m.endpoint.Blur()
			m.customModel.Blur()
			m.targetModel.Blur()
			m.streamMode.Blur()

			switch m.focused {
			case 0:
				m.name.Focus()
			case 1:
				m.endpoint.Focus()
			case 2:
				m.customModel.Focus()
			case 3:
				m.targetModel.Focus()
			case 4:
				m.streamMode.Focus()
			}
			return m, nil
		}
	}

	switch m.focused {
	case 0:
		m.name, cmd = m.name.Update(msg)
	case 1:
		m.endpoint, cmd = m.endpoint.Update(msg)
	case 2:
		m.customModel, cmd = m.customModel.Update(msg)
	case 3:
		m.targetModel, cmd = m.targetModel.Update(msg)
	case 4:
		m.streamMode, cmd = m.streamMode.Update(msg)
	}

	return m, cmd
}

func (m addViewModel) validate() error {
	if m.name.Value() == "" {
		return fmt.Errorf("名称不能为空")
	}
	if m.endpoint.Value() == "" {
		return fmt.Errorf("后端API URL不能为空")
	}
	if _, err := url.Parse(m.endpoint.Value()); err != nil {
		return fmt.Errorf("无效的API URL格式: %v", err)
	}
	if m.customModel.Value() == "" {
		return fmt.Errorf("自定义模型ID不能为空")
	}
	if m.targetModel.Value() == "" {
		return fmt.Errorf("目标模型ID不能为空")
	}
	return nil
}

func (m addViewModel) getStreamMode() string {
	val := m.streamMode.Value()
	if val == "" || val == "none" {
		return ""
	}
	return val
}

func (m addViewModel) view() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("添加 API 配置"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("错误: %v\n\n", m.err)))
	}

	s.WriteString(borderStyle.Render(fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n\n%s激活: %s\n\n%s保存 [回车]%s取消 [q]",
		makeInputField("名称", m.name, m.focused == 0),
		makeInputField("后端API URL", m.endpoint, m.focused == 1),
		makeInputField("自定义模型ID", m.customModel, m.focused == 2),
		makeInputField("目标模型ID", m.targetModel, m.focused == 3),
		makeInputField("流模式 (none/true/false)", m.streamMode, m.focused == 4),
		getCheckbox("", m.active, m.focused == 5),
		helpStyle.Render("[空格]切换"),
		helpStyle.Render(""),
		helpStyle.Render(""),
	)))
	return s.String()
}

// 编辑视图
type editViewModel struct {
	index       int
	name        textinput.Model
	endpoint    textinput.Model
	customModel textinput.Model
	targetModel textinput.Model
	streamMode  textinput.Model
	active      bool
	setActive   bool
	focused     int
	done        bool
	err         error
}

func newEditView() editViewModel {
	return editViewModel{
		index: -1,
	}
}

func newEditViewFromAPI(api models.API) editViewModel {
	name := textinput.New()
	name.SetValue(api.Name)
	name.Focus()

	endpoint := textinput.New()
	endpoint.SetValue(api.Endpoint)

	customModel := textinput.New()
	customModel.SetValue(api.CustomModelID)

	targetModel := textinput.New()
	targetModel.SetValue(api.TargetModelID)

	streamMode := textinput.New()
	if api.StreamMode != "" {
		streamMode.SetValue(api.StreamMode)
	} else {
		streamMode.SetValue("none")
	}

	return editViewModel{
		index:       -1, // 需要在外部设置
		name:        name,
		endpoint:    endpoint,
		customModel: customModel,
		targetModel: targetModel,
		streamMode:  streamMode,
		active:      api.Active,
		setActive:   false,
		focused:     0,
	}
}

func (m editViewModel) update(msg tea.Msg) (editViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" {
				if m.focused == 5 {
					m.done = true
					m.err = m.validate()
					return m, nil
				}
				if m.focused == 6 {
					m.done = true
					return m, nil
				}
				if m.focused == 4 {
					m.active = !m.active
					m.setActive = true
					return m, nil
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > 6 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = 6
			}

			m.name.Blur()
			m.endpoint.Blur()
			m.customModel.Blur()
			m.targetModel.Blur()
			m.streamMode.Blur()

			switch m.focused {
			case 0:
				m.name.Focus()
			case 1:
				m.endpoint.Focus()
			case 2:
				m.customModel.Focus()
			case 3:
				m.targetModel.Focus()
			case 4:
				m.streamMode.Focus()
			}
			return m, nil
		}
	}

	switch m.focused {
	case 0:
		m.name, cmd = m.name.Update(msg)
	case 1:
		m.endpoint, cmd = m.endpoint.Update(msg)
	case 2:
		m.customModel, cmd = m.customModel.Update(msg)
	case 3:
		m.targetModel, cmd = m.targetModel.Update(msg)
	case 4:
		m.streamMode, cmd = m.streamMode.Update(msg)
	}

	return m, cmd
}

func (m editViewModel) validate() error {
	if m.endpoint.Value() != "" {
		if _, err := url.Parse(m.endpoint.Value()); err != nil {
			return fmt.Errorf("无效的API URL格式: %v", err)
		}
	}
	return nil
}

func (m editViewModel) getStreamMode() string {
	val := m.streamMode.Value()
	if val == "" || val == "none" {
		return ""
	}
	return val
}

func (m editViewModel) view() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("编辑 API 配置"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("错误: %v\n\n", m.err)))
	}

	s.WriteString(borderStyle.Render(fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n\n%s激活: %s\n\n%s保存 [回车]%s取消 [q]",
		makeInputField("名称 (留空保持原值)", m.name, m.focused == 0),
		makeInputField("后端API URL (留空保持原值)", m.endpoint, m.focused == 1),
		makeInputField("自定义模型ID (留空保持原值)", m.customModel, m.focused == 2),
		makeInputField("目标模型ID (留空保持原值)", m.targetModel, m.focused == 3),
		makeInputField("流模式 (none/true/false)", m.streamMode, m.focused == 4),
		getCheckbox("", m.active, m.focused == 5),
		helpStyle.Render("[空格]切换"),
		helpStyle.Render(""),
		helpStyle.Render(""),
	)))
	return s.String()
}

// 域名视图
type domainViewModel struct {
	domain textinput.Model
	done   bool
	err    error
}

func newDomainView(currentDomain string) domainViewModel {
	d := textinput.New()
	d.SetValue(currentDomain)
	d.Focus()
	d.Placeholder = "api.openai.com"
	return domainViewModel{
		domain: d,
	}
}

func (m domainViewModel) update(msg tea.Msg) (domainViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.domain.Value() != "" {
				m.done = true
			}
		}
	}
	m.domain, cmd = m.domain.Update(msg)
	return m, cmd
}

func (m domainViewModel) view() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("设置代理域名"))
	s.WriteString("\n\n")
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("错误: %v\n\n", m.err)))
	}
	s.WriteString(borderStyle.Render(fmt.Sprintf(
		"%s\n\n%s保存 [回车]%s取消 [q]",
		makeInputField("域名", m.domain, true),
		helpStyle.Render(""),
		helpStyle.Render(""),
	)))
	return s.String()
}

// 证书视图
type certViewModel struct {
	domain       string
	done         bool
	generating   bool
	success      bool
	err          error
	autoConfig   bool
	configuring  bool
	configDone   bool
	configErr    error
	stage        string // "prompt", "generating", "generated", "configuring", "complete"
}

func newCertView(domain string) certViewModel {
	return certViewModel{
		domain: domain,
		stage:  "prompt",
	}
}

func (m certViewModel) update(msg tea.Msg) (certViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "y":
			if m.stage == "prompt" {
				// 开始生成证书
				m.stage = "generating"
				m.generating = true
				go func() {
					err := cert.GenerateCertificates(m.domain, "ca")
					if err == nil {
						m.success = true
						m.stage = "generated"
					} else {
						m.err = err
						m.stage = "error"
					}
					m.generating = false
				}()
			} else if m.stage == "generated" {
				// 证书生成成功后，询问是否自动配置
				m.stage = "ask_config"
			} else if m.stage == "ask_config" {
				// 用户选择执行自动配置
				m.stage = "configuring"
				m.configuring = true
				go func() {
					err := autoconfig.AutoConfigure(m.domain, "ca", true, true)
					if err == nil {
						m.configDone = true
						m.stage = "complete"
					} else {
						m.configErr = err
						m.stage = "config_error"
					}
					m.configuring = false
				}()
			} else {
				// 其他阶段按回车返回
				m.done = true
			}
		case "n":
			if m.stage == "prompt" {
				// 取消生成证书
				m.done = true
			} else if m.stage == "ask_config" {
				// 跳过自动配置
				m.stage = "skip_config"
			}
		case "q":
			m.done = true
		}
	}
	return m, nil
}

func (m certViewModel) view() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("生成 SSL 证书"))
	s.WriteString("\n\n")

	switch m.stage {
	case "prompt":
		s.WriteString(borderStyle.Render(fmt.Sprintf(
			"将为域名 %s 生成 SSL 证书\n\n%s生成 [y/回车]%s取消 [n/q]",
			m.domain,
			helpStyle.Render(""),
			helpStyle.Render(""),
		)))
	
	case "generating":
		s.WriteString(borderStyle.Render(fmt.Sprintf(
			"正在为域名 %s 生成证书...\n请稍候",
			m.domain,
		)))
	
	case "generated":
		s.WriteString(successStyle.Render("证书生成成功!\n\n"))
		s.WriteString(borderStyle.Render("证书文件已保存到 ca/ 目录"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("[回车]继续配置"))
		// 自动进入下一阶段
		go func() {
			// 使用一个小延迟让用户看到成功消息
			m.stage = "ask_config"
		}()
	
	case "ask_config":
		s.WriteString(successStyle.Render("证书生成成功!\n\n"))
		s.WriteString(borderStyle.Render(
			"是否自动配置系统？\n" +
			"（将安装CA证书到系统信任存储并更新hosts文件）\n\n" +
			"注意：需要管理员/root权限\n\n" +
			helpStyle.Render("配置 [y/回车]  ") +
			helpStyle.Render("跳过 [n]  ") +
			helpStyle.Render("退出 [q]"),
		))
	
	case "configuring":
		s.WriteString(borderStyle.Render(
			"正在配置系统...\n" +
			"- 安装CA证书到系统信任存储\n" +
			"- 更新hosts文件\n\n" +
			"请稍候...",
		))
	
	case "complete":
		s.WriteString(successStyle.Render("配置完成!\n\n"))
		s.WriteString(borderStyle.Render(
			"已完成以下配置：\n" +
			"- CA证书已安装到系统信任存储\n" +
			fmt.Sprintf("- hosts文件已更新（%s -> 127.0.0.1）\n\n", m.domain) +
			"您现在可以启动代理服务器了",
		))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("[回车/q]返回"))
	
	case "skip_config":
		s.WriteString(successStyle.Render("证书生成成功!\n\n"))
		s.WriteString(borderStyle.Render("已跳过自动配置"))
		s.WriteString("\n\n")
		s.WriteString("手动配置说明：\n")
		s.WriteString(helpStyle.Render(autoconfig.GetInstructions(m.domain, "ca")))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("[回车/q]返回"))
	
	case "config_error":
		s.WriteString(errorStyle.Render(fmt.Sprintf("自动配置失败: %v\n\n", m.configErr)))
		s.WriteString("手动配置说明：\n")
		s.WriteString(helpStyle.Render(autoconfig.GetInstructions(m.domain, "ca")))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("[回车/q]返回"))
	
	case "error":
		s.WriteString(errorStyle.Render(fmt.Sprintf("错误: %v\n\n", m.err)))
		s.WriteString(helpStyle.Render("[回车/q]返回"))
	}

	return s.String()
}

// 辅助函数
func makeInputField(label string, input textinput.Model, focused bool) string {
	style := lipgloss.NewStyle().Width(50)
	if focused {
		style = style.Foreground(lipgloss.Color("170"))
	}
	return fmt.Sprintf("%s:\n%s", label, style.Render(input.View()))
}

func getCheckbox(label string, checked bool, focused bool) string {
	checkbox := "[ ]"
	if checked {
		checkbox = "[x]"
	}
	style := lipgloss.NewStyle()
	if focused {
		style = style.Foreground(lipgloss.Color("170"))
	}
	return style.Render(fmt.Sprintf("%s%s", checkbox, label))
}

