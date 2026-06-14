# TUI Design System — Terminal UI Aesthetic

A design reference for building interfaces that look and feel like native terminal/TUI applications — inspired by tools like `cpos`, `htop`, `lazygit`, `ncurses` apps, and retro CRT dashboards.

---

## Core Philosophy

> Monospace everything. Borders from box-drawing chars. Color sparingly. Density over decoration.

The TUI aesthetic is **utilitarian-retro**: every pixel earns its place. No rounded corners, no drop shadows, no gradients. Structure comes from ASCII/Unicode box-drawing, whitespace, and a strict typographic grid — not from visual chrome.

---

## Typography

```
Font family:    Monospace only
                "JetBrains Mono", "Fira Code", "Cascadia Code",
                "Hack", "Iosevka", "Courier New" (fallback)

Font sizes (viewport-scaled via root clamp):
  base          clamp(12px, 1.6vh, 28px) — root, everything uses rem

  xs            0.8rem  — status bars, footnotes, legends
  sm            0.85rem — keybind bar, muted labels
  md            0.95rem — body, table data, filter bar
  lg            1rem    — default UI text
  xl            1.8rem  — stat card numbers
  title         0.9rem  — panel titles (uppercase)
  display       3rem    — LED big stat numbers

  Reference:    1rem = 13px at 810px viewport height
                1rem = 17px at 1080p
                1rem = 23px at 1440p
                1rem = 28px at 4K (clamped)

Line height:    1.5 (body), 1 (display numbers), 1.3 (panel titles)
Letter spacing: 0 to 0.08em (tighter for body, wider for labels)
```

All font sizes use `rem` units so the entire UI scales proportionally with viewport height via the root `font-size: clamp(12px, 1.6vh, 28px)`.

**LED / Dot-matrix display text** — `font-family: var(--font-display); font-size: 3rem; line-height: 1`. For big numbers and hero titles, simulate a segmented display using a font like `"Share Tech Mono"`, `"Courier New"`, or CSS pixel-block rendering.

---

## Color Palette

### Base (Dark Terminal)

```
Background        #0d0d0f   — near-black, slightly warm
Surface           #111318   — panel backgrounds
Border            #2a2d3a   — box-drawing border color
Muted             #3a3d4a   — dimmed elements

Text primary      #c8ccd8   — default readable text
Text muted        #5a5f72   — comments, disabled
Text dim          #3f4252   — very low emphasis
```

### Accent Colors (use sparingly — 1–2 per UI)

```
Purple / Violet   #9d7dea   — primary brand, titles, active tabs
Cyan              #4ec9b0   — success, connected states
Yellow / Amber    #e5c07b   — streaks, warnings, highlights
Red / Coral       #e06c75   — errors, danger, 0% progress
Green             #98c379   — solved counts, online status
Blue              #61afef   — links, info
```

### Semantic Usage

| Element              | Color          |
|----------------------|----------------|
| Active/selected tab  | `#9d7dea` bg   |
| Solved / OK          | `#98c379`      |
| Streak / Special     | `#e5c07b`      |
| Rating / Score       | `#e5c07b`      |
| Error / 0%           | `#e06c75`      |
| Connected            | `#4ec9b0`      |
| Borders              | `#2a2d3a`      |
| Muted labels         | `#5a5f72`      |

---

## Box-Drawing & Borders

Use Unicode box-drawing characters for all structural borders.

```
Single line:    ─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼
Double line:    ═ ║ ╔ ╗ ╚ ╝ ╠ ╣ ╦ ╩ ╬
Mixed:          ╒ ╕ ╘ ╛ ╞ ╡ ╤ ╧ ╪
Rounded:        ╭ ╮ ╰ ╯  (used for panel focus rings)
Dashed:         ╌ ╎
```

**Panel with title:**
```
┌─ Section Title ──────────────────────┐
│  content here                        │
│  more content                        │
└──────────────────────────────────────┘
```

**Inline label badge:**
```
◆ BFL   ← diamond bullet + all-caps label
```

**Status indicators:**
```
✓   connected / success
✗   error / disconnected
●   active
○   inactive
▶   running
```

---

## Layout System

TUI layouts are **grid-based with fixed character columns**, not fluid. The entire UI scales proportionally to viewport height.

```
Max width:        full viewport (no cap), content centered
                  optional: max-width: min(86rem, 1100px) for constrained layout
Base font:        clamp(12px, 1.6vh, 28px) on <html>
Scaling:          pure CSS via rem/em units — no JS transforms

Outer padding:    0.5em 1em   (main-area)
Gap between panels: 1em       (screen-content gap)
Gap in stat row:  0.8em       (stat-4up gap)
Panel padding:    3.5em 1em 1em  (top, sides, bottom)
Panel title:      absolute at top: -0.9em, font-size: 0.9rem
Header bar:       padding: 0.5em 1.2em, margin: 0.7em 1em 0
Title bar:        padding: 0.55em 1em
Keybind bar:      padding: 0.45em 1em
Status bar:       padding: 0.3em 1em

Stat card:        padding: 1.5em 0.8em
  number:         font-size: 1.8rem, line-height: 1
  label:          font-size: 0.8rem, margin-top: 0.6em, uppercase
```

### Panel Types

**Stat Card (4-up row)**
```
┌──────────────────┐
│       48         │  ← large number, accent color
│    PROBLEMS      │  ← small uppercase label, muted
└──────────────────┘
```

**Progress bar (text-mode)**
```
data structures    ░░░░░░░░░░░░░  0%
graphs             ████████░░░░░  40%
bfs                ██████████░░░  50%
```
Use `█` and `░` (or dotted blocks `▓`) for fill. Color the fill bar with the relevant accent (yellow for mid, red for 0%).

**Table / List panel**
```
┌─ Recommended Next ─────────────────────┐
│  877D    Olya and Energy Drinks   1900  │
│  919D    Substring                1900  │
│  525D    Arthur and Walls         2000  │
└────────────────────────────────────────┘
```
Columns are space-aligned (not CSS flex). Problem codes in muted text, titles in primary, ratings in amber.

---

## Navigation / Tabs

```
[ Dashboard ]  Problems  Contests  Analytics  Recommend  Config
```

- Active tab: filled background (`#9d7dea`), dark text
- Inactive tabs: no background, muted text, hover → dim highlight
- No underlines, no borders on tabs — the filled bg is enough

---

## Iconography

No icon libraries. Use:

```
◆  ♦  ▸  ►  ✦   — decorative bullets / brand marks
●  ○  ◉            — status dots
✓  ✗              — boolean states
>                  — prompt / CLI prefix for status bar
·  –  —            — separators inline
```

---

## Animation & Motion

TUI UIs are **mostly static** — animations should feel like terminal output, not web transitions.

| Effect             | Implementation                          |
|--------------------|-----------------------------------------|
| Tab change         | Instant swap, no slide                  |
| Number updates     | Counter tick (increment per frame)      |
| Loading            | Spinner: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` cycling at ~80ms  |
| LED title          | Slow pulse opacity (2s ease-in-out)     |
| Progress bar fill  | Linear width transition, 600ms          |
| New row added      | Fade-in from 0 opacity, 200ms           |
| Cursor blink       | `_` blinking at 1s interval             |

---

## Status Bar

```css
/* Bottom/top single-line bar */
background: #0d0d0f;
border-top: 1px solid #2a2d3a;
font-size: 0.85rem;
color: #5a5f72;
padding: 0.3em 1em;
```

Content pattern:
```
> Press 'r' to sync with Codeforces and CSES
```
Always prefix with `> ` or `›` to evoke a terminal prompt.

---

## CSS Variables Template

```css
:root {
  /* Backgrounds */
  --bg:          #0d0d0f;
  --surface:     #111318;
  --surface-alt: #161820;
  --border:      #2a2d3a;
  --border-dim:  #1e2130;

  /* Text */
  --text:        #c8ccd8;
  --text-muted:  #5a5f72;
  --text-dim:    #3f4252;

  /* Accents */
  --purple:      #9d7dea;
  --cyan:        #4ec9b0;
  --yellow:      #e5c07b;
  --red:         #e06c75;
  --green:       #98c379;
  --blue:        #61afef;

  /* Typography */
  --font-mono:   "JetBrains Mono", "Fira Code", "Cascadia Code", monospace;
  --font-display:"Share Tech Mono", "Courier New", monospace;
}

/* Viewport-adaptive scaling — base font drives all rem/em sizes */
html { font-size: clamp(12px, 1.6vh, 28px); }
```

---

## Do's and Don'ts

### ✓ Do
- Use monospace font everywhere, no exceptions
- Keep borders thin (1px) and use box-drawing chars
- Use 4–6 accent colors max, applied semantically
- Align content to a character grid
- Use ALL CAPS for labels and section headers
- Keep density high — TUI UIs are information-dense
- Simulate CRT/terminal atmosphere with near-black bg

### ✗ Don't
- No rounded corners (border-radius: 0 always)
- No drop shadows or blur effects
- No gradients (except very subtle bg noise texture)
- No icon fonts or SVG icon libraries
- No sans-serif or display fonts
- No animations longer than 600ms
- No whitespace-heavy "breathing room" layouts

---

## Quick Reference: Key Patterns

```
Viewport scaling:  html { font-size: clamp(12px, 1.6vh, 28px); }
Big stat display:  font: var(--font-display); font-size: 1.8rem; line-height: 1; color: var(--yellow);
Panel:             border: 1px solid var(--border); padding: 3.5em 1em 1em;
Panel title:       position: absolute; top: -0.9em; font-size: 0.9rem; line-height: 1.3; text-transform: uppercase;
Active tab:        color: var(--purple); border-bottom: 2px solid var(--purple); font-weight: bold;
Muted label:       font-size: 0.8rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.08em;
Progress bar:      █ fill chars + ░ empty chars (colored spans over block-char background)
Status bar:        border-top: 1px solid var(--border); padding: 0.3em 1em; font-size: 0.85rem; color: var(--text-dim);
Keybind bar:       border-top: 1px solid var(--border); padding: 0.45em 1em; font-size: 0.85rem;
Header bar:        border: 1px solid var(--purple); padding: 0.5em 1.2em; margin: 0.7em 1em 0;
```

---

*This system is framework-agnostic — applies equally to React, HTML/CSS, Vue, or Svelte.*

---

## Screen-by-Screen Pattern Library

### Problems Screen

**Three-panel stacked layout** (Filter → Table → Details):
```
┌─ Filter ──────────────────────────────┐
│  platform  All   48 shown             │
└───────────────────────────────────────┘

┌─ Problems ────────────────────────────┐
│  Problem   Name              Rating  Platform │
│  ○ 1672    Shortest Routes II  1700   CSES    │
│  ● 1640    Sum of Two Values   1100   CSES    │
│  ► 1158    Book Shop           1500   CSES    │  ← selected row
└───────────────────────────────────────┘

┌─ Details ─────────────────────────────┐
│  1158  Book Shop                       │
│  rating  1500   status  Solved  —      │
│  tags    Dynamic Programming           │
│  url     https://...                   │
│  file    /Users/.../BookShop.cpp       │
└───────────────────────────────────────┘
```

**Row states:**
```
○  1672   unsolved     — hollow dot, text muted
●  1640   solved       — filled green dot, text primary
►  1158   selected     — filled purple dot, full row highlight bg
```

**Selected row highlight:**
```css
background: #2a1f5a;   /* deep purple tint, not solid purple */
color: var(--purple);  /* text becomes purple */
```

**Solved dot colors:**
```
○  unsolved    color: var(--text-dim)
●  solved      color: var(--green)     #98c379
►  selected    color: var(--purple)
```

**Column layout (Problems table):**
```
Problem   8ch    muted color
Name      ~50ch  primary text (bold on selected)
Rating    6ch    right-align, yellow (#e5c07b)
Platform  12ch   right-align, muted text
```

**Details panel — label/value grid:**
```css
/* Two-column label:value layout */
label:   color: var(--text-muted); width: 6ch; display: inline-block;
value:   color: var(--text);
accent:  rating → var(--yellow); tags → var(--purple); url → var(--text-muted);
```

**Keybind bar (above status bar):**
```
j/k move   o open   U url   T test   s submit   / search   f rating   p platform   r sync
```
Pattern: `key` in bold/white, `action` in muted. Space-separated pairs. Single line.

**Filter bar inline tokens:**
```
platform  [All]   48 shown
          ↑bold white   ↑muted
```

---

### Contests Screen

**Single full-width table panel:**
```
┌─ Contests ──────────────────────────────────────────────┐
│  When         Contest                      Length  Starts(UTC) │
│  ► in 2d 2h   Codeforces Round 990 (Div.2) 2:10   Jun 05, 09:26 │  ← upcoming selected
│    in 4d 23h  Educational CF Round 178     2:00   Jun 08, 06:26 │  ← upcoming unselected
│    3d 0h ago  Codeforces Round 989          2:30   May 31, 06:26 │  ← past
└─────────────────────────────────────────────────────────┘
```

**Temporal color coding:**
```
Upcoming (next):   ► prefix + full row highlight bg (#2a1f5a) + cyan "When" text (#4ec9b0)
Upcoming (other):  "in Xd Yh" in cyan, contest name in primary
Past:              "Xd Yh ago" in muted (#5a5f72), contest name in muted/dim
```

**"When" column format:**
```
► in 2d 2h      — next upcoming, cyan + selected
  in 4d 23h     — future, cyan
  3d 0h ago     — past, muted
  11d 0h ago    — past further, more muted
```

**Column layout (Contests table):**
```
When      14ch   cyan (upcoming) / muted (past)
Contest   ~55ch  primary (upcoming) / muted (past)
Length    8ch    right-align, muted
Starts    18ch   right-align, yellow (upcoming) / muted (past)
```

**Keybind bar:**
```
1/6   j/k move   enter/o solve its problems   b open in browser
```
Note `1/6` = current position / total count — shown in muted at far left.

---

### Analytics Screen

**Three stacked panels: Rating History, Topic Breakdown, Activity Heatmap**

#### Rating History Panel
```
┌─ Rating History ───────────────────────────────────┐
│  1847  peak 1847   range 1043–1847                  │
│                                                      │
│   [▓][▓][▓][▓][▓][▓][▓][▓][▓][▓][▓][▓]           │  ← pixel bar chart
│   ─────────────────────────────────────            │
│   1043         rating over time         1847        │
└─────────────────────────────────────────────────────┘
```

**Pixel/block bar chart** — each bar is a rectangle of fixed width (~20px) and variable height, colored by value:
```
Low rating     gray   #4a4d5a
Mid rating     green  #98c379
High rating    cyan   #4ec9b0 / blue #61afef
Peak           cyan-bright
```
No axes labels except min/max at bottom. Axis line is a simple `─` or 1px border.

**Header stat line:**
```css
/* "1847 peak 1847  range 1043–1847" */
current-value: color: var(--purple); font-weight: bold;
label:         color: var(--text-muted);
range-value:   color: var(--text);
```

#### Topic Breakdown Panel
```
┌─ Topic Breakdown — weakest first ───────────────────┐
│  Topic              Solved   Rate    Progress   Avg  │
│  matrices            0/1      0%       —          —  │
│  divide and conqu…   0/1      0%       —          —  │
└─────────────────────────────────────────────────────┘
```

**Column layout:**
```
Topic      ~22ch  primary text (truncated with …)
Solved     8ch    "X/Y" format, muted
Rate       6ch    red if 0% (#e06c75), yellow if partial
Progress   12ch   dash (—) if no data, or progress bar
Avg        6ch    dash (—) if no data
```

**Panel subtitle** — inline after section title:
```
┌─ Topic Breakdown — weakest first ─
                  ↑ em dash separator, italic/muted subtitle
```
CSS: `color: var(--text-muted); font-style: italic;` for the subtitle part.

#### Activity Heatmap Panel
```
┌─ Activity — last 52 weeks ──────────────────────────┐
│  M  · · ▪ · · ▪ ▪ · ▪ ▪ ▪ ▪ ▪▪▪ ▪▪▪▪▪▪▪▪▪ ▪▪▪▪   │
│  W  · · · · ▪ · · · · · ▪▪ · ▪▪ · ▪▪▪▪▪▪▪ ▪▪▪▪   │
│  F  · · ▪ · · ▪ · · · · ▪ · · ▪ ▪ ▪▪▪▪▪▪▪ ▪▪▪▪   │
│     less  ░ ▒ ▓ █ ▪  more solves                   │
└─────────────────────────────────────────────────────┘
```

**Heatmap cell colors (intensity scale):**
```
0 solves      transparent / bg color
1 solve       #3d2f6e   — very dim purple
2–3 solves    #5a3ea0   — mid purple
4–6 solves    #7b55d4   — bright purple
7+ solves     #9d7dea   — full purple (var(--purple))
```

**Day-of-week labels:** Only M, W, F shown (not every day) — sparse label style.

**Legend row:**
```
less  [░][▒][▓][█][▪]  more solves
       ↑ 5 cells increasing intensity, inline with text
```

---

### Recommend Screen

**Two stacked panels: Targeting banner + Recommended Problems table**

#### Targeting Banner
```
┌─ Targeting Your Weak Topics ────────────────────────────────────────────┐
│  data structures 0%  dfs and similar 25%  graphs 40%  sortings 66%  dp 75% │
└─────────────────────────────────────────────────────────────────────────┘
```

**Inline topic+percentage tokens** (space-separated, single line):
```css
topic-name:  color: var(--purple); font-weight: bold;
percentage:  color: var(--text-muted);  /* immediately after, no separator */
/* e.g.: "data structures" purple + " 0%" muted */
```

Color the percentage by urgency:
```
0%      color: var(--red)     — critical weakness
25%     color: var(--red)     — still critical
40–65%  color: var(--yellow)  — needs work
66–74%  color: var(--yellow)
75%+    color: var(--green)
```

#### Recommended Problems Table
```
┌─ Recommended Problems ──────────────────────────────┐
│  Problem   Name                    Rating   Topics   │
│  ► 877D    Olya and Energy Drinks   1900   bfs, dfs  │  ← selected
│    919D    Substring                1900   dfs, dp   │
│    525D    Arthur and Walls         2000   dfs, greedy│
└─────────────────────────────────────────────────────┘
```

**Column layout:**
```
Problem   8ch    muted; purple+bold on selected row
Name      ~35ch  primary; purple+bold on selected row
Rating    8ch    yellow (#e5c07b) always
Topics    ~30ch  muted on unselected; cyan on selected row
```

**Topics column** — comma-separated tags, all lowercase:
```
bfs, dfs and similar
dsu, graphs
greedy, sortings, two pointers
```

**Selected row:** Same deep purple bg (`#2a1f5a`) as Problems screen. Topics text turns cyan on selection.

---

## Cross-Screen Patterns

### Title Bar (top line)
A single line across the top showing the app title or page context. No traffic lights, no OS chrome.

```
BruteForce Learning — Dashboard
```
```css
background: #1a1d26;
border-bottom: 1px solid var(--border);
padding: 0.55em 1em;
text-align: center;
font-size: 0.95rem;
color: var(--text-muted);
user-select: none;
```

### Header Bar (persistent across all screens)
```
rating 1847   solved 31/48   streak 10d   xp 450/600
```
```css
/* Layout: stats spread across bar */
display: flex; justify-content: space-between; align-items: center;
border: 1px solid var(--purple);
padding: 0.5em 1.2em;
margin: 0.7em 1em 0;

/* Stats */
rating label: color: var(--text-muted);
1847 value:   color: var(--yellow); font-weight: bold;
solved X/Y:   "solved" muted, X green, "/" muted, Y muted;
streak Nd:    "streak" muted, Nd yellow;
xp label:     color: var(--text-muted);
xp value:     color: var(--purple); font-weight: bold;
```

### Dual Status Bar (bottom 2 lines)
All screens end with exactly 2 lines:
```
Line 1:  keybind hints  (bold key + muted action, space-separated)
Line 2:  › Press 'r' to sync with Codeforces and CSES
```
```css
border-top: 1px solid var(--border);
padding: 0.45em 1em;
font-size: 0.85rem;

/* Line 1 keybinds */
key:     color: var(--text); font-weight: bold;
action:  color: var(--text-muted);

/* Line 2 prompt */
color: var(--text-dim);
padding: 0.3em 1em;
prefix: "› " — muted angle quote, not ">"
```

### Panel Title Style

Title sits on the top border — absolutely positioned with `background: var(--bg)` to hide the border behind it.

```css
.panel {
  border: 1px solid var(--border);
  position: relative;
  padding: 3.5em 1em 1em;
}
.panel-title {
  position: absolute;
  top: -0.9em;
  left: 0.8em;
  background: var(--bg);
  padding: 0 0.6em;
  color: var(--purple);
  font-weight: bold;
  font-size: 0.9rem;
  line-height: 1.3;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.p-sub {
  color: var(--text-muted);
  font-weight: normal;
  font-style: italic;
  text-transform: none;
  letter-spacing: 0;
}
.panel-body {
  padding: 0;
}
```

### Row Selection System (universal)
Used in Problems, Contests, and Recommend tables:
```css
/* Unselected row */
background: transparent;
color: var(--text);

/* Hovered row */
background: #1a1d2e;

/* Selected row */
background: #2a1f5a;   /* deep purple, NOT solid accent */
color: var(--purple);  /* text shifts to purple */

/* Selected row prefix */
content: "► ";  color: var(--purple);
/* Unselected prefix */
content: "  "; (space padding to preserve alignment)
```

### Number Formatting
```
Ratings:    always yellow (#e5c07b), no units
Counts X/Y: X in green, slash+Y in muted  e.g. "31/48"
Streaks:    value+unit together in yellow  e.g. "10d"
Timing:     "in 2d 2h" cyan, "3d 0h ago" muted
Ranges:     "1043–1847" — en-dash, no spaces, text color
```

### Truncation
Long text is truncated with `…` (ellipsis char, not `...`):
```
"divide and conqu…"
```
Always truncate at a fixed character column, not with CSS `text-overflow` (maintain the monospace grid).

---

## Responsive / Viewport Scaling

The UI is designed to fill the full viewport. All inner panels, tabs, tables use `border-radius: 0`. The entire interface scales proportionally with viewport height using pure CSS — no JavaScript transforms required.

```css
html {
  font-size: clamp(12px, 1.6vh, 28px);
  /* 1rem scales with viewport height:
     - 12px minimum (small screens / mobile)
     - 1.6vh fluid (17px at 1080p, 23px at 1440p)
     - 28px maximum (large screens / 4K)
  */
}

html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  background: var(--bg);
}

#app {
  flex: 1;
  display: flex;
  flex-direction: column;
  width: 100%;           /* full-width layout */
  margin: 0 auto;
}
```

**How it works:** The root `font-size` uses `clamp()` with a `vh`-based value. All component sizes use `rem` or `em` units, so they automatically scale proportionally with the viewport height. On a 1440p display the UI is ~1.8× larger than on 720p — no zoom transforms, no JS, no media queries.

**Stat card numbers** use `font-size: 1.8rem` with `line-height: 1` for large display values that scale naturally with the viewport.

**Optional max-width:** If a narrower layout is preferred (simulating an 80–120 col terminal), use `max-width: min(86rem, 1100px)` on `#app`.

