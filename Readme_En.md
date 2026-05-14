# Evidence-Capture

## 🔥 Why I Built This

**"Liberating Engineers and Testers from 'Screenshot Prison'"**

In our world, there are many "harsh environments" where modern IT tools and cloud services are strictly forbidden. Medical institutions, educational boards, local governments, and legacy financial giants—these are "sanctuaries" (or prisons) where adopting new technology is viewed as a security risk.

Cloud services are labeled as "data leaks," macros (VBA) are suspected as "viruses," and trying to install Python requires six months of bureaucratic approval. Consequently, in fields where errors are least tolerated, the most primitive method—manual screenshot pasting—remains the only "safe" way to work.

**"If you can't use tools, become part of the environment."**

`Evidence-Capture` is a survival strategy for those fighting against irrational constraints. By utilizing primitive OS APIs and eliminating external dependencies, it operates as a "portable, zero-install" solution in even the most restrictive environments, providing a shelter where engineers can work like human beings again.

---

## 🛠 Target: The Ultimate "Worst-Case" Environment

This tool is designed to succeed in **extreme conditions** where standard IT survival is nearly impossible.

* **"Stone Age" Networking**: No internet (air-gapped), no LAN access, or all ports blocked on a standalone terminal.
* **"Iron-Clad" Permissions**: Restricted execution (even renaming extensions is blocked) and standard user privileges where writing to Temp folders is monitored.
* **"Starved" Resources**: Less than 4GB RAM and under 1GB disk space. Legacy hardware where modern browsers struggle to launch.
* **"Aggressive" Asset Management**: Coexistence with over-protective security agents that flag any external Office automation as "suspicious behavior."

---

## 🚀 Features Specialized for the "Abnormal" Frontline

### 1. "Pure Standalone" That Defies Constraints
* **Runtime-free Native Binary**: No dependency on .NET Framework versions or Java. A single `.exe` built in Go operates entirely using OS-native functions.
* **Zero Footprint**: No interference with the Registry or Environment Variables. Run it from a USB drive, and when you unplug it, it's as if it was never there.

### 2. Advanced Excel Control That "Negotiates" Stability
* **Bypassing "COM Hell"**: Precision detection of Excel's freezing triggers (e.g., cell editing mode, open dialogs). It waits safely and resumes only when Excel is ready.
* **Automatic Resource Cleanup**: Maintains memory consumption below 100MB even when pasting thousands of images. It keeps legacy PCs running smoothly and auto-deletes temporary files.

### 3. Robust Fail-safes: "No Capture Left Behind"
* **Multi-path Capture**: If Win32 APIs are blocked, it instantly switches to BitBlt or simulated PrintScreen commands to ensure evidence is saved.
* **Capturing the "Invisible"**: Capability to capture windows that are off-screen or hidden behind other windows.
* **Auto-save & Recovery**: Protects your data even during a sudden system shutdown, minimizing "re-work" time.

### 4. Integrity of Evidence
* **Auto-generation of Table of Contents**: Automatically generates an index with links to every sheet. Includes seamless status management (OK/NG) for all evidence.

---

## 💡 To Those Considering Implementation

If you believe your workplace is the "most restrictive environment in the country," this tool is for you.

`Evidence-Capture` survives the whims of the OS and Office, reclaiming your "precious life (time)" from the abyss of hopeless manual labor.

---

### 📝 Technical Specifications (Summary)

| Item | Specification |
| :--- | :--- |
| **Language** | Go (Direct Win32 API calls) |
| **Supported OS** | Windows 8 / 10 / 11 (32bit/64bit) |
| **Dependencies** | None (Single binary, No installation required) |
| **Recommended Specs** | RAM: 512MB+ / Disk: 100MB+ |
| **Office Integration** | Microsoft Excel 2013 or later recommended |

---

## 👨‍💻 For Developers: Build Instructions (OSS Version)

This repository contains the core capture and Excel control logic (commercial licensing and encryption logic excluded).

### Prerequisites
* Go (1.20+ recommended)
* `rsrc` (Resource embedding tool)
    (Install via: `go install github.com/akavel/rsrc@latest`)

### Build Steps
Double-click `build_oss.bat` or run the following in your command prompt:

```cmd
:: 1. Generate manifest and icon resources
rsrc -manifest main.manifest -ico "icon\favicon.ico" -o cmd\capture\rsrc.syso

:: 2. Build executable
go build -ldflags "-H windowsgui -s -w" -o Evidence-Capture-OSS.exe .\cmd\capture
