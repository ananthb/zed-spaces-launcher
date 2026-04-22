// Cosmonaut unified window — native Fyne implementation of the redesign.
//
// Layout:
//
//   ┌────────────┬──────────────────────────────────────┐
//   │  Sidebar   │  Detail panel                        │
//   │  (logo,    │  (codespace detail / create / repo)  │
//   │   search,  │                                      │
//   │   tree,    │                                      │
//   │   account) │                                      │
//   └────────────┴──────────────────────────────────────┘
//
// The bulk of structure mirrors the existing gui_window.go; this file
// rewrites the visual wrapping (title bar, search chrome, captions,
// action buttons) to match the design system.
package daemon

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"image/color"

	"github.com/ananth/cosmonaut/internal/codespace"
)

const (
	cosmoWinW float32 = 820
	cosmoWinH float32 = 520
)

// newCosmoWindow replaces newUnifiedWindow. Swap the call site in
// gui_flow.go's showGUI() to use this constructor.
func (d *Daemon) newCosmoWindow() *unifiedWindow {
	win := d.app.NewWindow("Cosmonaut")
	win.Resize(fyne.NewSize(cosmoWinW, cosmoWinH))
	win.CenterOnScreen()

	uw := &unifiedWindow{
		daemon:  d,
		win:     win,
		content: container.NewStack(),
	}
	uw.loadRepos()

	// Background fetch of all user repos.
	go func() {
		allUserRepos, err := codespace.ListAllRepos(d.Runner)
		if err != nil {
			log.Printf("gui: fetch repos: %v", err)
			return
		}
		fyne.Do(func() {
			uw.allRepos = mergeRepos(uw.allRepos, allUserRepos)
			uw.applyFilter()
			uw.tree.Refresh()
		})
	}()

	sidebar := uw.buildCosmoSidebar()
	uw.showCosmoWelcome()

	split := container.NewHSplit(sidebar, uw.content)
	split.Offset = 0.32
	win.SetContent(split)
	return uw
}

// buildCosmoSidebar constructs the left pane with title row, search,
// repo tree, and account footer. Separator canvases give crisp 1px lines
// that respect the theme's border color.
func (uw *unifiedWindow) buildCosmoSidebar() fyne.CanvasObject {
	// Title row: mark + name + "+" action
	mark := canvas.NewImageFromResource(markIconResource())
	mark.SetMinSize(fyne.NewSize(22, 22))
	mark.FillMode = canvas.ImageFillContain

	title := canvas.NewText("Cosmonaut", cText)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 13

	newBtn := widget.NewButtonWithIcon("", widget.NewIcon(nil).Resource, func() {
		uw.showCreateNewGeneric()
	})
	newBtn.Importance = widget.LowImportance

	titleRow := container.NewBorder(nil, nil, container.NewHBox(mark, title), newBtn)

	// Search
	filterEntry := widget.NewEntry()
	filterEntry.PlaceHolder = "Filter repositories…"
	filterEntry.OnChanged = func(text string) {
		uw.filter = text
		uw.applyFilter()
		uw.tree.Refresh()
	}

	// Tree — reuse structure from gui_window.go but override selection callbacks.
	uw.tree = uw.buildTree()
	uw.tree.OnSelected = func(id widget.TreeNodeID) {
		if isRepoNode(id) {
			repo := repoFromNode(id)
			// Auto-expand repos that have codespaces.
			if len(codespace.FilterByRepo(uw.daemon.Codespaces(), repo)) > 0 {
				uw.tree.OpenBranch(id)
			}
			uw.showCosmoRepoSummary(repo)
		} else if isCsNode(id) {
			csName := csNameFromNode(id)
			repo := repoFromCsNode(id)
			uw.showCosmoCodespaceDetail(csName, repo)
		} else if isNewNode(id) {
			repo := repoFromNewNode(id)
			uw.showCosmoCreateNew(repo)
		}
	}

	// Account footer
	account := uw.buildAccountFooter()

	top := container.NewVBox(
		container.NewPadded(titleRow),
		container.NewPadded(filterEntry),
		thinDivider(),
	)

	bottom := container.NewVBox(
		thinDivider(),
		account,
	)

	return container.NewBorder(top, bottom, nil, nil, uw.tree)
}

// buildAccountFooter shows the signed-in GitHub handle with a small status dot.
func (uw *unifiedWindow) buildAccountFooter() fyne.CanvasObject {
	// Try to get the GitHub username.
	ghUser := "not authenticated"
	authed := false
	if out, err := uw.daemon.Runner.Run([]string{"auth", "status", "--hostname", "github.com"}); err == nil {
		// Parse "Logged in to github.com account <user>" from output.
		for _, line := range strings.Split(out, "\n") {
			if idx := strings.Index(line, "account "); idx >= 0 {
				parts := strings.Fields(line[idx:])
				if len(parts) >= 2 {
					ghUser = parts[1]
					authed = true
				}
				break
			}
		}
		if !authed {
			ghUser = "authenticated"
			authed = true
		}
	}

	dot := stateDot(func() string {
		if authed {
			return "Available"
		}
		return "Stopped"
	}())

	handle := canvas.NewText(ghUser, cText)
	handle.TextSize = 12
	handle.TextStyle = fyne.TextStyle{Bold: true}

	sub := canvas.NewText("github.com", cTextMute)
	sub.TextSize = 10
	sub.TextStyle = fyne.TextStyle{Monospace: true}

	info := container.NewVBox(handle, sub)
	return container.NewPadded(
		container.NewHBox(dot, info),
	)
}

// thinDivider returns a 1px canvas line using the theme border color.
func thinDivider() fyne.CanvasObject {
	r := canvas.NewRectangle(cBorder)
	r.SetMinSize(fyne.NewSize(1, 1))
	return r
}

// markIconResource returns the Cosmonaut app mark (used in the sidebar
// header). Points at the embedded SVG; reuses the same asset as the
// dock icon.
func markIconResource() fyne.Resource {
	return fyne.NewStaticResource("mark.svg", iconAppSVG)
}

// ── CODESPACE DETAIL ────────────────────────────────────────────────────

func (uw *unifiedWindow) showCosmoCodespaceDetail(csName, repo string) {
	var cs *codespace.Codespace
	for _, c := range uw.daemon.Codespaces() {
		if c.Name == csName {
			cs = &c
			break
		}
	}
	if cs == nil {
		uw.showCosmoWelcome()
		return
	}

	target, resolvedName := guiTargetForRepo(uw.daemon.Cfg, repo)

	// Header: breadcrumbs + "⋯" overflow
	crumbs := canvas.NewText(fmt.Sprintf("%s  /  %s", repo, csLabel(*cs)), cTextDim)
	crumbs.TextSize = 12
	header := container.NewPadded(container.NewBorder(nil, nil, crumbs, nil))

	// Hero: state pill + title + primary/secondary actions
	stateLbl := canvas.NewText(strings.ToUpper(cs.State), stateColor(cs.State))
	stateLbl.TextSize = 10
	stateLbl.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	heroTitle := canvas.NewText(csLabel(*cs), cText)
	heroTitle.TextSize = 20
	heroTitle.TextStyle = fyne.TextStyle{Bold: true}

	heroName := canvas.NewText(cs.Name, cTextMute)
	heroName.TextSize = 11
	heroName.TextStyle = fyne.TextStyle{Monospace: true}

	statusRow := container.NewHBox(stateDot(cs.State), stateLbl)

	openBtn := primaryButton("Open in Zed", func() {
		uw.daemon.runLaunchFlow(uw.win, target, resolvedName, cs)
	})
	sshBtn := widget.NewButton("SSH…", func() {})
	deleteBtn := destructiveButton("Delete", func() {
		go func() {
			_ = codespace.DeleteCodespace(uw.daemon.Runner, cs.Name)
			fyne.Do(func() {
				uw.tree.Refresh()
				uw.showCosmoWelcome()
			})
		}()
	})

	heroLeft := container.NewVBox(statusRow, heroTitle, heroName)
	heroRight := container.NewVBox(
		openBtn,
		container.NewHBox(sshBtn, deleteBtn),
	)
	hero := surfaceCard(container.NewBorder(nil, nil, nil, heroRight, heroLeft))

	// Meta grid
	branchStr := ""
	if cs.GitStatus != nil {
		branchStr = cs.GitStatus.Ref
		if branchStr == "" {
			branchStr = cs.GitStatus.Branch
		}
	}
	meta := container.NewGridWithColumns(4,
		metaCell("BRANCH", branchStr, true),
		metaCell("MACHINE", "4-core · Linux", false),
		metaCell("REGION", "us-west-2", true),
		metaCell("LAST USED", "2 min ago", false),
	)

	// SSH + Git status split
	sshInfo := widget.NewLabel(fmt.Sprintf(
		"host  cs.%s.github.dev\nuser  codespace\nport  2222\npath  %s",
		cs.Name, target.WorkspacePath,
	))
	sshInfo.TextStyle = fyne.TextStyle{Monospace: true}
	sshCard := surfaceCard(container.NewVBox(
		caption("SSH CONNECTION"),
		sshInfo,
		container.NewHBox(
			widget.NewButton("Copy config", func() {}),
			widget.NewButton("Open in Terminal", func() {}),
		),
	))

	gitStatus := widget.NewLabel(fmt.Sprintf("%s · ahead 2, behind 0", branchStr))
	gitCard := surfaceCard(container.NewVBox(
		caption("GIT STATUS"),
		gitStatus,
	))

	panels := container.NewGridWithColumns(2, sshCard, gitCard)

	body := container.NewVBox(hero, meta, panels)
	uw.setContent(container.NewBorder(header, nil, nil, nil,
		container.NewPadded(container.NewVScroll(container.NewPadded(body))),
	))
}

func stateColor(state string) color.Color {
	switch state {
	case "Available":
		return cLime
	case "Starting":
		return cOrange
	case "Error":
		return cRed
	}
	return cTextMute
}

// ── WELCOME ─────────────────────────────────────────────────────────────

func (uw *unifiedWindow) showCosmoWelcome() {
	mark := canvas.NewImageFromResource(markIconResource())
	mark.SetMinSize(fyne.NewSize(56, 56))
	mark.FillMode = canvas.ImageFillContain

	h := canvas.NewText("Welcome to Cosmonaut", cText)
	h.TextSize = 16
	h.TextStyle = fyne.TextStyle{Bold: true}
	h.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("Select a repository or codespace to get started.", cTextMute)
	sub.TextSize = 12
	sub.Alignment = fyne.TextAlignCenter

	uw.setContent(container.NewCenter(container.NewVBox(
		container.NewCenter(mark),
		h, sub,
	)))
}

// ── REPO SUMMARY ───────────────────────────────────────────────────────

func (uw *unifiedWindow) showCosmoRepoSummary(repo string) {
	all := uw.daemon.Codespaces()
	repoCS := codespace.FilterByRepo(all, repo)

	title := canvas.NewText(repo, cText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}

	countText := fmt.Sprintf("%d codespace(s)", len(repoCS))
	info := canvas.NewText(countText, cTextDim)
	info.TextSize = 13

	createBtn := primaryButton("Create new codespace", func() {
		uw.showCosmoCreateNew(repo)
	})

	uw.setContent(container.NewCenter(container.NewVBox(
		title, info,
		widget.NewSeparator(),
		createBtn,
	)))
}

// ── CREATE ──────────────────────────────────────────────────────────────

func (uw *unifiedWindow) showCreateNewGeneric() {
	uw.showCosmoCreateNew("")
}

func (uw *unifiedWindow) showCosmoCreateNew(repo string) {
	target, resolvedName := guiTargetForRepo(uw.daemon.Cfg, repo)

	title := canvas.NewText("Create a new codespace", cText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}

	hint := canvas.NewText("A short label makes it easier to find later.", cTextMute)
	hint.TextSize = 12

	repoLbl := widget.NewLabel(repo)

	// Branch selector — starts with config branch or "main", fetches real branches async.
	defaultBranch := target.Branch
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	branchSel := widget.NewSelect([]string{defaultBranch}, func(string) {})
	branchSel.Selected = defaultBranch

	// Fetch branches in background.
	if repo != "" {
		go func() {
			branches := fetchBranches(uw.daemon.Runner, repo)
			if len(branches) > 0 {
				fyne.Do(func() {
					branchSel.Options = branches
					// Keep current selection if it's in the list.
					found := false
					for _, b := range branches {
						if b == branchSel.Selected {
							found = true
							break
						}
					}
					if !found {
						branchSel.Selected = branches[0]
					}
					branchSel.Refresh()
				})
			}
		}()
	}

	labelEntry := widget.NewEntry()
	labelEntry.PlaceHolder = "e.g. fix indexer health checks"

	form := widget.NewForm(
		widget.NewFormItem("Repository", repoLbl),
		widget.NewFormItem("Branch", branchSel),
		widget.NewFormItem("Label", labelEntry),
	)

	createBtn := primaryButton("Create and open", func() {
		text := strings.TrimSpace(labelEntry.Text)
		createTarget := target
		createTarget.DisplayName = text
		createTarget.Branch = branchSel.Selected
		uw.daemon.runCreateAndLaunch(uw.win, createTarget, resolvedName)
	})
	cancelBtn := widget.NewButton("Cancel", func() { uw.showCosmoWelcome() })

	actions := container.NewHBox(layout.NewSpacer(), cancelBtn, createBtn)

	body := container.NewPadded(container.NewVBox(
		title, hint,
		widget.NewSeparator(),
		form,
		actions,
	))
	uw.setContent(container.NewScroll(body))
}

// fetchBranches returns branch names for a repo, default branch first.
func fetchBranches(runner codespace.GHRunner, repo string) []string {
	// Get default branch.
	defOut, err := runner.Run([]string{
		"api", fmt.Sprintf("repos/%s", repo),
		"--jq", ".default_branch",
	})
	defaultBranch := strings.TrimSpace(defOut)
	if err != nil || defaultBranch == "" {
		defaultBranch = "main"
	}

	// Get all branches.
	out, err := runner.Run([]string{
		"api", fmt.Sprintf("repos/%s/branches", repo),
		"--paginate", "--jq", ".[].name",
	})
	if err != nil {
		return []string{defaultBranch}
	}

	var branches []string
	branches = append(branches, defaultBranch) // default first
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		b := strings.TrimSpace(line)
		if b != "" && b != defaultBranch {
			branches = append(branches, b)
		}
	}
	return branches
}

