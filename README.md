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

This creates a `grb.exe` in the folder.

Move it to your PATH for global usage:
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

## ⚡ Usage Examples

### Save a snippet
```bash
grb save "git push origin main" --tag git --alias push
```

### List snippets
```bash
grb list
```

### Search
```bash
grb search git
```

### Copy snippet
```bash
grb copy 3
grb copy push
```

### Pin & Edit
```bash
grb pin 3
grb edit 3
```

### Stats
```bash
grb stats
```

### Daemon Mode
```bash
grb daemon
```

### Interactive TUI
```bash
grb
```

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
