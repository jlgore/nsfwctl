package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (m Model) View() string {
	log.Printf("Entering View function, current state: %v", m.state)

	switch m.state {
	case StateSelectingBranch:
		log.Printf("Rendering StateSelectingBranch view")
		log.Printf("List has %d items", len(m.list.Items()))

		var itemsDebug strings.Builder
		for i, item := range m.list.Items() {
			itemsDebug.WriteString(fmt.Sprintf("Item %d: %+v\n", i, item))
		}
		log.Printf("Items in list:\n%s", itemsDebug.String())

		log.Printf("List width: %d, height: %d", m.list.Width(), m.list.Height())

		visibleItems := []list.Item{}
		for i := m.list.Index(); i < len(m.list.Items()) && len(visibleItems) < m.list.Height(); i++ {
			visibleItems = append(visibleItems, m.list.Items()[i])
		}
		log.Printf("List visible items: %+v", visibleItems)

		listView := m.list.View()
		log.Printf("List view: \n%s", listView)

		fullView := appStyle.Render(fmt.Sprintf(
			"nsfwctl\n\n"+
				"Repository: %s\n"+
				"Status: %s\n\n%s",
			m.repoPath,
			m.status,
			listView,
		))

		log.Printf("Full view: \n%s", fullView)

		return fullView

	case StateViewingSlides:
		log.Printf("Rendering StateViewingSlides view")
		slideView := m.slideModel.View()
		log.Printf("Slide view: \n%s", slideView)
		return slideView

	case StateDeploymentOptions:
		log.Printf("Rendering StateDeploymentOptions view")
		deploymentView := appStyle.Render(fmt.Sprintf(
			"nsfwctl - Deploy Branch: %s\n\n"+
				"Deployment Options:\n"+
				"1. Deploy\n"+
				"2. Cancel\n\n"+
				"Press 1 to deploy or 2 to cancel",
			m.selectedBranch,
		))
		log.Printf("Deployment view: \n%s", deploymentView)
		return deploymentView

	default:
		log.Printf("Unknown state: %v", m.state)
		return "Unknown state"
	}
}
