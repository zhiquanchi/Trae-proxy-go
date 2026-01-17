package tui

import (
	"fmt"
	"trae-proxy-go/internal/config"
	"trae-proxy-go/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
)

const configFile = "config.yaml"

type vimMode int

const (
	vimModeNormal vimMode = iota
	vimModeCommand
)

type model struct {
	view       viewType
	config     *models.Config
	selected   int
	err        error
	listView   listViewModel
	addView    addViewModel
	editView   editViewModel
	domainView domainViewModel
	certView   certViewModel
	vimMode    vimMode
	commandBuf string
}

type viewType int

const (
	viewList viewType = iota
	viewAdd
	viewEdit
	viewDomain
	viewCert
)

type errMsg struct {
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

func InitialModel() model {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		cfg = config.GetDefaultConfig()
		if err := config.SaveConfig(cfg, configFile); err != nil {
			return model{err: err}
		}
	}

	return model{
		view:       viewList,
		config:     cfg,
		selected:   0,
		listView:   newListView(cfg),
		addView:    newAddView(),
		editView:   newEditView(),
		domainView: newDomainView(cfg.Domain),
		certView:   newCertView(cfg.Domain),
		vimMode:    vimModeNormal,
		commandBuf: "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle vim command mode
		if m.vimMode == vimModeCommand {
			return m.handleCommandMode(msg)
		}

		// Global keybindings in normal mode
		if m.vimMode == vimModeNormal && m.view == viewList {
			switch msg.String() {
			case ":":
				m.vimMode = vimModeCommand
				m.commandBuf = ""
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.view == viewList && m.vimMode == vimModeNormal {
				return m, tea.Quit
			}
			if m.view != viewList {
				// 从其他视图返回列表视图
				m.view = viewList
				m.vimMode = vimModeNormal
				// 重新加载配置
				cfg, err := config.LoadConfig(configFile)
				if err == nil {
					m.config = cfg
					m.listView = newListView(cfg)
				}
			}
			return m, nil
		case "esc":
			// Esc always returns to normal mode
			m.vimMode = vimModeNormal
			if m.view != viewList {
				m.view = viewList
				cfg, err := config.LoadConfig(configFile)
				if err == nil {
					m.config = cfg
					m.listView = newListView(cfg)
				}
			}
			return m, nil
		}

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	switch m.view {
	case viewList:
		m.listView, cmd = m.listView.update(msg, m.config, m.vimMode)
		if m.listView.action != nil {
			switch m.listView.action.action {
			case actionAdd:
				m.view = viewAdd
				m.addView = newAddView()
				m.listView.action = nil
			case actionEdit:
				m.view = viewEdit
				m.editView = newEditViewFromAPI(m.config.APIs[m.listView.action.index])
				m.editView.index = m.listView.action.index
				m.listView.action = nil
			case actionDelete:
				if len(m.config.APIs) > 1 {
					m.config.APIs = append(m.config.APIs[:m.listView.action.index], m.config.APIs[m.listView.action.index+1:]...)
					if err := config.SaveConfig(m.config, configFile); err != nil {
						m.err = err
					} else {
						m.listView = newListView(m.config)
					}
				}
				m.listView.action = nil
			case actionActivate:
				for i := range m.config.APIs {
					m.config.APIs[i].Active = false
				}
				m.config.APIs[m.listView.action.index].Active = true
				if err := config.SaveConfig(m.config, configFile); err != nil {
					m.err = err
				} else {
					m.listView = newListView(m.config)
				}
				m.listView.action = nil
			case actionDomain:
				m.view = viewDomain
				m.domainView = newDomainView(m.config.Domain)
				m.listView.action = nil
			case actionCert:
				m.view = viewCert
				m.certView = newCertView(m.config.Domain)
				m.listView.action = nil
			}
		}
	case viewAdd:
		m.addView, cmd = m.addView.update(msg)
		if m.addView.done {
			if m.addView.err == nil {
				// 保存新API
				newAPI := models.API{
					Name:          m.addView.name.Value(),
					Endpoint:      m.addView.endpoint.Value(),
					CustomModelID: m.addView.customModel.Value(),
					TargetModelID: m.addView.targetModel.Value(),
					StreamMode:    m.addView.getStreamMode(),
					Active:        m.addView.active,
				}
				if newAPI.Active {
					for i := range m.config.APIs {
						m.config.APIs[i].Active = false
					}
				}
				m.config.APIs = append(m.config.APIs, newAPI)
				if err := config.SaveConfig(m.config, configFile); err != nil {
					m.err = err
				} else {
					m.config, _ = config.LoadConfig(configFile)
					m.listView = newListView(m.config)
					m.view = viewList
				}
			}
			m.addView.done = false
		}
	case viewEdit:
		m.editView, cmd = m.editView.update(msg)
		if m.editView.done {
			if m.editView.err == nil && m.editView.index >= 0 && m.editView.index < len(m.config.APIs) {
				// 更新API
				api := &m.config.APIs[m.editView.index]
				if m.editView.name.Value() != "" {
					api.Name = m.editView.name.Value()
				}
				if m.editView.endpoint.Value() != "" {
					api.Endpoint = m.editView.endpoint.Value()
				}
				if m.editView.customModel.Value() != "" {
					api.CustomModelID = m.editView.customModel.Value()
				}
				if m.editView.targetModel.Value() != "" {
					api.TargetModelID = m.editView.targetModel.Value()
				}
				api.StreamMode = m.editView.getStreamMode()
				if m.editView.setActive {
					api.Active = m.editView.active
					if api.Active {
						for i := range m.config.APIs {
							if i != m.editView.index {
								m.config.APIs[i].Active = false
							}
						}
					}
				}
				if err := config.SaveConfig(m.config, configFile); err != nil {
					m.err = err
				} else {
					m.config, _ = config.LoadConfig(configFile)
					m.listView = newListView(m.config)
					m.view = viewList
				}
			}
			m.editView.done = false
		}
	case viewDomain:
		m.domainView, cmd = m.domainView.update(msg)
		if m.domainView.done {
			if m.domainView.err == nil {
				m.config.Domain = m.domainView.domain.Value()
				if err := config.SaveConfig(m.config, configFile); err != nil {
					m.err = err
				} else {
					m.config, _ = config.LoadConfig(configFile)
					m.listView = newListView(m.config)
					m.view = viewList
				}
			}
			m.domainView.done = false
		}
	case viewCert:
		m.certView, cmd = m.certView.update(msg)
		if m.certView.done {
			m.view = viewList
			m.certView.done = false
		}
	}

	return m, cmd
}

func (m model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := m.executeCommand()
		m.vimMode = vimModeNormal
		m.commandBuf = ""
		return m, cmd
	case "esc":
		m.vimMode = vimModeNormal
		m.commandBuf = ""
		return m, nil
	case "backspace":
		if len(m.commandBuf) > 0 {
			m.commandBuf = m.commandBuf[:len(m.commandBuf)-1]
		}
		return m, nil
	default:
		// Only accept valid vim command characters
		if len(msg.String()) == 1 && isValidCommandChar(msg.String()[0]) {
			m.commandBuf += msg.String()
		}
		return m, nil
	}
}

// isValidCommandChar checks if a character is valid for vim commands
func isValidCommandChar(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
	       (char >= '0' && char <= '9') || char == '-' || char == '_'
}

func (m model) executeCommand() tea.Cmd {
	switch m.commandBuf {
	case "q", "quit":
		return tea.Quit
	case "w", "write":
		// Configuration is auto-saved in this TUI, no action needed
		return nil
	case "wq":
		return tea.Quit
	}
	return nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("错误: %v\n\n按 q 退出", m.err)
	}

	var content string
	switch m.view {
	case viewList:
		content = m.listView.view(m.vimMode)
	case viewAdd:
		content = m.addView.view()
	case viewEdit:
		content = m.editView.view()
	case viewDomain:
		content = m.domainView.view()
	case viewCert:
		content = m.certView.view()
	default:
		content = "未知视图"
	}

	// Add vim mode indicator and command buffer at the bottom
	if m.view == viewList {
		modeIndicator := ""
		switch m.vimMode {
		case vimModeNormal:
			modeIndicator = vimModeStyle.Render("-- NORMAL --")
		case vimModeCommand:
			modeIndicator = vimModeStyle.Render(":" + m.commandBuf)
		}
		if modeIndicator != "" {
			content += "\n" + modeIndicator
		}
	}

	return content
}

func Run() error {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

