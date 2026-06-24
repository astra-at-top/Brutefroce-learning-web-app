# Admin Panel: Per-Item CRUD for Lessons

**Date:** 2026-06-24
**Status:** Approved

## Overview

Add per-item management to the Admin panel using modal dialogs for adding, editing, and deleting individual quiz questions, flashcard pairs, and coding exercises.

## Current State

The admin panel uses a 5-step wizard (Info, Notes, Quiz, Coding, Flashcards) with batch saves:
- Quiz: All questions saved as a single array
- Flashcards: Add/remove by count (not individual pairs)
- Coding: Single prompt/starter/solution object

Backend (`handleAdminLessonByID`) already supports individual component updates via separate PUT requests with different JSON keys.

## Changes

### 1. Quiz Questions

**UI:** Replace batch form with a scrollable list of question cards.

Each card shows:
- Question number and text
- Correct answer preview
- Edit (✏️) and Delete (✕) buttons

**Add/Edit Modal:**
- Question textarea
- 4 option inputs (A, B, C, D)
- Radio buttons to select correct answer
- Save/Cancel buttons

**Backend:** No changes needed. Existing `PUT /api/admin/lessons/{id}` with `{quiz: [...]}` key works.

### 2. Flashcards

**UI:** Replace counter-based add/remove with a list of flashcard pair cards.

Each card shows:
- Card number
- Front text preview (truncated)
- Back text preview (truncated)
- Edit (✏️) and Delete (✕) buttons

**Add/Edit Modal:**
- Front textarea
- Back textarea
- Live preview of card appearance
- Save/Cancel buttons

**Backend:** No structural changes. Existing flashcard save logic handles individual pairs.

### 3. Coding Exercises

**UI:** Keep as single card but show Edit button.

**Edit Modal:**
- Prompt textarea
- Starter code textarea
- Solution textarea
- Save/Cancel buttons

**Backend:** No changes needed.

### 4. Modal Component (Shared)

Create reusable modal infrastructure:
- HTML modal structure (hidden by default)
- Open modal function with form population
- Close modal function
- Delete confirmation dialog

## Implementation Notes

- Modal HTML added once in admin panel, shown/hidden via JS
- Each item type (quiz, flashcard, coding) has its own modal template
- Edit populates form from existing item data
- Add opens modal with empty form
- Delete shows confirmation before removing
- Lists re-render after any add/edit/delete operation

## Files to Modify

- `web/index.html`: Add modal HTML, modify admin wizard steps to use list views
- `main.go`: No changes needed (existing endpoints support individual operations)
