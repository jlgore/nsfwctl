package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type SlideModel struct {
	content      []string
	currentSlide int
	renderer     *glamour.TermRenderer
}

func NewSlideModel(markdownContent string) (SlideModel, error) {
	var slides []string
	for _, slide := range strings.Split(markdownContent, "---") {
		slides = append(slides, strings.TrimSpace(slide))
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return SlideModel{}, err
	}

	return SlideModel{
		content:      slides,
		currentSlide: 0,
		renderer:     renderer,
	}, nil
}

func (m SlideModel) Update(msg tea.Msg) (SlideModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "n"))):
			if m.currentSlide < len(m.content)-1 {
				m.currentSlide++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "p"))):
			if m.currentSlide > 0 {
				m.currentSlide--
			}
		}
	}

	return m, nil
}

func (m SlideModel) View() string {
	if len(m.content) == 0 {
		return "No content to display"
	}

	renderedContent, err := m.renderer.Render(m.content[m.currentSlide])
	if err != nil {
		return "Error rendering markdown: " + err.Error()
	}

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	progress := progressStyle.Render(fmt.Sprintf("Slide %d of %d", m.currentSlide+1, len(m.content)))

	return fmt.Sprintf("%s\n\n%s\n\n(Use arrow keys to navigate, 'd' for deployment options)", progress, renderedContent)
}
