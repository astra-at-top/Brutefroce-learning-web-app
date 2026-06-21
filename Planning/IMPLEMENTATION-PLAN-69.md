# Implementation Plan: Clear Old Admin Data + Add Precise Original Content

## Overview
Replace all old/dummy seed data (course structure, lesson content, quizzes, coding exercises, flashcards, achievements) from the prototype phase with original, precise educational content. Update frontend hardcoded data to match. Delete old SQLite DB so fresh seed takes effect.

## New Curriculum (per PLAN.md)
```
JavaScript (foundation)
├── Variables & Types
├── Functions & Scope
├── Arrays & Objects
├── Closures
├── Promises & Async

React
├── JSX & Components
├── useState Hook
├── useEffect Hook
├── Data Fetching

Git
├── Git Basics
├── Branching & Merging

Node.js
├── Node.js Intro
```

## Task List

### Task 1: Restructure course curriculum & rewrite lesson content
**File:** `main.go`
- Rewrite `courses` slice: 4 top-level courses (JavaScript, React, Git, Node.js) with correct `parent_id` and `sort_order`
- Rewrite `lessons` slice: all 12 lessons with original notes (detailed, educational), mnemonics, appropriate status/progress
- Remove old duplicate/confusing structure (JS Basics under React + standalone JS conflict)
- Keep DB schema unchanged

**Acceptance criteria:**
- [ ] Course tree shows 4 correct top-level courses with proper hierarchy
- [ ] Each lesson has detailed, original notes (≥3 sentences)
- [ ] Each lesson has a mnemonic (or empty where none makes sense)
- [ ] Status/progress values form a logical learning path

### Task 2: Rewrite quiz questions, coding exercises, and flashcards
**File:** `main.go`
- Every lesson gets 2-4 original quiz questions (not generic ones like "Which React hook adds state?")
- JavaScript lessons get coding exercises with prompt, starter, solution
- React lessons with coding exercises too
- 3-5 flashcards per major lesson covering key concepts
- All quiz questions test understanding, not trivia

**Acceptance criteria:**
- [ ] Every lesson has quiz questions (minimum 2)
- [ ] Quiz questions are original and topic-specific
- [ ] Coding exercises have working starter/solution
- [ ] Flashcards cover key concepts with precise Q&A

### Task 3: Update frontend hardcoded data in index.html
**File:** `web/index.html`
- Quiz tab lesson list (line 988): update 6 lessons → 12 lessons matching new curriculum
- Code tab exercise data (lines 1054-1059): update to match new coding exercises
- Dashboard fallback stats (line 526): realistic values
- Leaderboard AI competitors (lines 1763-1771): more creative names
- Daily checklist items (lines 1842-1851): match new curriculum
- Achievements list (lines 1906-1913): more creative achievements

**Acceptance criteria:**
- [ ] Quiz tab shows all lessons from new curriculum
- [ ] Code tab exercises match seed data
- [ ] Checklist items reference real lesson IDs
- [ ] All hardcoded arrays align with seed data

### Task 4: Wipe old database, build, and verify
- Delete `data/brutforse.db`
- Build Go binary: `go build -o bin/brutforse`
- Run server: `./bin/brutforse`
- Verify admin panel loads with new course/lesson tree
- Verify learn tab shows courses
- Verify quiz tab works with new questions
- Verify code tab works with new exercises

**Acceptance criteria:**
- [ ] Fresh seed runs without errors
- [ ] Admin panel shows 4 courses, 12 lessons with new content
- [ ] Quiz questions load and are answerable
- [ ] Coding exercises load correctly
- [ ] Flashcards render in review mode
- [ ] Progress dashboard shows correct initial values

## Dependencies
- Tasks 1-2 must come before Task 3 (frontend must match backend)
- Task 4 comes last (verification step)

## Verification
1. `go build ./...` — compiles clean
2. `./bin/brutforse` — runs without error
3. Manual check: all tabs functional in browser
