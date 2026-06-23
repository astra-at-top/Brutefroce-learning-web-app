# JSON Import/Export — Admin Design

**Date:** 2026-06-23
**Status:** Approved
**Approach:** Stacked panel (Approach 1)

## Problem

User generates learning content (courses, lessons, quiz, flashcards) from AI tools and wants to import it into BruteForce Learning. Currently the admin has a JSON import section but:
- No way to export current data to see the format
- Import does a full wipe instead of merge

## Solution

Add a collapsible **"Export as JSON"** section above the existing JSON Import section at the bottom of the admin page. The import mode is changed from full-wipe to merge/upsert.

## Workflow

1. Click **"Export as JSON"** → textarea fills with current data as JSON
2. Click **"Copy"** → clipboard
3. Paste into ChatGPT/Claude, modify
4. Copy modified JSON
5. Paste into existing **"JSON Import"** textarea
6. Click **"Import (Merge)"** → upserts courses/lessons

## Backend Changes

### `POST /api/admin/import` — Add merge mode
- Add `mode` field to request body (`"merge"` or `"replace"`)
- **Merge mode** (default):
  - Courses: `INSERT ... ON CONFLICT(id) DO UPDATE`
  - Lessons: `INSERT ... ON CONFLICT(id) DO UPDATE` — preserves progress, flashcard_reviews
  - Items in DB but not in JSON → keep untouched
- **Replace mode** (existing behavior via `mode: "replace"`):
  - Wipes and re-inserts as before

### `GET /api/admin/export` — New endpoint
- Returns all data as JSON in the import format:
  ```json
  {
    "courses": [
      {
        "id": "...",
        "name": "...",
        "lessons": [
          {
            "id": "...",
            "title": "...",
            "notes": "...",
            "quiz": [...],
            "coding": {...},
            "flashcards": [...]
          }
        ]
      }
    ]
  }
  ```

## Frontend Changes

### `web/index.html` — Add export section
- Collapsible **"Export as JSON"** section with:
  - "▶ Export as JSON" button
  - Read-only textarea showing the JSON
  - "📋 Copy" button
  - Size info (e.g. "2 courses, 12 lessons — 4.2 KB")

### `web/index.html` — Update import section
- Import mode defaults to merge
- Import button shows status accordingly

## Merge Logic Detail

### Courses
```sql
INSERT INTO courses (id, name, parent_id, sort_order)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  name = excluded.name,
  parent_id = excluded.parent_id,
  sort_order = excluded.sort_order
```

### Lessons
```sql
INSERT INTO lessons (id, title, notes, mnemonic, status, type, course_id, progress, quiz_count, quiz_data, coding_data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  title = excluded.title,
  notes = excluded.notes,
  mnemonic = excluded.mnemonic,
  status = excluded.status,
  type = excluded.type,
  course_id = excluded.course_id,
  quiz_data = excluded.quiz_data,
  coding_data = excluded.coding_data,
  quiz_count = excluded.quiz_count
  -- progress is NOT updated (preserve user progress)
```

### Flashcards
- For merged lessons, delete existing flashcards and re-insert from JSON
- Flashcard reviews (spaced repetition data) are preserved

## Files Changed

| File | Change |
|------|--------|
| `main.go` | Add `GET /api/admin/export` handler + register route |
| `main.go` | Modify `handleAdminImport` to support merge mode |
| `web/index.html` | Add export section + update import for merge |
