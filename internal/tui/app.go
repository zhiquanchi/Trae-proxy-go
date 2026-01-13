package tui

import (
	"fmt"
	"trae-proxy-go/internal/config"
	"trae-proxy-go/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
)

const configFile = "config.yaml"

type model struct {
	view      viewType
	config    *models.Config
	selected  int
	err       error
	listView  listViewModel
	addView   addViewModel
	editView  editViewModel
	domainView domainViewModel
	certView  certViewModel
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
		view:      viewList,
		config:    cfg,
		selected:  0,
		listView:  newListView(cfg),
		addView:   newAddView(),
		editView:  newEditView(),
		domainView: newDomainView(cfg.Domain),
		certView:  newCertView(cfg.Domain),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.view == viewList {
				return m, tea.Quit
			}
			// 从其他视图返回列表视图
			m.view = viewList
			// 重新加载配置
			cfg, err := config.LoadConfig(configFile)
			if err == nil {
				m.config = cfg
				m.listView = newListView(cfg)
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
		m.listView, cmd = m.listView.update(msg, m.config)
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

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("错误: %v\n\n按 q 退出", m.err)
	}

	switch m.view {
	case viewList:
		return m.listView.view()
	case viewAdd:
		return m.addView.view()
	case viewEdit:
		return m.editView.view()
	case viewDomain:
		return m.domainView.view()
	case viewCert:
		return m.certView.view()
	default:
		return "未知视图"
	}
}

func Run() error {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

