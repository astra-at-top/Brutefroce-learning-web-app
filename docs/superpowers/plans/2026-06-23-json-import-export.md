# JSON Import/Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add JSON export and merge import to the admin panel so users can copy data, modify in AI tools, and paste back.

**Architecture:** Backend Go handler exports all courses/lessons/flashcards as JSON matching the import format. Import handler gets a `mode: "merge"` branch that upserts instead of wiping. Frontend adds an export section with textarea + copy button alongside the existing import section.

**Tech Stack:** Go, modernc.org/sqlite, vanilla JS single-page app

## Global Constraints

- Follow existing code patterns in `main.go` and `web/index.html`
- Use `jsonResp` / `jsonError` helpers for all API responses
- SQLite `INSERT ... ON CONFLICT ... DO UPDATE` for merge logic
- Must compile with `go build -o brutforselearning`

---

### Task 1: Add GET /api/admin/export handler

**Files:**
- Modify: `main.go` (add handler + register route)

**Interfaces:**
- Produces: `GET /api/admin/export` returns `{"courses": [{id, name, parent_id, lessons: [{id, title, notes, mnemonic, status, type, quiz, coding, flashcards}]}]}`

- [ ] **Register the route** in `main.go` after line 78:
  ```go
  mux.HandleFunc("/api/admin/export", handleAdminExport)
  ```

- [ ] **Add the handler function** before `handleAdminImport`:
  ```go
  func handleAdminExport(w http.ResponseWriter, r *http.Request) {
  	if r.Method != "GET" {
  		jsonError(w, "method not allowed", 405)
  		return
  	}

  	// Load all courses
  	rows, err := db.Query("SELECT id, name, parent_id, sort_order FROM courses ORDER BY sort_order")
  	if err != nil {
  		jsonError(w, err.Error(), 500)
  		return
  	}
  	defer rows.Close()

  	type courseExport struct {
  		ID       string          `json:"id"`
  		Name     string          `json:"name"`
  		ParentID string          `json:"parent_id"`
  		Lessons  []lessonExport  `json:"lessons"`
  	}
  	type lessonExport struct {
  		ID         string                   `json:"id"`
  		Title      string                   `json:"title"`
  		Notes      string                   `json:"notes"`
  		Mnemonic   string                   `json:"mnemonic"`
  		Status     string                   `json:"status"`
  		Type       string                   `json:"type"`
  		Quiz       []map[string]interface{} `json:"quiz"`
  		Coding     map[string]interface{}   `json:"coding"`
  		Flashcards []map[string]string      `json:"flashcards"`
  	}

  	var courses []courseExport
  	for rows.Next() {
  		var id, name, parentID string
  		var sort int
  		rows.Scan(&id, &name, &parentID, &sort)
  		courses = append(courses, courseExport{ID: id, Name: name, ParentID: parentID})
  	}

  	// Load lessons for each course
  	lRows, err := db.Query("SELECT id, title, notes, mnemonic, status, type, course_id, quiz_data, coding_data FROM lessons ORDER BY id")
  	if err != nil {
  		jsonError(w, err.Error(), 500)
  		return
  	}
  	defer lRows.Close()

  	type lessonRow struct {
  		ID, Title, Notes, Mnemonic, Status, Ltype, CourseID, QuizData, CodingData string
  	}
  	lessonMap := map[string][]lessonRow{}
  	for lRows.Next() {
  		var l lessonRow
  		lRows.Scan(&l.ID, &l.Title, &l.Notes, &l.Mnemonic, &l.Status, &l.Ltype, &l.CourseID, &l.QuizData, &l.CodingData)
  		lessonMap[l.CourseID] = append(lessonMap[l.CourseID], l)
  	}

  	// Load flashcards
  	fcRows, err := db.Query("SELECT lesson_id, front, back FROM flashcard_data ORDER BY id")
  	if err == nil {
  		defer fcRows.Close()
  	}
  	fcMap := map[string][]map[string]string{}
  	if fcRows != nil {
  		for fcRows.Next() {
  			var lessonID, front, back string
  			fcRows.Scan(&lessonID, &front, &back)
  			fcMap[lessonID] = append(fcMap[lessonID], map[string]string{"front": front, "back": back})
  		}
  	}

  	// Build response
  	for i, c := range courses {
  		lessons := lessonMap[c.ID]
  		for _, l := range lessons {
  			var quiz []map[string]interface{}
  			json.Unmarshal([]byte(l.QuizData), &quiz)
  			var coding map[string]interface{}
  			json.Unmarshal([]byte(l.CodingData), &coding)
  			fcs := fcMap[l.ID]
  			if fcs == nil {
  				fcs = []map[string]string{}
  			}
  			courses[i].Lessons = append(courses[i].Lessons, lessonExport{
  				ID: l.ID, Title: l.Title, Notes: l.Notes, Mnemonic: l.Mnemonic,
  				Status: l.Status, Type: l.Ltype, Quiz: quiz, Coding: coding, Flashcards: fcs,
  			})
  		}
  	}

  	jsonResp(w, map[string]interface{}{"courses": courses})
  }
  ```

- [ ] **Build and verify it compiles**

  Run: `go build -o brutforselearning 2>&1`
  Expected: no errors

- [ ] **Commit**

  ```bash
  git add main.go
  git commit -m "feat: add GET /api/admin/export endpoint"
  ```

---

### Task 2: Modify handleAdminImport for merge mode

**Files:**
- Modify: `main.go` (handleAdminImport function)

**Interfaces:**
- Consumes: `POST /api/admin/import` with `{"mode": "merge", "courses": [...]}` or `{"mode": "replace", "courses": [...]}`
- Mode defaults to `"merge"` if not specified
- For merge: upsert courses and lessons, preserve progress/flashcard reviews

- [ ] **Add merge mode check** at the top of `handleAdminImport`, after parsing the body JSON:
  ```go
  mode := "merge"
  if m, ok := data["mode"].(string); ok {
  	mode = m
  }
  ```

- [ ] **In the courses import branch** (Mode 1), replace the wipe logic with:
  ```go
  if mode == "replace" {
  	db.Exec("DELETE FROM flashcard_reviews")
  	db.Exec("DELETE FROM flashcard_data")
  	db.Exec("DELETE FROM lessons")
  	db.Exec("DELETE FROM courses")
  }
  ```
  Then change the INSERT to:
  ```go
  if mode == "replace" {
  	db.Exec("INSERT INTO courses (id,name,parent_id,sort_order) VALUES (?,?,?,?)", id, name, parentID, sort)
  } else {
  	db.Exec("INSERT INTO courses (id,name,parent_id,sort_order) VALUES (?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, parent_id=excluded.parent_id, sort_order=excluded.sort_order", id, name, parentID, sort)
  }
  ```

- [ ] **Update the lesson insert** in Mode 1's lesson loop:
  ```go
  if mode == "replace" {
  	db.Exec("DELETE FROM lessons WHERE id=?", lid)
  	db.Exec("INSERT INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
  		lid, title, notes, mnemonic, status, ltype, id, progress, quizCount, quizData, codingData)
  } else {
  	// For merge, only wipe flashcards (they get re-inserted below)
  	db.Exec("DELETE FROM flashcard_data WHERE lesson_id=?", lid)
  	db.Exec("INSERT INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, notes=excluded.notes, mnemonic=excluded.mnemonic, status=excluded.status, type=excluded.type, course_id=excluded.course_id, quiz_data=excluded.quiz_data, coding_data=excluded.coding_data, quiz_count=excluded.quiz_count",
  		lid, title, notes, mnemonic, status, ltype, id, progress, quizCount, quizData, codingData)
  }
  ```

- [ ] **Update Mode 2 (single lesson)** similarly — check mode before DELETE + use ON CONFLICT for merge

- [ ] **Build and verify it compiles**
  Run: `go build -o brutforselearning 2>&1`
  Expected: no errors

- [ ] **Commit**
  ```bash
  git add main.go
  git commit -m "feat: add merge mode to admin import"
  ```

---

### Task 3: Add Export UI to admin frontend

**Files:**
- Modify: `web/index.html` (add export section in the admin panel)

- [ ] **Add the Export as JSON section** after the existing sync section (after line 1675) and before the existing JSON import section:

  ```javascript
  // Export as JSON section
  const exportArea = document.createElement('div');
  exportArea.style.cssText = 'border-top:1px solid var(--border-dim);flex-shrink:0';
  exportArea.innerHTML = `
    <div style="display:flex;justify-content:space-between;align-items:center;padding:0.4em 1em;cursor:pointer;background:var(--bg2)" id="jsonExportToggle">
      <span style="font-size:0.7rem;color:var(--blue);letter-spacing:0.05em;text-transform:uppercase">📤 Export as JSON — copy current data to modify</span>
      <span style="font-size:0.65rem;color:var(--text-dim)" id="jsonExportArrow">▼</span>
    </div>
    <div id="jsonExportBody" style="padding:0.5em 1em">
      <div class="btn-grp" style="margin-bottom:0.4em">
        <button class="btn" id="jsonExportBtn">▶ Export as JSON</button>
        <button class="btn btn-alt" id="jsonExportCopyBtn" style="display:none">📋 Copy</button>
        <span id="jsonExportStatus" style="font-size:0.7rem;color:var(--text-dim);align-self:center"></span>
      </div>
      <textarea id="jsonExportTextarea" readonly placeholder="Click 'Export as JSON' to load current data..." style="width:100%;min-height:120px;font-family:'Courier New',monospace;font-size:0.7rem;padding:0.5em;background:var(--surface-alt);border:1px solid var(--border-dim);border-radius:4px;color:var(--text);resize:vertical"></textarea>
      <div style="font-size:0.6rem;color:var(--text-dim);margin-top:0.3em;line-height:1.6">
        <strong>Usage:</strong> Export → Copy → Modify in AI → Paste into <strong>JSON Import</strong> below → Import
      </div>
    </div>`;
  document.querySelector('.admin-box')?.closest('.admin-editor-panel')?.after(exportArea);

  // Collapse toggle
  let exportExpanded = true;
  document.getElementById('jsonExportToggle')?.addEventListener('click', () => {
    exportExpanded = !exportExpanded;
    document.getElementById('jsonExportBody').style.display = exportExpanded ? '' : 'none';
    document.getElementById('jsonExportArrow').textContent = exportExpanded ? '▼' : '▶';
  });

  // Export button
  document.getElementById('jsonExportBtn')?.addEventListener('click', async () => {
    const btn = document.getElementById('jsonExportBtn');
    const status = document.getElementById('jsonExportStatus');
    const textarea = document.getElementById('jsonExportTextarea');
    const copyBtn = document.getElementById('jsonExportCopyBtn');
    btn.textContent = '⏳ Exporting...';
    btn.disabled = true;
    try {
      const res = await fetch('/api/admin/export');
      const data = await res.json();
      const pretty = JSON.stringify(data, null, 2);
      textarea.value = pretty;
      textarea.style.display = '';
      copyBtn.style.display = '';
      const courseCount = data.courses ? data.courses.length : 0;
      let lessonCount = 0;
      if (data.courses) for (const c of data.courses) lessonCount += (c.lessons || []).length;
      status.textContent = '✅ ' + courseCount + ' courses, ' + lessonCount + ' lessons — ' + (pretty.length / 1024).toFixed(1) + ' KB';
    } catch (e) {
      status.textContent = '❌ Export failed: ' + e.message;
    }
    btn.textContent = '▶ Export as JSON';
    btn.disabled = false;
  });

  // Copy button
  document.getElementById('jsonExportCopyBtn')?.addEventListener('click', async () => {
    const textarea = document.getElementById('jsonExportTextarea');
    try {
      await navigator.clipboard.writeText(textarea.value);
      document.getElementById('jsonExportStatus').textContent = '✅ Copied to clipboard!';
    } catch {
      textarea.select();
      document.execCommand('copy');
      document.getElementById('jsonExportStatus').textContent = '✅ Copied!';
    }
  });
  ```

- [ ] **Commit**
  ```bash
  git add web/index.html
  git commit -m "feat: add JSON export section to admin panel"
  ```

---

### Task 4: Update Import UI for merge mode

**Files:**
- Modify: `web/index.html` (update existing JSON import section)

- [ ] **Update the JSON Import section header** text to indicate merge mode:
  Change line 1682 from:
  ```
  📥 JSON Import — paste course/topic data
  ```
  to:
  ```
  📥 JSON Import (Merge) — paste modified data here
  ```

- [ ] **Update the JSON import format hint** (line 1715-1718) to note merge behavior:
  Change to:
  ```
  <strong>Format:</strong> Same as Export format above. Uses <strong>merge mode</strong> — 
  existing courses/lessons are updated, new ones added, data not in JSON is preserved, 
  progress and flashcard reviews are kept.
  ```

- [ ] **Commit**
  ```bash
  git add web/index.html
  git commit -m "feat: update import UI for merge mode"
  ```

---

### Task 5: Build and test on separate port

- [ ] **Build the binary**
  Run: `go build -o brutforselearning 2>&1`
  Expected: no errors

- [ ] **Run on separate port**
  Run: `PORT=8081 DATA_DIR=./data ./brutforselearning`
  Expected: `BruteForce Learning running on :8081`
