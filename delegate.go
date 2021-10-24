package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title, containerID, containerState string

		if i, ok := m.SelectedItem().(item); ok {
			title = i.Title()
			containerID = i.containerID
			containerState = i.containerState
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				if containerState == "exited" {
					return tea.Batch(m.NewStatusMessage(statusMessageStyle("Starting "+title)), StartDockerContainerCmd(containerID))
				}

				if containerState == "running" {
					return tea.Batch(m.NewStatusMessage(statusMessageStyle("Stopping "+title)), StopDockerContainerCmd(containerID))
				}
			}
		}

		return nil
	}

	help := []key.Binding{keys.choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type StartDockerContainerMsg struct {
	ContainerID string
}

type StopDockerContainerMsg struct {
	ContainerID string
}

func StartDockerContainerCmd(containerId string) tea.Cmd {
	return func() tea.Msg {
		return StartDockerContainerMsg{ContainerID: containerId}
	}
}

func StopDockerContainerCmd(containerId string) tea.Cmd {
	return func() tea.Msg {
		return StopDockerContainerMsg{ContainerID: containerId}
	}
}

type delegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Start/Stop container"),
		),
	}
}
