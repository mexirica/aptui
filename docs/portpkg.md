# Export & Import Packages

**aptui** lets you export your installed packages to a JSON file and import them on another machine (or after a fresh install) to quickly restore your environment.

---

## Export

### Export all installed packages (`E`)

Press **`E`** to export **all** installed packages to a JSON file.

### Export manually installed packages (`M`)

Press **`M`** to export only **manually installed** packages (excluding auto-installed dependencies). This uses `apt-mark showmanual` to determine which packages were explicitly installed by the user.

> **Tip:** Exporting only manual packages produces a smaller, cleaner list that is usually sufficient to reproduce your setup — dependencies will be installed automatically.

### Overwrite confirmation

If the export file already exists, APTUI will ask for confirmation. Press the same key again (`E` or `M`) to overwrite, or `Esc` to cancel.

---

## Import (`I`)

1. Press **`I`** to start the import flow.
2. Enter the file path when prompted (leave empty to use the default path).
   - Supports `~/` expansion for home directory paths.
3. APTUI reads the JSON file and filters out packages that are already installed.
4. A confirmation overlay appears showing the number of packages to install and the file path.
5. Choose an action:

| Key | Action |
|-----|--------|
| `y` | Confirm and install all listed packages |
| `n` / `esc` | Cancel import |
| `d` | Toggle detail view (paginated list of packages to install) |
| `←` / `h` | Previous page (in detail view) |
| `→` / `l` | Next page (in detail view, 15 packages per page) |

---

## File format

The export file uses a simple JSON format:

```json
{
  "packages": [
    {"name": "curl"},
    {"name": "git"},
    {"name": "vim"}
  ]
}
```

Packages are sorted **alphabetically** on export.

---

## File location

| | Path |
|---|---|
| Default export/import path | `~/aptui-packages.json` |

When running under `sudo`, APTUI resolves the real user's home directory via the `SUDO_USER` environment variable.
