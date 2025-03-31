package model

import (
	"fmt"
	"math"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

type focusableSection int

const (
	focusNotes focusableSection = iota
	focusItems
)

const (
	cardsPerPage    = 3
	cardWidth       = 35
	cardMargin      = 1
	fixedCardHeight = 9
)

var maxBlurbLines = max(1, fixedCardHeight-4)

type CollectionModel struct {
	GameSave         *data.FullGameSave
	width            int
	notesCurrentPage int
	itemsCurrentPage int
	focusedSection   focusableSection
}

// NewCollectionModel initializes pagination state
func NewCollectionModel(gameSave *data.FullGameSave) CollectionModel {
	return CollectionModel{
		GameSave:         gameSave,
		notesCurrentPage: 1,
		itemsCurrentPage: 1,
		focusedSection:   focusNotes,
		width:            80,
	}
}

func (m CollectionModel) Init() tea.Cmd {
	return nil
}

func (m *CollectionModel) SetWidth(w int) {
	m.width = w
}

// update remains simple for now, potentially add sorting/filtering later
func (m CollectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.GameSave == nil {
		return m, nil
	}

	collection := m.GameSave.Collection

	totalNoteCards := 0
	for _, note := range collection.ResearchNotes {
		if note.Quantity > 0 {
			totalNoteCards++
		}
	}
	totalItemCards := 0
	for _, item := range collection.Items {
		if item.Quantity > 0 {
			totalItemCards++
		}
	}

	totalNotePages := 1
	if totalNoteCards > 0 {
		totalNotePages = int(math.Ceil(float64(totalNoteCards) / float64(cardsPerPage)))
	}
	totalItemPages := 1
	if totalItemCards > 0 {
		totalItemPages = int(math.Ceil(float64(totalItemCards) / float64(cardsPerPage)))
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.focusedSection == focusItems {
				m.focusedSection = focusNotes
			}
		case "down", "j":
			if m.focusedSection == focusNotes && totalItemCards > 0 {
				m.focusedSection = focusItems
			}
		case "left", "h":
			if m.focusedSection == focusNotes {
				if m.notesCurrentPage > 1 {
					m.notesCurrentPage--
				}
			} else if m.focusedSection == focusItems {
				if m.itemsCurrentPage > 1 {
					m.itemsCurrentPage--
				}
			}
		case "right", "l":
			if m.focusedSection == focusNotes {
				if m.notesCurrentPage < totalNotePages {
					m.notesCurrentPage++
				}
			} else if m.focusedSection == focusItems {
				if m.itemsCurrentPage < totalItemPages {
					m.itemsCurrentPage++
				}
			}
		}
	}
	return m, nil
}

// helper function to get color based on tier
func getTierColor(tier int) lipgloss.Color {
	switch tier {
	case 1:
		return lipgloss.Color("15")
	case 2:
		return lipgloss.Color("40")
	case 3:
		return lipgloss.Color("214")
	case 4:
		return lipgloss.Color("199")
	case 5:
		return lipgloss.Color("196")
	default:
		return lipgloss.Color("248")
	}
}

// truncateLines limits the number of lines and adds ellipsis to the last visible line
func truncateLines(renderedWrappedText string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(renderedWrappedText, "\n")
	if len(lines) <= maxLines {
		return renderedWrappedText
	}

	linesToShow := lines[:maxLines]
	ellipsisLineIndex := maxLines - 1

	linesToShow[ellipsisLineIndex] = linesToShow[ellipsisLineIndex] + "..."

	return strings.Join(linesToShow, "\n")
}

// view renders the Collection view with pagination
func (m CollectionModel) View() string {
	if m.GameSave == nil {
		return "Error: Collection data unavailable (GameSave is nil)."
	}

	collection := m.GameSave.Collection

	mainTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).MarginBottom(1)
	sectionTitleBaseStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	focusedSectionTitleStyle := sectionTitleBaseStyle.Copy().Foreground(lipgloss.Color("214"))
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginRight(cardMargin).
		Width(cardWidth).
		Height(fixedCardHeight).
		BorderForeground(lipgloss.Color("240"))
	cardDetailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	cardBlurbRenderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(cardWidth - 2)
	paginationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	focusedPaginationStyle := paginationStyle.Copy().Bold(true).Foreground(lipgloss.Color("250"))
	headerLineStyle := lipgloss.NewStyle().MarginBottom(1)
	cardRowStyle := lipgloss.NewStyle()

	var allNoteCardStrings []string
	sort.SliceStable(collection.ResearchNotes, func(i, j int) bool { return collection.ResearchNotes[i].Tier < collection.ResearchNotes[j].Tier })
	for _, note := range collection.ResearchNotes {
		if note.Quantity > 0 {
			tierColor := getTierColor(note.Tier)
			dynamicCardTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(tierColor)
			title := dynamicCardTitleStyle.Render(note.Name)
			details := cardDetailStyle.Render(fmt.Sprintf("Tier: %d | Qty: %d | XP: %d", note.Tier, note.Quantity, note.XP))
			renderedFullBlurb := cardBlurbRenderStyle.Render(note.Blurb)
			truncatedBlurb := truncateLines(renderedFullBlurb, maxBlurbLines)

			content := lipgloss.JoinVertical(lipgloss.Left, title, details, truncatedBlurb)
			renderedCard := cardStyle.Render(content)
			allNoteCardStrings = append(allNoteCardStrings, renderedCard)
		}
	}

	var allItemCardStrings []string
	sort.SliceStable(collection.Items, func(i, j int) bool {
		if collection.Items[i].Tier != collection.Items[j].Tier {
			return collection.Items[i].Tier < collection.Items[j].Tier
		}
		return collection.Items[i].Name < collection.Items[j].Name
	})
	for _, item := range collection.Items {
		if item.Quantity > 0 {
			tierColor := getTierColor(item.Tier)
			dynamicCardTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(tierColor)
			title := dynamicCardTitleStyle.Render(item.Name)
			details := cardDetailStyle.Render(fmt.Sprintf("Tier: %d | Qty: %d", item.Tier, item.Quantity))
			renderedFullDesc := cardBlurbRenderStyle.Render(item.Description)
			truncatedDesc := truncateLines(renderedFullDesc, maxBlurbLines)

			content := lipgloss.JoinVertical(lipgloss.Left, title, details, truncatedDesc)
			renderedCard := cardStyle.Render(content)
			allItemCardStrings = append(allItemCardStrings, renderedCard)
		}
	}

	totalNoteCards := len(allNoteCardStrings)
	totalNotePages := 1
	if totalNoteCards > 0 {
		totalNotePages = int(math.Ceil(float64(totalNoteCards) / float64(cardsPerPage)))
	}
	totalItemCards := len(allItemCardStrings)
	totalItemPages := 1
	if totalItemCards > 0 {
		totalItemPages = int(math.Ceil(float64(totalItemCards) / float64(cardsPerPage)))
	}

	notesPage := max(1, min(m.notesCurrentPage, totalNotePages))
	itemsPage := max(1, min(m.itemsCurrentPage, totalItemPages))

	notesStart := (notesPage - 1) * cardsPerPage
	notesEnd := min(notesStart+cardsPerPage, totalNoteCards)
	currentPageNoteCards := allNoteCardStrings[notesStart:notesEnd]

	itemsStart := (itemsPage - 1) * cardsPerPage
	itemsEnd := min(itemsStart+cardsPerPage, totalItemCards)
	currentPageItemCards := allItemCardStrings[itemsStart:itemsEnd]

	var mainBuilder strings.Builder
	mainBuilder.WriteString(mainTitleStyle.Render("PLAYER Collection") + "\n")

	var calculatedUsedCapacity int
	for _, note := range collection.ResearchNotes {
		calculatedUsedCapacity += note.Quantity
	}
	for _, item := range collection.Items {
		calculatedUsedCapacity += item.Quantity
	}
	currentUsedCapacity := calculatedUsedCapacity
	maxCapacity := collection.MaxCapacity
	capacityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	if maxCapacity > 0 {
		usageRatio := float64(currentUsedCapacity) / float64(maxCapacity)
		if usageRatio >= 1.0 {
			capacityStyle = capacityStyle.Foreground(lipgloss.Color("196")).Bold(true)
		} else if usageRatio > 0.85 {
			capacityStyle = capacityStyle.Foreground(lipgloss.Color("208"))
		}
	}
	mainBuilder.WriteString(capacityStyle.Render(fmt.Sprintf("Capacity: %d / %d", currentUsedCapacity, maxCapacity)) + "\n\n")

	var sections []string
	createHeaderLine := func(titleStr, paginationStr string, availableWidth int) string {
		titleWidth := lipgloss.Width(titleStr)
		paginationWidth := lipgloss.Width(paginationStr)
		spacing := availableWidth - titleWidth - paginationWidth
		if spacing < 1 {
			spacing = 1
		}
		spacer := strings.Repeat(" ", spacing)
		return lipgloss.JoinHorizontal(lipgloss.Left, titleStr, spacer, paginationStr)
	}

	if totalNoteCards > 0 {
		titleStyle := sectionTitleBaseStyle
		pageStyle := paginationStyle
		if m.focusedSection == focusNotes {
			titleStyle, pageStyle = focusedSectionTitleStyle, focusedPaginationStyle
		}
		titleStr := titleStyle.Render("Research Notes")
		paginationStr := pageStyle.Render(fmt.Sprintf("[ < Page %d / %d > ]", notesPage, totalNotePages))
		headerLine := createHeaderLine(titleStr, paginationStr, m.width+32)
		sections = append(sections, headerLineStyle.Render(headerLine))

		notesContent := lipgloss.JoinHorizontal(lipgloss.Top, currentPageNoteCards...)
		sections = append(sections, cardRowStyle.Render(notesContent))
	} else if m.focusedSection == focusNotes {
		titleStr := focusedSectionTitleStyle.Render("Research Notes")
		headerLine := createHeaderLine(titleStr, "", m.width)
		sections = append(sections, headerLineStyle.Render(headerLine), paginationStyle.Render("  (No notes)"))
	}

	marginTopStyle := lipgloss.NewStyle()
	if len(sections) > 0 {
		marginTopStyle = marginTopStyle.MarginTop(1)
	}

	if totalItemCards > 0 {
		titleStyle := sectionTitleBaseStyle
		pageStyle := paginationStyle
		if m.focusedSection == focusItems {
			titleStyle, pageStyle = focusedSectionTitleStyle, focusedPaginationStyle
		}
		titleStr := titleStyle.Render("Items")
		paginationStr := pageStyle.Render(fmt.Sprintf("[ < Page %d / %d > ]", itemsPage, totalItemPages))
		headerLine := createHeaderLine(titleStr, paginationStr, m.width+32)
		sections = append(sections, marginTopStyle.Render(headerLineStyle.Render(headerLine)))

		itemsContent := lipgloss.JoinHorizontal(lipgloss.Top, currentPageItemCards...)
		sections = append(sections, cardRowStyle.Render(itemsContent))
	} else if m.focusedSection == focusItems {
		titleStr := focusedSectionTitleStyle.Render("Items")
		headerLine := createHeaderLine(titleStr, "", m.width)
		sections = append(sections, marginTopStyle.Render(headerLineStyle.Render(headerLine)))
		sections = append(sections, paginationStyle.Render("  (No items)"))
	}

	if len(sections) > 0 {
		mainBuilder.WriteString(lipgloss.JoinVertical(lipgloss.Left, sections...))
	} else {
		mainBuilder.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Render("  Collection is empty.") + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(mainBuilder.String())
}
