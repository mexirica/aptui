# Mirror Fetch

**aptui** can automatically detect your Linux distribution, test available mirrors for latency, and apply the fastest ones to your APT sources.

<p align="center">
    <img src="../assets/mirror.gif" alt="Mirror testing" width="900" />
</p>

---

## Opening the mirror view

Press **`f`** on the main screen. APTUI will:

1. Detect your distribution from `/etc/os-release`
2. Fetch a list of available mirrors
3. Test each mirror's latency concurrently
4. Display results sorted by speed, with scores

## Controls

| Key       | Action                                     |
|-----------|--------------------------------------------|
| `f`       | Open mirror fetch view                     |
| `space`   | Toggle mirror selection                    |
| `enter`   | Apply selected mirrors                     |
| `esc`     | Cancel / close mirror view                 |
| `↑` / `k` | Move selection up                         |
| `↓` / `j` | Move selection down                       |
| `pgup` / `ctrl+u` | Page up                           |
| `pgdown` / `ctrl+d` | Page down                       |

---

## How it works

### Mirror discovery

| Distro family | Source |
|---------------|--------|
| **Ubuntu-based** (Ubuntu, Pop!_OS, Linux Mint, elementary, Zorin, KDE neon) | Launchpad RSS feed |
| **Debian-based** (Debian, Kali, MX Linux, antiX, Devuan) | debian.org mirror list |

Up to **50 mirrors** are sampled from the full list for testing.

### Latency testing

- All mirrors are tested **concurrently** (up to 25 at a time)
- Each mirror is tested with an HTTP HEAD request to `{url}/dists/`
- Timeout: **3 seconds** per mirror
- Mirrors that fail or time out are excluded from results

### Scoring

Mirrors are scored from **100** (fastest) downward based on relative latency. The top 3 fastest mirrors are automatically pre-selected.

---

## Applying mirrors

After selecting mirrors with `space`, press `enter` to apply. APTUI writes the selected mirrors to:

```
/etc/apt/sources.list.d/aptui-mirrors.list
```

The file includes entries for:
- Main repository
- Updates
- Security components

After writing, APTUI runs `apt-get update` to refresh the package indexes.

---

## Supported distributions

### Ubuntu-based
- Ubuntu
- Pop!_OS
- Linux Mint
- elementary OS
- Zorin OS
- KDE neon

### Debian-based
- Debian
- Kali Linux
- MX Linux
- antiX
- Devuan
