package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0088CC")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title          string
	description    string
	containerID    string
	containerState string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	toggleHelpMenu      key.Binding
	toggleAllContainers key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		toggleAllContainers: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "toggle all/running containers"),
		),
	}
}

type model struct {
	list          list.Model
	keys          *listKeyMap
	delegateKeys  *delegateKeyMap
	dc            DockerClientWrapper
	allContainers bool
}

func newModel(dc DockerClientWrapper) model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
		all          = true
	)

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	containerList := list.NewModel(GetContainerListItems(dc, all), delegate, 0, 0)
	containerList.Title = "Docker containers"
	containerList.Styles.Title = titleStyle
	containerList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.toggleAllContainers,
		}
	}

	m := model{
		list:          containerList,
		keys:          listKeys,
		delegateKeys:  delegateKeys,
		dc:            dc,
		allContainers: all,
	}

	m.SetListTitleForMode(all)

	return m
}

func (m *model) SetListTitleForMode(all bool) {
	title := "Docker containers"

	if all == true {
		title += " (all)"
	} else {
		title += " (running)"
	}

	m.list.Title = title
}

func (m *model) ToggleAllContainers() tea.Cmd {
	m.allContainers = !m.allContainers
	m.SetListTitleForMode(m.allContainers)

	return RefreshListCmd()
}

func (m model) GetDockerClientWrapper() DockerClientWrapper {
	return m.dc
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		case key.Matches(msg, m.keys.toggleAllContainers):
			cmd := m.ToggleAllContainers()
			return m, cmd
		}
	case StartDockerContainerMsg:
		m.GetDockerClientWrapper().StartContainer(msg.ContainerID)
		return m, nil
	case StopDockerContainerMsg:
		m.GetDockerClientWrapper().StopContainer(msg.ContainerID)
		return m, nil
	case RefreshListMsg:
		setItemsCmd := m.list.SetItems(GetContainerListItems(m.dc, m.allContainers))
		return m, setItemsCmd
	case ErrorNotificationMsg:
		return m, m.list.NewStatusMessage(statusMessageStyle(msg.msg))
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func main() {
	var dc DockerClientWrapper
	p := tea.NewProgram(newModel(dc))

	go func() {
		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}()

	msgs, errs := dc.GetDeamonEventStreams()

	for {
		select {
		case err := <-errs:
			p.Send(ErrorNotificationMsg{err.Error()})
		case msg := <-msgs:
			if msg.Action == "start" || msg.Action == "stop" {
				p.Send(RefreshListMsg{})
			}
		}
	}
}

type RefreshListMsg struct{}

func RefreshListCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshListMsg{}
	}
}

type ErrorNotificationMsg struct{ msg string }

func GetContainerListItems(dc DockerClientWrapper, all bool) []list.Item {
	containers := dc.GetContainerList(all)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Names[0] < containers[j].Names[0]
	})

	items := make([]list.Item, 0)

	for _, c := range containers {
		items = append(items, item{
			title:          c.Names[0],
			description:    fmt.Sprintf("%s %s", c.ID[0:12], c.State),
			containerID:    c.ID,
			containerState: c.State,
		})
	}

	return items
}
