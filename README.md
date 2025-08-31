# grb (grab) 🚀
**Smart Clipboard & Snippet Manager (Terminal-based, written in Go)**

[![Go](https://img.shields.io/badge/Go-1.22-blue?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ✨ What is `grb`?
`grb` (short for **grab**) is a **terminal-based smart clipboard & snippet manager** for Windows.  
Unlike the default clipboard (which only stores the last copied item), `grb` lets you:

- Save, search, and reuse snippets  
- Tag, pin, and alias snippets  
- Launch an interactive TUI (like `fzf`)  
- Auto-copy on save  
- Edit snippets in your default editor  
- Track usage stats  
- Run as a clipboard daemon (auto-capture everything you copy)  

---

## 🔥 Features
- ✔ Save snippets with tags and aliases  
- ✔ Automatically copy on save  
- ✔ List all snippets with **colorful table output** (📌 pinned first)  
- ✔ Search snippets by text, tag, or alias  
- ✔ Copy snippets back into clipboard by ID or alias  
- ✔ Pin/unpin snippets for quick access  
- ✔ Edit snippets in your default editor (Notepad, Nano, etc.)  
- ✔ Show usage stats (most used snippets, tags, total count)  
- ✔ Clipboard daemon mode (capture everything you copy)  
- ✔ Interactive TUI mode (fuzzy search, arrow key navigation)  
- ✔ Data persists across restarts (stored in `%APPDATA%/grb`)  

---

## 📦 Installation

### 1. From Source (Go required)
```bash
git clone https://github.com/ShubhmDalvi/grb.git
cd grb
go mod tidy
go build -o grb.exe
```

Move `grb.exe` to a folder in your PATH for global usage:  
```bash
move grb.exe C:\Users\<YourName>\bin\
```
(then add that folder to **System PATH** if not already)

---

### 2. Run Locally
```bash
.\grb.exe help
```

---

## ⚡ Commands & Usage  

| Command | Example | Description |
|---------|----------|-------------|
| **Save a snippet** | `grb save "git push origin main" --tag git --alias push` | Saves a snippet with a tag and alias. Automatically copies it to clipboard. |
| **List snippets** | `grb list` | Lists all snippets in a table (📌 pinned appear first). |
| **Search snippets** | `grb search git` | Finds snippets by text, tag, or alias. |
| **Copy snippet** | `grb copy 3` <br> `grb copy push` | Copies snippet by ID or alias back into clipboard. |
| **Pin snippet** | `grb pin 3` | Pins snippet so it always shows at the top of list. |
| **Edit snippet** | `grb edit 3` | Opens snippet in your default editor (Notepad, Nano, etc.). |
| **Update alias** | `grb alias 3 deploy` | Updates alias of a snippet. |
| **Delete snippet** | `grb delete 3` | Deletes snippet by ID or alias. |
| **Clear snippets** | `grb clear --all` <br> `grb clear --tag git` <br> `grb clear --unpinned` | Deletes all snippets, by tag, or only unpinned ones. |
| **Stats** | `grb stats` | Shows usage stats: total snippets, most used, top tags. |
| **Daemon mode** | `grb daemon` | Runs in background and auto-saves every copied text. |
| **Interactive TUI** | `grb` | Launches full-screen fuzzy search UI (like `fzf`). |
| **Help** | `grb help` | Shows all available commands and examples. |

---

## 📊 Example Output  

```bash
✔ Saved snippet [3]
─────────────────────────────────────────────
ID   Snippet                     Tag      Alias
3    git push origin main        git      push
─────────────────────────────────────────────
📋 Copied to clipboard!
💡 Tip: Run 'grb list' to view snippets
```

---

## 🛠 Development
If you want to contribute:  
```bash
git clone https://github.com/ShubhmDalvi/grb.git
cd grb
go run grb_main_fixed.go
```

---

## 📄 License
MIT License © 2025 [ShubhmDalvi](https://github.com/ShubhmDalvi)
