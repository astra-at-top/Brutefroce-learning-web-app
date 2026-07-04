# BruteForce Learning — Full Plan

## 1. Architecture

| Layer | Tech |
|---|---|
| Frontend | Plain HTML + CSS + vanilla JS |
| Backend | Go (single binary, serves static + API on `localhost:8080`) |
| Database | SQLite (local, managed by Go) |
| AI | OpenCode Go API (`opencode.ai/zen/go/v1`) via OpenAI Go SDK + local LLM fallback |

**No accounts.** One user per machine. Sync via relay code (encrypted upload/download).

## 2. Sections & Content

**Coding** (shipped first): JavaScript → React → Next.js → Git → DSA → Node.js
**Self-help**: Added later, same structure. Books added on demand.

Content from **3 sources**: curated (you write), user-uploaded, AI-generated.

## 3. Skill Tree

```
React
├── JS Basics (locked until 80% pass)
│   ├── Variables & Types → quiz + coding
│   ├── Functions & Scope → quiz + coding
│   └── Arrays & Objects → quiz + coding
├── useState (locked until JS Basics complete)
│   ├── What is state? → notes + quiz
│   └── Counter exercise → live coding
├── useEffect (locked until useState complete)
└── ...
```

Each node: **Notes** (simple language + memory hooks) → **Quiz** → **Coding exercise** → **Spaced-rep review** → **Unlock next**

## 4. Memorization Engine

| Technique | How it's applied |
|---|---|
| Spaced repetition (SM-2) | Every card has a schedule: 1d → 3d → 7d → 30d → 90d |
| Active recall | Every question forces answer before revealing |
| Chunking | Lessons are max 5 concepts, never more |
| Mnemonics | AI generates acronyms/stories for each concept |
| Interleaving | Daily review mixes different topics |
| Feynman technique | "Explain this in one sentence" prompts |
| Dual coding | Text + diagrams side by side |

## 5. Exercise Types per Node

- **Multiple choice / fill-blank** — quiz mode
- **Code tracing** — "What does this output?"
- **Drag & drop** — order code blocks
- **Live coding** — iframe sandbox, user writes and runs code

## 6. Addiction Mechanics

All 6:
- **Streaks & loss aversion** — "You'll lose your 12-day streak"
- **Variable rewards** — random XP bonuses, surprise double XP
- **Fake social pressure** — AI-generated competitors, "Your rank dropped to #7"
- **Progress bars & micro-commitments** — "83% complete... just 5 more minutes"
- **Sound/vibration** — satisfying feedback, combos, celebration animations
- **Dopamine menu** — after each chapter: claim XP / unlock badge / power-up

## 7. Scoring

```
Power Score = XP (daily cap) × Mastery% × Streak_multiplier
```

Leaderboard shows rank with AI-generated fake competitors.

## 8. Habit Tracker

Three-in-one: daily checklist + session timer + milestone achievements. Calendar view.

## 9. UI Layout

**Tabbed**: Learn | Quiz | Code | Progress | Stats/Habits

Dashboard shows: daily goal, streak, XP today, "Continue where you left off", habit calendar.

## 10. Sync

User clicks "Export" → Go app encrypts SQLite data → uploads to relay → shows code "ABCD-1234". On other machine: "Import" → enter code → download + decrypt.
