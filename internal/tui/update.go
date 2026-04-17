package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/flyingnobita/llml/internal/models"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case themeToastClearMsg:
		m.ui.themeToast = ""
		m = m.layoutTable()
		return m, nil

	case tea.WindowSizeMsg:
		m.layout.width = msg.Width
		m.layout.height = msg.Height
		m = m.layoutTable()
		if m.params.open {
			m.params.editInput.SetWidth(m.paramEditInnerWidth())
		}
		return m, nil

	case runtimeReadyMsg:
		m.runtime = msg.runtime
		m.runtimeScanned = true
		return m, nil

	case startupCacheHitMsg:
		m.loading = false
		m.loadErr = nil
		m.runtime = msg.runtime
		m.runtimeScanned = true
		m.table.files = msg.files
		m.table.lastScan = msg.lastScan
		sortModelFiles(m.table.files, m.table.sortCol, m.table.sortDesc)
		m = m.layoutTable()
		m.table.hscroll.SetXOffset(0)
		if len(m.table.files) > 0 {
			m.table.tbl.SetCursor(0)
		}
		return m.maybeSetMissingRuntimeFooterNote()

	case startupNeedFullScanMsg:
		return m, applyAndFullScanCmd()

	case fullScanDoneMsg:
		m.loading = false
		m.loadErr = nil
		m.runtime = msg.runtime
		m.runtimeScanned = true
		m.table.files = msg.files
		m.table.lastScan = msg.lastScan
		sortModelFiles(m.table.files, m.table.sortCol, m.table.sortDesc)
		m = m.layoutTable()
		m.table.hscroll.SetXOffset(0)
		if len(m.table.files) > 0 {
			m.table.tbl.SetCursor(0)
		}
		if msg.writeErr != nil {
			m = m.withLastRunError("Could not save config: " + msg.writeErr.Error())
		} else {
			m = m.withLastRunCleared()
		}
		return m.maybeSetMissingRuntimeFooterNote()

	case modelRescanDoneMsg:
		m.loading = false
		m.loadErr = nil
		m.table.files = msg.files
		m.table.lastScan = msg.lastScan
		sortModelFiles(m.table.files, m.table.sortCol, m.table.sortDesc)
		m = m.layoutTable()
		m.table.hscroll.SetXOffset(0)
		if len(m.table.files) > 0 && m.table.tbl.Cursor() >= len(m.table.files) {
			m.table.tbl.SetCursor(len(m.table.files) - 1)
		}
		if msg.writeErr != nil {
			m = m.withLastRunError("Could not save config: " + msg.writeErr.Error())
		} else {
			m = m.withLastRunCleared()
		}
		return m.maybeSetMissingRuntimeFooterNote()

	case runtimeReloadErrMsg:
		m = m.withLastRunError(msg.err.Error())
		return m, nil

	case modelsLoadedMsg:
		m.loading = false
		m.loadErr = nil
		m.table.files = msg.files
		sortModelFiles(m.table.files, m.table.sortCol, m.table.sortDesc)
		m = m.layoutTable()
		m.table.hscroll.SetXOffset(0)
		if len(m.table.files) > 0 {
			m.table.tbl.SetCursor(0)
		}
		return m.maybeSetMissingRuntimeFooterNote()

	case modelsErrMsg:
		m.loading = false
		m.loadErr = msg.err
		return m, nil

	case runServerErrMsg:
		m = m.withLastRunError(msg.err.Error())
		return m, nil

	case llamaServerExitedMsg:
		if m.server.running {
			m.server.exited = true
			m.server.cmd = nil
			m.server.msgCh = nil
			if msg.err != nil {
				m = m.withLastRunError(msg.err.Error())
				m = m.appendServerLogLine(fmt.Sprintf("%s · %s", msg.err.Error(), splitPanePressEnterToClose))
			} else {
				m = m.withLastRunCleared()
				m = m.appendServerLogLine(splitServerStoppedWithHint)
			}
			m = m.layoutTable()
			return m, nil
		}
		if msg.err != nil {
			m = m.withLastRunError(msg.err.Error())
		} else {
			m = m.withLastRunCleared()
		}
		return m, nil

	case serverSplitReadyMsg:
		m.server.running = true
		m.server.exited = false
		m.server.splitFocused = false
		m.preview.focused = false
		m.server.cmd = msg.cmd
		m.server.msgCh = msg.ch
		m.table.tbl.Focus()
		m = m.layoutTable()
		return m, readNextServerMsg(msg.ch)

	case serverLogMsg:
		m = m.appendServerLogLine(msg.line)
		m = m.layoutTable()
		return m, readNextServerMsg(m.server.msgCh)

	case splitInterruptMsg:
		return m.stopSplitServer()

	case tea.MouseWheelMsg:
		if m.server.running {
			var cmd tea.Cmd
			if m.server.splitFocused {
				m.server.viewport, cmd = m.server.viewport.Update(msg)
			} else {
				m.table.tbl, cmd = m.table.tbl.Update(msg)
			}
			return m, cmd
		}
		if m.preview.focused {
			var cmd tea.Cmd
			m.preview.viewport, cmd = m.preview.viewport.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.table.tbl, cmd = m.table.tbl.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if m.params.open {
			return m.updateParamPanelKey(msg)
		}
		if m.rc.open {
			return m.updateRuntimeConfigKey(msg)
		}
		if m.server.running {
			return m.updateServerSplitKeys(msg)
		}
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		if m.preview.focused && isTabKey(msg) {
			m.preview.focused = false
			m.table.tbl.Focus()
			m = m.applyMainPaneFocusStyles()
			return m, nil
		}
		if key.Matches(msg, m.keys.ConfigPort) {
			return m.openRuntimeConfig()
		}
		if key.Matches(msg, m.keys.Parameters) {
			if m.loading {
				m = m.withLastRunError("Wait for the model scan to finish.")
				return m, nil
			}
			return m.openParamPanel()
		}
		if key.Matches(msg, m.keys.ToggleTheme) {
			var cmd tea.Cmd
			m, cmd = m.cycleTheme()
			return m, cmd
		}
		if key.Matches(msg, m.keys.RescanModels) {
			if m.loading {
				return m, nil
			}
			if m.server.running {
				m = m.withLastRunError("Stop the server before re-scanning models.")
				return m, nil
			}
			m.loading = true
			m.loadErr = nil
			m = m.withLastRunCleared()
			return m, rescanModelsCmd()
		}
		if key.Matches(msg, m.keys.Refresh) {
			if m.loading {
				return m, nil
			}
			if m.server.running {
				m = m.withLastRunError("Stop the server before reloading runtime.")
				return m, nil
			}
			m = m.withLastRunCleared()
			return m, reloadRuntimeCmd()
		}
		mode := runServerKeyMode(msg)
		if mode != runServerModeNone {
			if m.loading {
				m = m.withLastRunError("Wait for the model scan to finish.")
				return m, nil
			}
			p, be := m.SelectedModel()
			if p == "" {
				m = m.withLastRunError("Select a model row first.")
				return m, nil
			}
			m = m.withLastRunCleared()
			params, _ := loadModelParamsForRun(p)
			if mode == runServerModeFullscreen {
				if be == models.BackendVLLM {
					return m, runVLLMServerCmd(p, m.runtime, params)
				}
				return m, runLlamaServerCmd(p, m.runtime, params)
			}
			if be == models.BackendVLLM {
				return m, runVLLMServerSplitCmd(p, m.runtime, params)
			}
			return m, runLlamaServerSplitCmd(p, m.runtime, params)
		}
		if key.Matches(msg, m.keys.ScrollLeft) {
			m.table.hscroll.ScrollLeft(hScrollStep)
			return m, nil
		}
		if key.Matches(msg, m.keys.ScrollRight) {
			m.table.hscroll.ScrollRight(hScrollStep)
			return m, nil
		}
		if key.Matches(msg, m.keys.CopyPath) {
			m = copyLaunchCommandToClipboard(m)
			return m, nil
		}
		if key.Matches(msg, m.keys.SortColumn) {
			if m.loading || len(m.table.files) == 0 {
				return m, nil
			}
			sel := m.SelectedPath()
			m.table.sortCol = (m.table.sortCol + 1) % tableSortColCount
			m = m.applyTableSort(sel)
			return m, nil
		}
		if key.Matches(msg, m.keys.SortReverse) {
			if m.loading || len(m.table.files) == 0 {
				return m, nil
			}
			sel := m.SelectedPath()
			m.table.sortDesc = !m.table.sortDesc
			m = m.applyTableSort(sel)
			return m, nil
		}
		if m.preview.focused {
			var cmd tea.Cmd
			m.preview.viewport, cmd = m.preview.viewport.Update(msg)
			return m, cmd
		}
		if launchPreviewScrollable(m) {
			if key.Matches(msg, m.keys.LaunchPreviewScrollUp) {
				m.preview.viewport.ScrollUp(1)
				return m, nil
			}
			if key.Matches(msg, m.keys.LaunchPreviewScrollDown) {
				m.preview.viewport.ScrollDown(1)
				return m, nil
			}
		}
		if !m.loading && len(m.table.files) > 0 && launchPreviewScrollable(m) && isTabKey(msg) {
			m.preview.focused = true
			m.table.tbl.Blur()
			m = m.applyMainPaneFocusStyles()
			return m, nil
		}
		var cmd tea.Cmd
		m.table.tbl, cmd = m.table.tbl.Update(msg)
		m = m.withLaunchPreviewSynced()
		return m, cmd
	}

	var cmd tea.Cmd
	if m.params.open && m.params.editKind != paramEditNone {
		m.params.editInput, cmd = m.params.editInput.Update(msg)
		return m, cmd
	}
	if m.rc.open {
		m.rc.inputs[m.rc.focus], cmd = m.rc.inputs[m.rc.focus].Update(msg)
		return m, cmd
	}
	m.table.tbl, cmd = m.table.tbl.Update(msg)
	return m, cmd
}

// updateServerSplitTableKeys handles keys when the split-pane server is running
// and focus is on the model table (navigation, refresh, horizontal scroll, etc.).
func (m Model) updateServerSplitTableKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.RescanModels) {
		if m.loading {
			return m, nil
		}
		if m.server.running && !m.server.exited {
			m = m.withLastRunError("Stop the server before re-scanning models.")
			return m, nil
		}
		m.loading = true
		m.loadErr = nil
		m = m.withLastRunCleared()
		return m, rescanModelsCmd()
	}
	if key.Matches(msg, m.keys.Refresh) {
		if m.loading {
			return m, nil
		}
		if m.server.running && !m.server.exited {
			m = m.withLastRunError("Stop the server before reloading runtime.")
			return m, nil
		}
		m = m.withLastRunCleared()
		return m, reloadRuntimeCmd()
	}
	if key.Matches(msg, m.keys.ConfigPort) {
		return m.openRuntimeConfig()
	}
	if key.Matches(msg, m.keys.Parameters) {
		if m.loading {
			m = m.withLastRunError("Wait for the model scan to finish.")
			return m, nil
		}
		return m.openParamPanel()
	}
	if key.Matches(msg, m.keys.ToggleTheme) {
		var cmd tea.Cmd
		m, cmd = m.cycleTheme()
		return m, cmd
	}
	if runServerKeyMode(msg) != runServerModeNone {
		if m.server.exited {
			m = m.withLastRunError("Dismiss the log (enter, esc, or q) before starting another.")
		} else {
			m = m.withLastRunError("Stop the server (esc or q) before starting another.")
		}
		return m, nil
	}
	if key.Matches(msg, m.keys.ScrollLeft) {
		m.table.hscroll.ScrollLeft(hScrollStep)
		return m, nil
	}
	if key.Matches(msg, m.keys.ScrollRight) {
		m.table.hscroll.ScrollRight(hScrollStep)
		return m, nil
	}
	if key.Matches(msg, m.keys.CopyPath) {
		m = copyLaunchCommandToClipboard(m)
		return m, nil
	}
	if key.Matches(msg, m.keys.SortColumn) {
		if m.loading || len(m.table.files) == 0 {
			return m, nil
		}
		sel := m.SelectedPath()
		m.table.sortCol = (m.table.sortCol + 1) % tableSortColCount
		m = m.applyTableSort(sel)
		return m, nil
	}
	if key.Matches(msg, m.keys.SortReverse) {
		if m.loading || len(m.table.files) == 0 {
			return m, nil
		}
		sel := m.SelectedPath()
		m.table.sortDesc = !m.table.sortDesc
		m = m.applyTableSort(sel)
		return m, nil
	}
	var cmd tea.Cmd
	m.table.tbl, cmd = m.table.tbl.Update(msg)
	m = m.withLaunchPreviewSynced()
	return m, cmd
}

// copyLaunchCommandToClipboard writes the launch preview command and sets lastRunNote feedback.
func copyLaunchCommandToClipboard(m Model) Model {
	cmd := launchPreviewCommandLine(m)
	if cmd == "" {
		return m.withLastRunError(CopyCommandFeedbackFailure)
	}
	if err := clipboard.WriteAll(cmd); err != nil {
		return m.withLastRunError(CopyCommandFeedbackFailure)
	}
	return m.withLastRunSuccess(CopyCommandFeedbackSuccess)
}
