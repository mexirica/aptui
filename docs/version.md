# Version Selection & Downgrade

**aptui** lets you view all available versions of a package and install any of them — including older versions (downgrade).

---

## Opening the version selector

Press **`v`** on any package in the list. APTUI queries `apt-cache policy` to load all available versions and opens an overlay listing them from newest to oldest.

## Controls

| Key              | Action                          |
|------------------|---------------------------------|
| `v`              | Open version selector           |
| `↑` / `k`       | Move selection up               |
| `↓` / `j`       | Move selection down             |
| `pgup` / `ctrl+u` | Page up                       |
| `pgdown` / `ctrl+d` | Page down                   |
| `enter`          | Install selected version        |
| `esc`            | Cancel and return to package list |

---

## Version list

Each entry in the version list shows:

| Column | Description |
|--------|-------------|
| Status | `●` if this version is currently installed, blank otherwise |
| Version | The version string |
| Origin | Repository origin (visible when terminal width > 60) |

The currently installed version is highlighted, making it easy to identify which version you're running.

---

## Downgrading

When you select a version older than the one currently installed, APTUI automatically detects it as a downgrade. The operation:

- Uses `apt-get install <package>=<version> --allow-downgrades` to install the target version
- Records both the original version (`FromVersion`) and the target version (`ToVersion`) in the transaction history
- Shows "Downgrading" instead of "Installing" in the status bar

---

## Undo / Redo

Version changes (upgrades and downgrades) are fully integrated with the transaction history:

- **Undo (`z`)** — reverts the package to its previous version
- **Redo (`x`)** — re-applies the version change

See [Transaction History](history.md) for more details.
