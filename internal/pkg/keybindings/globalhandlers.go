package keybindings

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/lawrencegripper/azbrowse/internal/pkg/style"
	"github.com/lawrencegripper/azbrowse/internal/pkg/views"
	"github.com/lawrencegripper/azbrowse/internal/pkg/wsl"
	"github.com/stuartleeks/gocui"
)

////////////////////////////////////////////////////////////////////
type QuitHandler struct {
	GlobalHandler
}

func NewQuitHandler() *QuitHandler {
	handler := &QuitHandler{}
	handler.id = HandlerIDQuit
	return handler
}

func (h QuitHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type CopyHandler struct {
	GlobalHandler
	Content   *views.ItemWidget
	StatusBar *views.StatusbarWidget
}

var _ Command = &CopyHandler{}

func NewCopyHandler(content *views.ItemWidget, statusbar *views.StatusbarWidget) *CopyHandler {
	handler := &CopyHandler{
		Content:   content,
		StatusBar: statusbar,
	}
	handler.id = HandlerIDCopy
	return handler
}

func (h CopyHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return h.Invoke()
	}
}

func (h *CopyHandler) DisplayText() string {
	return "Copy content"
}

func (h *CopyHandler) IsEnabled() bool {
	return true
}

func (h *CopyHandler) Invoke() error {
	var err error
	if wsl.IsWSL() {
		err = wsl.TrySetClipboard(h.Content.GetContent())
	} else {
		err = clipboard.WriteAll(h.Content.GetContent())
	}
	if err != nil {
		h.StatusBar.Status(fmt.Sprintf("Failed to copy to clipboard: %s", err.Error()), false)
		return nil
	}
	h.StatusBar.Status("Current resource's content copied to clipboard", false)
	return nil
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type FullscreenHandler struct {
	GlobalHandler
	List         *views.ListWidget
	IsFullscreen *bool
	Content      *views.ItemWidget
}

func NewFullscreenHandler(list *views.ListWidget, content *views.ItemWidget, isFullscreen *bool) *FullscreenHandler {
	handler := &FullscreenHandler{
		List:         list,
		Content:      content,
		IsFullscreen: isFullscreen,
	}
	handler.id = HandlerIDFullScreen
	return handler
}

func (h FullscreenHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		tmp := toggle(*h.IsFullscreen)
		h.IsFullscreen = &tmp // memory leak?
		if *h.IsFullscreen {
			g.Cursor = true
			maxX, maxY := g.Size()
			v, _ := g.SetView("fullscreenContent", 0, 0, maxX, maxY)
			v.Editable = true
			v.Frame = false
			v.Wrap = true
			keyBindings := GetKeyBindingsAsStrings()
			v.Title = fmt.Sprintf("JSON Response - Fullscreen (%s to exit)", strings.ToUpper(strings.Join(keyBindings["fullscreen"], ",")))

			content := h.Content.GetContent()
			fmt.Fprint(v, style.ColorJSON(content))

			g.SetCurrentView("fullscreenContent")
		} else {
			g.Cursor = false
			g.DeleteView("fullscreenContent")
			g.SetCurrentView("listWidget")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type HelpHandler struct {
	GlobalHandler
	ShowHelp *bool
}

func NewHelpHandler(showHelp *bool) *HelpHandler {
	handler := &HelpHandler{
		ShowHelp: showHelp,
	}
	handler.id = HandlerIDHelp
	return handler
}

func (h HelpHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		tmp := toggle(*h.ShowHelp)
		h.ShowHelp = &tmp // memory leak?

		// If we're up and running clear and redraw the view
		// if w.g != nil {
		if *h.ShowHelp {
			v, err := g.SetView("helppopup", 1, 1, 145, 45)
			g.SetCurrentView("helppopup")
			if err != nil && err != gocui.ErrUnknownView {
				panic(err)
			}
			keyBindings := GetKeyBindingsAsStrings()
			views.DrawHelp(keyBindings, v)
		} else {
			g.DeleteView("helppopup")
			g.SetCurrentView("listWidget")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type ConfirmDeleteHandler struct {
	GlobalHandler
	notificationWidget *views.NotificationWidget
}

func NewConfirmDeleteHandler(notificationWidget *views.NotificationWidget) *ConfirmDeleteHandler {
	handler := &ConfirmDeleteHandler{
		notificationWidget: notificationWidget,
	}
	handler.id = HandlerIDConfirmDelete
	return handler
}

func (h *ConfirmDeleteHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		h.notificationWidget.ConfirmDelete()
		return nil
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type ClearPendingDeleteHandler struct {
	GlobalHandler
	notificationWidget *views.NotificationWidget
}

func NewClearPendingDeleteHandler(notificationWidget *views.NotificationWidget) *ClearPendingDeleteHandler {
	handler := &ClearPendingDeleteHandler{
		notificationWidget: notificationWidget,
	}
	handler.id = HandlerIDClearPendingDeletes
	return handler
}

func (h *ClearPendingDeleteHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		h.notificationWidget.ClearPendingDeletes()
		return nil
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type OpenCommandPanelHandler struct {
	GlobalHandler
	gui                *gocui.Gui
	commandPanelWidget *views.CommandPanelWidget
	list               *views.ListWidget
	commands           []Command
}

func NewOpenCommandPanelHandler(gui *gocui.Gui, commandPanelWidget *views.CommandPanelWidget, commands []Command) *OpenCommandPanelHandler {
	handler := &OpenCommandPanelHandler{
		gui:                gui,
		commandPanelWidget: commandPanelWidget,
		commands:           commands,
	}
	handler.id = HandlerIDToggleOpenCommandPanel
	return handler
}

func (h *OpenCommandPanelHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		options := []views.CommandPanelListOption{}
		for _, command := range h.commands {
			if command.IsEnabled() {
				options = append(options, command)
			}
		}
		h.commandPanelWidget.ShowWithText("Command Palette", "", &options, h.CommandPanelNotification)
		return nil
	}
}

func (h *OpenCommandPanelHandler) CommandPanelNotification(state views.CommandPanelNotification) {
	if state.EnterPressed {
		h.commandPanelWidget.Hide()
		for _, command := range h.commands {
			if command.ID() == state.SelectedID {
				// invoke via Update to allow Hide to restore preview view state
				h.gui.Update(func(gui *gocui.Gui) error {
					command.Invoke()
					return nil
				})
				return
			}
		}
	}
}

////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////
type CommandPanelFilterHandler struct {
	GlobalHandler
	commandPanelWidget *views.CommandPanelWidget
	list               *views.ListWidget
}

var _ Command = &CommandPanelFilterHandler{}

func NewCommandPanelFilterHandler(commandPanelWidget *views.CommandPanelWidget, list *views.ListWidget) *CommandPanelFilterHandler {
	handler := &CommandPanelFilterHandler{
		commandPanelWidget: commandPanelWidget,
		list:               list,
	}
	handler.id = HandlerIDFilter
	return handler
}

func (h *CommandPanelFilterHandler) Fn() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		h.Invoke()
		return nil
	}
}
func (h *CommandPanelFilterHandler) DisplayText() string {
	return "filter (/)"
}
func (h *CommandPanelFilterHandler) IsEnabled() bool {
	return true
}
func (h *CommandPanelFilterHandler) Invoke() error {
	h.commandPanelWidget.ShowWithText("Filter", "", nil, h.CommandPanelNotification)
	return nil
}
func (h *CommandPanelFilterHandler) CommandPanelNotification(state views.CommandPanelNotification) {
	h.list.SetFilter(state.CurrentText)
	if state.EnterPressed {
		h.commandPanelWidget.Hide()
	}
}

////////////////////////////////////////////////////////////////////

type Command interface {
	ID() string
	DisplayText() string
	IsEnabled() bool
	Invoke() error
}
