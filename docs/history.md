# Transaction History

**aptui** records every package operation and lets you undo or redo past transactions.

<p align="center">
    <img src="../assets/transaction.png" alt="Transaction history view" width="900" />
</p>

---

## Opening the transaction view

Press **`t`** on the main screen to switch to the Transactions tab, or click the **⟳ Transactions** tab header.

## Controls

| Key       | Action                                  |
|-----------|-----------------------------------------|
| `t`       | Open transaction history                |
| `z`       | Undo selected transaction               |
| `x`       | Redo selected transaction               |
| `↑` / `k` | Move selection up                      |
| `↓` / `j` | Move selection down                    |
| `pgup` / `ctrl+u` | Page up                        |
| `pgdown` / `ctrl+d` | Page down                    |
| `tab`     | Switch to another tab                   |

---

## What is recorded

The following operations are saved to the transaction history:

| Operation | Description |
|-----------|-------------|
| `install` | Package installation |
| `remove` | Package removal |
| `purge` | Package purge (remove + config files) |
| `upgrade` | Individual package upgrade |
| `upgrade-all` | Full system upgrade (`dist-upgrade`) |

Operations that are **not recorded**: `update` (index refresh), `cleanup-all` (autoremove), PPA add/remove.

---

## Transaction details

Each transaction entry contains:

| Field | Description |
|-------|-------------|
| ID | Auto-incremented identifier |
| Operation | The type of operation performed |
| Packages | List of packages affected |
| Date | Timestamp of when the operation was executed |
| Status | Success or failure |

When a transaction is selected, the detail panel shows the full package list and dependencies (loaded via `apt-cache depends`).

---

## Undo / Redo rules

### Undo (`z`)

Reverses the selected transaction:

| Original operation | Undo action |
|--------------------|-------------|
| `install` | Removes the installed packages |
| `remove` | Reinstalls the removed packages |
| `purge` | Reinstalls the purged packages |

**Restrictions:**
- **Upgrades cannot be undone** — downgrading is not supported.
- **Failed transactions cannot be undone** — only successful operations are reversible.
- **Essential packages are protected** — undo will not remove essential system packages.

### Redo (`x`)

Re-executes the original operation of the selected transaction.

---

## Storage

Transaction history is stored in:

```
~/.local/share/aptui/history.json
```

When running under `sudo`, APTUI resolves the real user's home directory via the `SUDO_USER` environment variable, so data is stored in the correct location.
