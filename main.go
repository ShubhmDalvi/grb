package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
    
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

// ------------------ GLOBALS ------------------

var db *bbolt.DB

// DB schema: text|tag|alias|pinned|useCount|createdAt

// ------------------ DB PATH ------------------

func getDBPath() string {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		return filepath.Join(appdata, "grb", "grb.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".grb", "grb.db")
}

func initDB() {
	dbPath := getDBPath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	var err error
	db, err = bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("snippets"))
		return err
	})
}

// ------------------ TABLE HELPER ------------------

// stripAnsi removes ANSI color codes to get the actual text length
func stripAnsi(str string) string {
	result := ""
	inEscape := false
	for _, char := range str {
		if char == '\033' {
			inEscape = true
		} else if inEscape && char == 'm' {
			inEscape = false
		} else if !inEscape {
			result += string(char)
		}
	}
	return result
}

// getDisplayWidth returns the actual display width of a string (ignoring ANSI codes)
func getDisplayWidth(str string) int {
	return utf8.RuneCountInString(stripAnsi(str))
}

// padRight pads a string to a specific width, accounting for ANSI codes
func padRight(str string, width int) string {
	displayWidth := getDisplayWidth(str)
	if displayWidth >= width {
		return str
	}
	return str + strings.Repeat(" ", width-displayWidth)
}

func printSnippetTable(rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths based on actual text (not ANSI codes)
	
	
	// Header
	fmt.Println("┌─────┬────────────────────────────────────────────────────┬──────────────────────┬──────────────────────┐")
	fmt.Printf("│ %-3s │ %-50s │ %-20s │ %-20s │\n", "ID", "Snippet", "Tag", "Alias")
	fmt.Println("├─────┼────────────────────────────────────────────────────┼──────────────────────┼──────────────────────┤")

	// Rows
	for _, row := range rows {
		if len(row) >= 4 {
			// Truncate long snippets
			snippet := row[1]
			if getDisplayWidth(snippet) > 50 {
				snippet = stripAnsi(snippet)[:47] + "..."
			}
			
			fmt.Printf("│ %s │ %s │ %s │ %s │\n",
				padRight(row[0], 3),
				padRight(snippet, 50),
				padRight(row[2], 20),
				padRight(row[3], 20))
		}
	}
	
	fmt.Println("└─────┴────────────────────────────────────────────────────┴──────────────────────┴──────────────────────┘")
}

// ------------------ MAIN ------------------

func main() {
	initDB()
	defer db.Close()

	rootCmd := &cobra.Command{
    Use:   "grb",
    Short: "grb - Smart Clipboard & Snippet Manager",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("💡 Tip: Run \"grb help\" to see all commands and features")
        launchTUI() // default = TUI
    },
}

rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
    cyan := color.New(color.FgCyan).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()

    fmt.Println("─────────────────────────────────────────────")
    fmt.Println("   grb (grab) - Smart Clipboard Manager")
    fmt.Println("─────────────────────────────────────────────\n")

    fmt.Println(cyan("📦 Features"))
    fmt.Println("─────────────────────────────────────────────")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Save snippets", yellow("grb save \"text\" --tag t --alias a"))
    fmt.Printf("%s %-22s %s\n", green("✔"), "Auto-copy on save", "(copies immediately to clipboard)")
    fmt.Printf("%s %-22s %s\n", green("✔"), "List all snippets", "grb list")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Search snippets", "grb search <word>")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Copy snippet", "grb copy <id|alias>")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Pin/Unpin snippet", "grb pin <id|alias>")
	fmt.Printf("%s %-22s %s\n", green("✔"), "Update alias", "grb alias <id|oldAlias> <newAlias>")
fmt.Printf("%s %-22s %s\n", green("✔"), "Delete snippet", "grb delete <id|alias>")
fmt.Printf("%s %-22s %s\n", green("✔"), "Clear snippets", "grb clear --all/--tag/--unpinned")

    fmt.Printf("%s %-22s %s\n", green("✔"), "Edit snippet", "grb edit <id|alias>")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Show usage stats", "grb stats")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Clipboard history", "grb daemon")
    fmt.Printf("%s %-22s %s\n", green("✔"), "Interactive TUI", "grb tui   (or just 'grb')")

    fmt.Println("\n📋 Notes")
    fmt.Println("─────────────────────────────────────────────")
    fmt.Println("Data is stored persistently at:")
    fmt.Println("   %APPDATA%\\grb   (Windows)")
    fmt.Println("   ~/.grb          (Linux/Mac)")

    fmt.Println("\n💡 Tip: Run 'grb' with no command to launch TUI.\n")
})

	// ------------------ SAVE ------------------
	saveCmd := &cobra.Command{
		Use:   "save [text]",
		Short: "Save a snippet (auto copies too)",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Provide text to save")
				return
			}
			text := strings.Join(args, " ")
			tag, _ := cmd.Flags().GetString("tag")
			alias, _ := cmd.Flags().GetString("alias")
			saveSnippet(text, tag, alias)
		},
	}
	saveCmd.Flags().String("tag", "", "Add a tag")
	saveCmd.Flags().String("alias", "", "Give an alias")
	rootCmd.AddCommand(saveCmd)

	// ------------------ LIST ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List snippets",
		Run: func(cmd *cobra.Command, args []string) {
			listSnippets()
		},
	})

	// ------------------ DELETE ------------------
rootCmd.AddCommand(&cobra.Command{
    Use:   "delete [id|alias]",
    Short: "Delete a snippet",
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) == 0 {
            fmt.Println("Provide snippet id or alias to delete")
            return
        }
        deleteSnippet(args[0])
    },
})

// ------------------ CLEAR ------------------
clearCmd := &cobra.Command{
    Use:   "clear",
    Short: "Clear snippets (dangerous!)",
}
clearCmd.Flags().Bool("all", false, "Delete all snippets")
clearCmd.Flags().String("tag", "", "Delete all snippets with a tag")
clearCmd.Flags().Bool("unpinned", false, "Delete all unpinned snippets")
clearCmd.Run = func(cmd *cobra.Command, args []string) {
    all, _ := cmd.Flags().GetBool("all")
    tag, _ := cmd.Flags().GetString("tag")
    unpinned, _ := cmd.Flags().GetBool("unpinned")
    clearSnippets(all, tag, unpinned)
}
rootCmd.AddCommand(clearCmd)

	// ------------------ SEARCH ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [query]",
		Short: "Search snippets",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Provide a search term")
				return
			}
			searchSnippets(args[0])
		},
	})

	// ------------------ COPY ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "copy [id|alias]",
		Short: "Copy snippet to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Provide snippet id or alias")
				return
			}
			copySnippet(args[0])
		},
	})
	
	// ✅ Add this
	rootCmd.AddCommand(&cobra.Command{
    Use:   "tui",
    Short: "Launch interactive TUI mode",
    Run: func(cmd *cobra.Command, args []string) {
        launchTUI()
    },
})

	// ------------------ PIN ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "pin [id|alias]",
		Short: "Pin/unpin a snippet",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Provide snippet id or alias")
				return
			}
			pinSnippet(args[0])
		},
	})

	// ------------------ EDIT ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "edit [id|alias]",
		Short: "Edit a snippet in default editor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Provide snippet id or alias to edit")
				return
			}
			editSnippet(args[0])
		},
	})

	// ------------------ ALIAS ------------------
	rootCmd.AddCommand(&cobra.Command{
    Use:   "alias [id|oldAlias] [newAlias]",
    Short: "Update alias for a snippet",
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) < 2 {
            fmt.Println("Usage: grb alias [id|oldAlias] [newAlias]")
            return
        }
        updateAlias(args[0], args[1])
    },
})

	// ------------------ STATS ------------------
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show snippet usage stats",
		Run: func(cmd *cobra.Command, args []string) {
			showStats()
		},
	})

	// ------------------ DAEMON ------------------
	    rootCmd.AddCommand(&cobra.Command{
        Use:   "daemon",
        Short: "Run clipboard watcher (history mode)",
        Run: func(cmd *cobra.Command, args []string) {
            startDaemon()
        },
    })

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

// ------------------ TUI ------------------

type item struct {
    id      string
    text    string
    tag     string
    alias   string
    pin     string
    section string // "header", "snippet"
}

func (i item) Title() string {
    if i.section == "header" {
        return color.CyanString(i.text)
    }
    if i.pin == "true" {
        return color.YellowString("📌 %s", i.text)
    }
    return i.text
}

func (i item) Description() string {
    if i.section == "header" {
        return ""
    }
    desc := ""
    if i.tag != "" {
        desc += color.MagentaString("🏷 %s  ", i.tag)
    }
    if i.alias != "" {
        desc += color.YellowString("📖 %s", i.alias)
    }
    return desc
}

func (i item) FilterValue() string {
    return i.text + " " + i.tag + " " + i.alias
}

type model struct {
    list list.Model
}

func newModel(snippets []item) model {
    items := make([]list.Item, len(snippets))
    for i, s := range snippets {
        items[i] = s
    }

    l := list.New(items, list.NewDefaultDelegate(), 80, 20)
    l.Title = "📋 grb - Smart Clipboard Manager"
    l.SetShowStatusBar(false)
    l.SetShowHelp(false) // we'll use footer

    return model{list: l}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
    if i, ok := m.list.SelectedItem().(item); ok {
        if i.section == "header" {
            return m, nil
        }
        clipboard.WriteAll(i.text)
        m.list.NewStatusMessage(color.GreenString("✅ Copied: %s", i.text))
        // do NOT quit, just keep browsing
        return m, nil
    }

        case "q", "esc":
            return m, tea.Quit
        }
    }
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.list.View() + "\n" +
        color.CyanString("↑/↓ move") + " | " +
        color.GreenString("Enter copy") + " | " +
        color.YellowString("q quit")
}

func launchTUI() {
    var pinned []item
    var others []item

    db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")
            pinnedFlag := "false"
            if len(fields) > 3 {
                pinnedFlag = fields[3]
            }

            itm := item{
                id:      string(k),
                text:    fields[0],
                tag:     fields[1],
                alias:   fields[2],
                pin:     pinnedFlag,
                section: "snippet",
            }

            if pinnedFlag == "true" {
                pinned = append(pinned, itm)
            } else {
                others = append(others, itm)
            }
        }
        return nil
    })

    var snippets []item
    if len(pinned) > 0 {
        snippets = append(snippets, item{text: "📌 Pinned", section: "header"})
        snippets = append(snippets, pinned...)
    }
    if len(others) > 0 {
        snippets = append(snippets, item{text: "Others", section: "header"})
        snippets = append(snippets, others...)
    }

    p := tea.NewProgram(newModel(snippets))
    if err := p.Start(); err != nil {
        fmt.Println("Error running TUI:", err)
    }
}

// ------------------ SAVE SNIPPET ------------------
func saveSnippet(text, tag, alias string) {
    var newID string

    err := db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        id, _ := b.NextSequence()
        newID = fmt.Sprintf("%d", id)

        val := fmt.Sprintf("%s|%s|%s|%t|%d|%d",
            text, tag, alias, false, 0, time.Now().Unix())

        return b.Put([]byte(newID), []byte(val))
    })
    if err != nil {
        log.Fatal(err)
    }

    // Copy immediately
    clipboard.WriteAll(text)

    // Colors
    cyan := color.New(color.FgCyan).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    magenta := color.New(color.FgMagenta).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()

    // Polished output
    fmt.Printf("%s Saved snippet [%s]\n", green("✅"), cyan(newID))
    
    // Use custom table formatting
    printSnippetTable([][]string{
        {cyan(newID), text, magenta(tag), yellow(alias)},
    })
    
    fmt.Println("📋 Copied to clipboard!")
    fmt.Println("💡 Tip: Run 'grb list' to view snippets")
}

// ------------------ LIST SNIPPETS ------------------

func listSnippets() {
    total := 0
    pinnedRows := [][]string{}
    otherRows := [][]string{}

    db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            total++
            fields := strings.Split(string(v), "|")

            id := string(k)
            text := fields[0]
            tag := "-"
            alias := "-"
            pinnedFlag := "false"

            if len(fields) > 1 && fields[1] != "" {
                tag = fields[1]
            }
            if len(fields) > 2 && fields[2] != "" {
                alias = fields[2]
            }
            if len(fields) > 3 {
                pinnedFlag = fields[3]
            }

            cyan := color.New(color.FgCyan).SprintFunc()
            yellow := color.New(color.FgYellow).SprintFunc()
            magenta := color.New(color.FgMagenta).SprintFunc()

            row := []string{cyan(id), text, magenta(tag), yellow(alias)}

            if pinnedFlag == "true" {
                pinnedRows = append(pinnedRows, row)
            } else {
                otherRows = append(otherRows, row)
            }
        }
        return nil
    })

    cyan := color.New(color.FgCyan).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()

    fmt.Println("─────────────────────────────────────────────")
    fmt.Printf("%s (total: %d)\n", cyan("📋 Saved Snippets"), total)
    fmt.Println("─────────────────────────────────────────────")

    // Pinned section
    if len(pinnedRows) > 0 {
        fmt.Println(yellow("📌 Pinned"))
        printSnippetTable(pinnedRows)
        fmt.Println()
    }

    // Others section
    if len(otherRows) > 0 {
        fmt.Println(cyan("Others"))
        printSnippetTable(otherRows)
        fmt.Println()
    }

    if total == 0 {
        color.Yellow("⚠ No snippets found.")
        fmt.Println("💡 Tip: Use 'grb save \"text\"' to create your first snippet")
    } else {
        fmt.Println("💡 Tip: Use 'grb search <word>' to filter, or 'grb tui' for interactive mode.")
    }
}

// ------------------ SEARCH ------------------

func searchSnippets(query string) {
    resultsPinned := [][]string{}
    resultsOthers := [][]string{}

    db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")

            id := string(k)
            text := fields[0]
            tag := "-"
            alias := "-"
            pinnedFlag := "false"

            if len(fields) > 1 && fields[1] != "" {
                tag = fields[1]
            }
            if len(fields) > 2 && fields[2] != "" {
                alias = fields[2]
            }
            if len(fields) > 3 {
                pinnedFlag = fields[3]
            }

            // Search match
            if strings.Contains(strings.ToLower(text), strings.ToLower(query)) ||
                strings.Contains(strings.ToLower(tag), strings.ToLower(query)) ||
                strings.Contains(strings.ToLower(alias), strings.ToLower(query)) {

                cyan := color.New(color.FgCyan).SprintFunc()
                yellow := color.New(color.FgYellow).SprintFunc()
                magenta := color.New(color.FgMagenta).SprintFunc()

                row := []string{cyan(id), text, magenta(tag), yellow(alias)}
                if pinnedFlag == "true" {
                    resultsPinned = append(resultsPinned, row)
                } else {
                    resultsOthers = append(resultsOthers, row)
                }
            }
        }
        return nil
    })

    cyan := color.New(color.FgCyan).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()

    if len(resultsPinned)+len(resultsOthers) == 0 {
        color.Yellow("⚠ No snippets found for \"%s\"", query)
        fmt.Println("💡 Tip: Use 'grb list' to see all snippets")
        return
    }

    fmt.Println("─────────────────────────────────────────────")
    fmt.Printf("%s \"%s\"\n", cyan("🔍 Search Results for:"), query)
    fmt.Println("─────────────────────────────────────────────")

    // Pinned
    if len(resultsPinned) > 0 {
        fmt.Println(yellow("📌 Pinned"))
        printSnippetTable(resultsPinned)
        fmt.Println()
    }

    // Others
    if len(resultsOthers) > 0 {
        fmt.Println(cyan("Others"))
        printSnippetTable(resultsOthers)
        fmt.Println()
    }

    fmt.Println("💡 Tip: Use 'grb copy <id|alias>' to reuse a snippet")
}

// ------------------ COPY ------------------

func copySnippet(idOrAlias string) {
    found := false

    db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")

            id := string(k)
            text := fields[0]
            tag := "-"
            alias := "-"
            useCount := "0"

            if len(fields) > 1 && fields[1] != "" {
                tag = fields[1]
            }
            if len(fields) > 2 && fields[2] != "" {
                alias = fields[2]
            }
            if len(fields) > 4 {
                useCount = fields[4]
            }

            if id == idOrAlias || alias == idOrAlias {
                // Copy to clipboard
                clipboard.WriteAll(text)
                found = true

                // Increment usage count
                count := 0
                fmt.Sscanf(useCount, "%d", &count)
                count++
                newVal := fmt.Sprintf("%s|%s|%s|%s|%d|%d",
                    text, tag, alias, fields[3], count, time.Now().Unix())
                b.Put(k, []byte(newVal))

                // Polished output
                cyan := color.New(color.FgCyan).SprintFunc()
                yellow := color.New(color.FgYellow).SprintFunc()
                magenta := color.New(color.FgMagenta).SprintFunc()
                green := color.New(color.FgGreen).SprintFunc()

                fmt.Println(green("✅ Copied snippet [" + id + "]"))
                
                printSnippetTable([][]string{
                    {cyan(id), text, magenta(tag), yellow(alias)},
                })
                
                fmt.Println("💡 Tip: Paste it anywhere with Ctrl+V")
                return nil
            }
        }
        return nil
    })

    if !found {
        color.Yellow("⚠ Snippet not found for \"%s\"", idOrAlias)
        fmt.Println("💡 Tip: Run 'grb list' to see available snippets")
    }
}

// ------------------ PIN TOGGLE ------------------

func pinSnippet(idOrAlias string) {
    found := false

    db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")

            id := string(k)
            text := fields[0]
            tag := "-"
            alias := "-"
            pinnedFlag := "false"

            if len(fields) > 1 && fields[1] != "" {
                tag = fields[1]
            }
            if len(fields) > 2 && fields[2] != "" {
                alias = fields[2]
            }
            if len(fields) > 3 {
                pinnedFlag = fields[3]
            }

            if id == idOrAlias || alias == idOrAlias {
                found = true

                // Toggle pin state
                newPinned := "true"
                action := "📌 Snippet pinned"
                if pinnedFlag == "true" {
                    newPinned = "false"
                    action = "📍 Snippet unpinned"
                }

                // Save updated snippet
                newVal := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
                    text,
                    tag,
                    alias,
                    newPinned,
                    fields[4],
                    time.Now().Unix(),
                )
                b.Put(k, []byte(newVal))

                // Polished output
                cyan := color.New(color.FgCyan).SprintFunc()
                yellow := color.New(color.FgYellow).SprintFunc()
                magenta := color.New(color.FgMagenta).SprintFunc()

                fmt.Printf("%s [%s]\n", action, cyan(id))
                
                printSnippetTable([][]string{
                    {cyan(id), text, magenta(tag), yellow(alias)},
                })

                if newPinned == "true" {
                    fmt.Println("💡 Tip: Run 'grb list' to see pinned snippets at the top")
                } else {
                    fmt.Println("💡 Tip: Run 'grb list' to see all snippets")
                }

                return nil
            }
        }
        return nil
    })

    if !found {
        color.Yellow("⚠ Snippet not found for \"%s\"", idOrAlias)
        fmt.Println("💡 Tip: Run 'grb list' to see available snippets")
    }
}

// ------------------ UPDATE ALIAS ------------------
func updateAlias(idOrAlias, newAlias string) {
    updated := false

    db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")
            id := string(k)
            alias := fields[2]

            if id == idOrAlias || alias == idOrAlias {
                fields[2] = newAlias
                newVal := strings.Join(fields, "|")
                b.Put(k, []byte(newVal))
                updated = true

                cyan := color.New(color.FgCyan).SprintFunc()
                yellow := color.New(color.FgYellow).SprintFunc()
                magenta := color.New(color.FgMagenta).SprintFunc()
                green := color.New(color.FgGreen).SprintFunc()

                fmt.Printf("%s Updated alias for snippet [%s]\n", green("✅"), cyan(id))
                
                printSnippetTable([][]string{
                    {cyan(id), fields[0], magenta(fields[1]), yellow(fields[2])},
                })
                
                fmt.Println("💡 Tip: Run 'grb list' to confirm changes")
                break
            }
        }
        return nil
    })

    if !updated {
        color.Yellow("⚠ Snippet not found for \"%s\"", idOrAlias)
        fmt.Println("💡 Tip: Run 'grb list' to see available snippets")
    }
}

// ------------------ DELETE ------------------

func deleteSnippet(idOrAlias string) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("snippets"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fields := strings.Split(string(v), "|")
			alias := fields[2]

			if string(k) == idOrAlias || alias == idOrAlias {
				b.Delete(k)
				color.Red("🗑 Snippet deleted!")
				return nil
			}
		}
		color.Yellow("⚠ Snippet not found!")
		return nil
	})
}

// ------------------ CLEAR ------------------

func clearSnippets(all bool, tag string, unpinned bool) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("snippets"))
		c := b.Cursor()
		deleted := 0

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fields := strings.Split(string(v), "|")
			text, t, alias, pinned := fields[0], fields[1], fields[2], fields[3]

			if all || (tag != "" && t == tag) || (unpinned && pinned == "false") {
				b.Delete(k)
				color.Red("🗑 Deleted [%s] %s (%s) %s", string(k), text, t, alias)
				deleted++
			}
		}

		if deleted == 0 {
			color.Yellow("⚠ No matching snippets found.")
		} else {
			color.Green("✅ %d snippet(s) deleted.", deleted)
		}
		return nil
	})
}

// ------------------ EDIT ------------------

func editSnippet(idOrAlias string) {
    tmpFile := filepath.Join(os.TempDir(), "grb_edit.txt")

    var key []byte
    var original []string
    var id string

    // Find snippet
    db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")
            alias := "-"
            if len(fields) > 2 && fields[2] != "" {
                alias = fields[2]
            }
            if string(k) == idOrAlias || alias == idOrAlias {
                key = append([]byte{}, k...)
                id = string(k)
                original = fields
                os.WriteFile(tmpFile, []byte(fields[0]), 0644)
            }
        }
        return nil
    })

    if key == nil {
        color.Yellow("⚠ Snippet not found for \"%s\"", idOrAlias)
        fmt.Println("💡 Tip: Run 'grb list' to see available snippets")
        return
    }

    // Capture old text before editing
    oldText := original[0]

    // Open in default editor
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.Command("notepad", tmpFile)
    } else {
        editor := os.Getenv("EDITOR")
        if editor == "" {
            editor = "nano"
        }
        cmd = exec.Command(editor, tmpFile)
    }
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

    // Read back and update DB
    edited, _ := os.ReadFile(tmpFile)
    newText := string(edited)

    db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        original[0] = newText
        newVal := strings.Join(original, "|")
        return b.Put(key, []byte(newVal))
    })

    // Polished output
    cyan := color.New(color.FgCyan).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    magenta := color.New(color.FgMagenta).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()

    fmt.Printf("%s Snippet [%s] updated\n", green("✅"), cyan(id))
    
    fmt.Println("Before")
    fmt.Println("─────────────────────────────────────────────")
    printSnippetTable([][]string{
        {cyan(id), oldText, magenta(original[1]), yellow(original[2])},
    })

    fmt.Println("\nAfter")
    fmt.Println("─────────────────────────────────────────────")
    printSnippetTable([][]string{
        {cyan(id), newText, magenta(original[1]), yellow(original[2])},
    })

    fmt.Println("💡 Tip: Run 'grb list' to confirm all snippets")
}

// ------------------ STATS ------------------

func showStats() {
    total := 0
    tagCount := map[string]int{}
    var topSnippet string
    maxCount := 0
    var topTag string
    maxTagCount := 0

    db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte("snippets"))
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            fields := strings.Split(string(v), "|")
            total++

            tag := ""
            count := 0

            if len(fields) > 1 && fields[1] != "" {
                tag = fields[1]
            }
            if len(fields) > 4 {
                fmt.Sscanf(fields[4], "%d", &count)
            }

            if tag != "" {
                tagCount[tag]++
                if tagCount[tag] > maxTagCount {
                    maxTagCount = tagCount[tag]
                    topTag = tag
                }
            }
            if count > maxCount {
                maxCount = count
                topSnippet = fmt.Sprintf("[%s] %s", string(k), fields[0])
            }
        }
        return nil
    })

    cyan := color.New(color.FgCyan).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    fmt.Println("─────────────────────────────────────────────")
    fmt.Println(cyan("📊 grb Stats"))
    fmt.Println("─────────────────────────────────────────────")
    fmt.Printf("%-18s : %s\n", "Total snippets", green(fmt.Sprintf("%d", total)))
    if topSnippet != "" {
        fmt.Printf("%-18s : %s (%s)\n", "Most used", yellow(topSnippet), red(fmt.Sprintf("🔥 %d times", maxCount)))
    }
    if topTag != "" {
        fmt.Printf("%-18s : 🏷 %s (%d snippets)\n", "Top tag", yellow(topTag), maxTagCount)
    }

    if len(tagCount) > 0 {
        fmt.Println("─────────────────────────────────────────────")
        fmt.Println(cyan("Tag Breakdown"))
        fmt.Println("─────────────────────────────────────────────")
        
        // Simple table for tag breakdown
        fmt.Println("┌──────────────────────┬───────┐")
        fmt.Printf("│ %-20s │ %-5s │\n", "Tag", "Count")
        fmt.Println("├──────────────────────┼───────┤")
        
        for t, c := range tagCount {
            tagDisplay := yellow("🏷 " + t)
            countDisplay := green(fmt.Sprintf("%d", c))
            fmt.Printf("│ %s │ %s │\n",
                padRight(tagDisplay, 20),
                padRight(countDisplay, 5))
        }
        fmt.Println("└──────────────────────┴───────┘")
    }
}

// ------------------ DAEMON ------------------

func startDaemon() {
    color.Yellow("📡 grb Daemon started. Watching clipboard...")
    fmt.Println("─────────────────────────────────────────────")

    last := ""

    for {
        text, _ := clipboard.ReadAll()
        if text != "" && text != last {
            var newID string

            db.Update(func(tx *bbolt.Tx) error {
                b := tx.Bucket([]byte("snippets"))
                id, _ := b.NextSequence()
                newID = fmt.Sprintf("%d", id)

                val := fmt.Sprintf("%s|%s|%s|%t|%d|%d",
                    text, "auto", "", false, 0, time.Now().Unix())

                return b.Put([]byte(newID), []byte(val))
            })

            // Colors
            cyan := color.New(color.FgCyan).SprintFunc()
            yellow := color.New(color.FgYellow).SprintFunc()
            magenta := color.New(color.FgMagenta).SprintFunc()
            green := color.New(color.FgGreen).SprintFunc()

            // Polished output
            fmt.Printf("\n%s snippet [%s]\n", green("✅ Captured"), cyan(newID))
            
            printSnippetTable([][]string{
                {cyan(newID), text, magenta("auto"), yellow("-")},
            })
            
            fmt.Println("💡 Tip: Press Ctrl+C to stop daemon")

            last = text
        }
        time.Sleep(1 * time.Second)
    }
}