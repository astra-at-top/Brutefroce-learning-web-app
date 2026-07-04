# Admin Panel: Per-Item CRUD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-item management (add/edit/delete) for quiz questions, flashcards, and coding exercises via modal dialogs in the admin panel.

**Architecture:** Convert the batch forms in the admin wizard steps (Quiz, Flashcards, Coding) into list views with individual item cards. Each card has Edit/Delete buttons that open a modal form. Adding opens a modal with empty form.

**Tech Stack:** Vanilla JS frontend, Go backend (no changes needed).

---

## Files to Modify

- `web/index.html`: Add modal HTML, modify admin wizard steps to use list views

---

## Task 1: Add Modal Infrastructure

**Files:**
- Modify: `web/index.html` — add modal HTML structure and CSS

**Interfaces:**
- Produces: `adminModal` object with `open(type, data)`, `close()`, `confirmDelete(type, index)` methods

- [ ] **Step 1: Add modal CSS**

Find existing admin styles in `web/index.html` (around line 380+). Add this CSS after existing admin styles:

```css
/* ===== Admin Modal ===== */
.admin-modal-overlay {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center;
  z-index: 200; display: none;
}
.admin-modal-overlay.active { display: flex; }
.admin-modal {
  background: var(--surface); border: 2px solid var(--purple); border-radius: 12px;
  max-width: 600px; width: 90%; max-height: 90vh; overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0,0,0,0.5);
}
.admin-modal-header {
  background: var(--purple); padding: 12px 16px; display: flex; justify-content: space-between; align-items: center;
}
.admin-modal-header h3 { color: white; margin: 0; font-size: 1rem; }
.admin-modal-close { background: none; border: none; color: white; font-size: 1.2rem; cursor: pointer; padding: 0; }
.admin-modal-body { padding: 20px; }
.admin-modal-footer { padding: 12px 16px; display: flex; gap: 8px; justify-content: flex-end; border-top: 1px solid var(--border); }
.admin-modal .form-group { margin-bottom: 16px; }
.admin-modal .form-group label { display: block; color: var(--purple); font-size: 0.75rem; text-transform: uppercase; margin-bottom: 4px; }
.admin-modal textarea { width: 100%; height: 60px; background: var(--bg-deep); border: 1px solid var(--border); color: var(--text); border-radius: 4px; padding: 8px; font-family: inherit; resize: vertical; }
.admin-modal input[type="text"] { width: 100%; background: var(--bg-deep); border: 1px solid var(--border); color: var(--text); border-radius: 4px; padding: 8px; }
.admin-modal .option-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.admin-modal .option-row input[type="radio"] { accent-color: var(--purple); }
.admin-modal .option-row input[type="text"] { flex: 1; }
.admin-modal .option-label { color: var(--text-dim); font-size: 0.8rem; width: 20px; }
```

- [ ] **Step 2: Add modal HTML**

Find `<!-- admin wizard end -->` or similar marker. Add modal HTML before the closing `</div>` of admin panel:

```html
<!-- Admin Modal Overlay -->
<div class="admin-modal-overlay" id="adminModalOverlay">
  <div class="admin-modal" id="adminModal">
    <div class="admin-modal-header">
      <h3 id="adminModalTitle">Edit Item</h3>
      <button class="admin-modal-close" onclick="adminModal.close()">✕</button>
    </div>
    <div class="admin-modal-body" id="adminModalBody"></div>
    <div class="admin-modal-footer" id="adminModalFooter"></div>
  </div>
</div>
```

- [ ] **Step 3: Add modal JavaScript**

Find the admin JavaScript section (around line 1750+). Add this object after admin state:

```javascript
// Admin Modal Controller
const adminModal = {
  overlay: null,
  modal: null,
  titleEl: null,
  bodyEl: null,
  footerEl: null,
  currentType: null,  // 'quiz' | 'flashcard' | 'coding'
  currentIndex: null,
  lessonId: null,
  
  init() {
    this.overlay = document.getElementById('adminModalOverlay');
    this.modal = document.getElementById('adminModal');
    this.titleEl = document.getElementById('adminModalTitle');
    this.bodyEl = document.getElementById('adminModalBody');
    this.footerEl = document.getElementById('adminModalFooter');
    this.overlay.addEventListener('click', (e) => { if (e.target === this.overlay) this.close(); });
  },
  
  open(type, index, lessonId) {
    this.currentType = type;
    this.currentIndex = index;
    this.lessonId = lessonId;
    if (type === 'quiz') this.openQuizModal(index);
    else if (type === 'flashcard') this.openFlashcardModal(index);
    else if (type === 'coding') this.openCodingModal();
    this.overlay.classList.add('active');
  },
  
  close() {
    this.overlay.classList.remove('active');
    this.currentType = null;
    this.currentIndex = null;
  },
  
  confirmDelete(type, index) {
    if (confirm('Delete this ' + type + '?')) {
      this.deleteItem(type, index);
    }
  },
  
  async deleteItem(type, index) {
    // Implemented in Task 4
  }
};
adminModal.init();
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): add modal infrastructure"
```

---

## Task 2: Convert Quiz Step to List View

**Files:**
- Modify: `web/index.html` — find `renderAdminStep('quiz', ...)` function

**Interfaces:**
- Uses: `adminModal.open('quiz', index, lessonId)`, `adminModal.confirmDelete('quiz', index)`
- Produces: Updated quiz step showing list of question cards

- [ ] **Step 1: Find quiz step function**

Search for `renderAdminStep` function. The quiz case starts around line 1650+. Current quiz step uses a loop with `v6AddQuiz()`.

- [ ] **Step 2: Replace with list view rendering**

Find the quiz case and replace with:

```javascript
case 'quiz': {
  html += '<div class="v1-ename">Step 3: Quiz</div>';
  const questions = adminForm.quiz || [];
  html += '<div style="max-height: 300px; overflow-y: auto; margin-bottom: 12px;">';
  if (questions.length === 0) {
    html += '<div style="color: var(--text-dim); text-align: center; padding: 20px;">No questions yet</div>';
  } else {
    for (let i = 0; i < questions.length; i++) {
      const q = questions[i];
      html += '<div class="admin-item-card">';
      html += '<div class="admin-item-content">';
      html += '<div class="admin-item-num">Q' + (i + 1) + '</div>';
      html += '<div class="admin-item-text">' + escapeHtml(q.question) + '</div>';
      html += '<div class="admin-item-meta">Correct: ' + escapeHtml(q.options[q.correct_index]) + '</div>';
      html += '</div>';
      html += '<div class="admin-item-actions">';
      html += '<button class="btn btn-sm" onclick="adminModal.open(\'quiz\', ' + i + ', adminForm.id)">✏️</button>';
      html += '<button class="btn btn-sm btn-red" onclick="adminModal.confirmDelete(\'quiz\', ' + i + ')">✕</button>';
      html += '</div>';
      html += '</div>';
    }
  }
  html += '</div>';
  html += '<button class="btn btn-sm btn-alt" onclick="adminModal.open(\'quiz\', -1, adminForm.id)">+ Add Question</button>';
  break;
}
```

- [ ] **Step 3: Add admin item card CSS**

Add to admin styles:

```css
.admin-item-card {
  display: flex; justify-content: space-between; align-items: flex-start;
  background: var(--bg); border: 1px solid var(--border); border-radius: 6px;
  padding: 10px 12px; margin-bottom: 8px;
}
.admin-item-content { flex: 1; min-width: 0; }
.admin-item-num { color: var(--purple); font-size: 0.7rem; font-weight: bold; margin-bottom: 2px; }
.admin-item-text { color: var(--text); font-size: 0.85rem; margin-bottom: 4px; }
.admin-item-meta { color: var(--text-dim); font-size: 0.7rem; }
.admin-item-actions { display: flex; gap: 4px; flex-shrink: 0; }
.btn-red { background: var(--red-dim) !important; color: white !important; }
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): convert quiz step to list view with edit/delete"
```

---

## Task 3: Add Quiz Modal Form

**Files:**
- Modify: `web/index.html` — add `adminModal.openQuizModal()` method

**Interfaces:**
- Uses: `adminForm.quiz` array
- Produces: Modal form for adding/editing quiz questions

- [ ] **Step 1: Add openQuizModal method**

Add to `adminModal` object:

```javascript
openQuizModal(index) {
  const isEdit = index >= 0;
  const question = isEdit ? adminForm.quiz[index] : { question: '', options: ['', '', '', ''], correct_index: 0 };
  
  this.titleEl.textContent = isEdit ? 'Edit Question' : 'Add Question';
  
  this.bodyEl.innerHTML = `
    <div class="form-group">
      <label>Question</label>
      <textarea id="modalQQuestion">${escapeHtml(question.question)}</textarea>
    </div>
    <div class="form-group">
      <label>Options (select correct answer)</label>
      ${[0,1,2,3].map(i => `
        <div class="option-row">
          <input type="radio" name="correctOpt" value="${i}" ${question.correct_index === i ? 'checked' : ''}>
          <span class="option-label">${String.fromCharCode(65 + i)}.</span>
          <input type="text" id="modalQOpt${i}" value="${escapeHtml(question.options[i])}" placeholder="Option ${i + 1}">
        </div>
      `).join('')}
    </div>
  `;
  
  this.footerEl.innerHTML = `
    <button class="btn btn-sm" onclick="adminModal.close()">Cancel</button>
    <button class="btn btn-sm btn-primary" onclick="adminModal.saveQuiz(${index})">Save</button>
  `;
},
```

- [ ] **Step 2: Add saveQuiz method**

Add to `adminModal` object:

```javascript
saveQuiz(index) {
  const question = document.getElementById('modalQQuestion').value.trim();
  const options = [0,1,2,3].map(i => document.getElementById('modalQOpt' + i).value.trim());
  const correct_index = parseInt(document.querySelector('input[name="correctOpt"]:checked')?.value || '0');
  
  if (!question || options.some(o => !o)) {
    alert('Question and all options required');
    return;
  }
  
  if (!adminForm.quiz) adminForm.quiz = [];
  
  const newQuestion = { question, options, correct_index };
  
  if (index >= 0) {
    adminForm.quiz[index] = newQuestion;
  } else {
    adminForm.quiz.push(newQuestion);
  }
  
  adminDirty = true;
  this.close();
  renderAdminLesson();
}
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): add quiz modal form with add/edit"
```

---

## Task 4: Implement Delete and Connect to Backend

**Files:**
- Modify: `web/index.html` — implement `adminModal.deleteItem()` and connect to backend

**Interfaces:**
- Uses: `api()` function, `adminModal.close()`
- Produces: Working delete that updates UI and saves to server

- [ ] **Step 1: Add deleteItem implementation**

Replace the stub in `adminModal`:

```javascript
async deleteItem(type, index) {
  if (type === 'quiz') {
    adminForm.quiz.splice(index, 1);
  } else if (type === 'flashcard') {
    adminForm.flashcards.splice(index, 1);
  }
  
  adminDirty = true;
  this.close();
  
  // Save to backend
  await api('/api/admin/lessons/' + this.lessonId, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ 
      quiz: adminForm.quiz,
      flashcards: adminForm.flashcards 
    })
  });
  
  renderAdminLesson();
}
```

- [ ] **Step 2: Verify escapeHtml helper exists**

Search for `escapeHtml` function. If not found, add it near top of script section:

```javascript
function escapeHtml(s) {
  if (!s) return '';
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): implement delete with backend sync"
```

---

## Task 5: Convert Flashcards Step to List View

**Files:**
- Modify: `web/index.html` — find and replace flashcards case in `renderAdminStep`

**Interfaces:**
- Uses: `adminModal.open('flashcard', index, lessonId)`, `adminModal.confirmDelete('flashcard', index)`
- Produces: Flashcard list with Edit/Delete buttons

- [ ] **Step 1: Find flashcards case**

Search for `case 'flashcards':` in `renderAdminStep`. Around line 1700+.

- [ ] **Step 2: Replace with list view**

```javascript
case 'flashcards': {
  html += '<div class="v1-ename">Step 5: Flashcards</div>';
  const cards = adminForm.flashcards || [];
  html += '<div style="max-height: 300px; overflow-y: auto; margin-bottom: 12px;">';
  if (cards.length === 0) {
    html += '<div style="color: var(--text-dim); text-align: center; padding: 20px;">No flashcards yet</div>';
  } else {
    for (let i = 0; i < cards.length; i++) {
      const card = cards[i];
      html += '<div class="admin-item-card">';
      html += '<div class="admin-item-content">';
      html += '<div class="admin-item-num">Card ' + (i + 1) + '</div>';
      html += '<div class="admin-item-text">' + escapeHtml(card.front.substring(0, 50)) + (card.front.length > 50 ? '...' : '') + '</div>';
      html += '<div class="admin-item-meta">Back: ' + escapeHtml(card.back.substring(0, 30)) + (card.back.length > 30 ? '...' : '') + '</div>';
      html += '</div>';
      html += '<div class="admin-item-actions">';
      html += '<button class="btn btn-sm" onclick="adminModal.open(\'flashcard\', ' + i + ', adminForm.id)">✏️</button>';
      html += '<button class="btn btn-sm btn-red" onclick="adminModal.confirmDelete(\'flashcard\', ' + i + ')">✕</button>';
      html += '</div>';
      html += '</div>';
    }
  }
  html += '</div>';
  html += '<button class="btn btn-sm btn-alt" onclick="adminModal.open(\'flashcard\', -1, adminForm.id)">+ Add Flashcard</button>';
  break;
}
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): convert flashcards step to list view"
```

---

## Task 6: Add Flashcard Modal Form

**Files:**
- Modify: `web/index.html` — add `adminModal.openFlashcardModal()` and `adminModal.saveFlashcard()`

**Interfaces:**
- Uses: `adminForm.flashcards` array
- Produces: Flashcard modal with front/back fields

- [ ] **Step 1: Add openFlashcardModal method**

```javascript
openFlashcardModal(index) {
  const isEdit = index >= 0;
  const card = isEdit ? adminForm.flashcards[index] : { front: '', back: '' };
  
  this.titleEl.textContent = isEdit ? 'Edit Flashcard' : 'Add Flashcard';
  
  this.bodyEl.innerHTML = `
    <div class="form-group">
      <label>Front (Question)</label>
      <textarea id="modalFcFront">${escapeHtml(card.front)}</textarea>
    </div>
    <div class="form-group">
      <label>Back (Answer)</label>
      <textarea id="modalFcBack">${escapeHtml(card.back)}</textarea>
    </div>
  `;
  
  this.footerEl.innerHTML = `
    <button class="btn btn-sm" onclick="adminModal.close()">Cancel</button>
    <button class="btn btn-sm btn-primary" onclick="adminModal.saveFlashcard(${index})">Save</button>
  `;
},
```

- [ ] **Step 2: Add saveFlashcard method**

```javascript
saveFlashcard(index) {
  const front = document.getElementById('modalFcFront').value.trim();
  const back = document.getElementById('modalFcBack').value.trim();
  
  if (!front || !back) {
    alert('Both front and back are required');
    return;
  }
  
  if (!adminForm.flashcards) adminForm.flashcards = [];
  
  if (index >= 0) {
    adminForm.flashcards[index] = { front, back };
  } else {
    adminForm.flashcards.push({ front, back });
  }
  
  adminDirty = true;
  this.close();
  renderAdminLesson();
}
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): add flashcard modal form"
```

---

## Task 7: Add Coding Modal (Edit Only)

**Files:**
- Modify: `web/index.html` — add coding modal

**Interfaces:**
- Produces: Coding edit modal with prompt/starter/solution fields

- [ ] **Step 1: Add openCodingModal method**

```javascript
openCodingModal() {
  const coding = adminForm.coding || { prompt: '', starter: '', solution: '' };
  
  this.titleEl.textContent = 'Edit Coding Exercise';
  
  this.bodyEl.innerHTML = `
    <div class="form-group">
      <label>Prompt</label>
      <textarea id="modalCodPrompt" style="height: 80px">${escapeHtml(coding.prompt)}</textarea>
    </div>
    <div class="form-group">
      <label>Starter Code</label>
      <textarea id="modalCodStarter" style="height: 100px; font-family: monospace;">${escapeHtml(coding.starter)}</textarea>
    </div>
    <div class="form-group">
      <label>Solution</label>
      <textarea id="modalCodSolution" style="height: 100px; font-family: monospace;">${escapeHtml(coding.solution)}</textarea>
    </div>
  `;
  
  this.footerEl.innerHTML = `
    <button class="btn btn-sm" onclick="adminModal.close()">Cancel</button>
    <button class="btn btn-sm btn-primary" onclick="adminModal.saveCoding()">Save</button>
  `;
},
```

- [ ] **Step 2: Add saveCoding method**

```javascript
saveCoding() {
  const coding = {
    prompt: document.getElementById('modalCodPrompt').value,
    starter: document.getElementById('modalCodStarter').value,
    solution: document.getElementById('modalCodSolution').value
  };
  
  adminForm.coding = coding;
  adminDirty = true;
  this.close();
  renderAdminLesson();
}
```

- [ ] **Step 3: Update coding step in renderAdminStep**

Find the coding case and replace with a card + edit button:

```javascript
case 'coding': {
  html += '<div class="v1-ename">Step 4: Coding</div>';
  const coding = adminForm.coding || {};
  html += '<div class="admin-item-card" style="margin-bottom: 12px;">';
  html += '<div class="admin-item-content">';
  html += '<div class="admin-item-meta">Prompt: ' + escapeHtml((coding.prompt || '').substring(0, 60)) + '</div>';
  html += '<div class="admin-item-meta">Starter: ' + escapeHtml((coding.starter || '').substring(0, 40)) + '</div>';
  html += '</div>';
  html += '<button class="btn btn-sm" onclick="adminModal.open(\'coding\', 0, adminForm.id)">✏️</button>';
  html += '</div>';
  break;
}
```

- [ ] **Commit**

```bash
git add web/index.html
git commit -m "feat(admin): add coding modal for editing"
```

---

## Task 8: Verify and Test

**Files:**
- None (verification only)

- [ ] **Step 1: Build and run**

```bash
cd /home/hello/BrutforseLearning && go build -o brutforse_binary main.go
./brutforse_binary &
```

- [ ] **Step 2: Open admin panel and verify**

1. Navigate to admin tab
2. Select a lesson with quiz questions
3. Go to Quiz step - should see list of question cards with Edit/Delete
4. Click Edit on a question - modal should open with pre-filled form
5. Click Add Question - modal should open empty
6. Test delete confirmation
7. Repeat for Flashcards step
8. Test Coding edit modal

- [ ] **Step 3: Run tests**

```bash
go test -v
```

---

## Verification Checklist

- [ ] Quiz: List view shows all questions with Edit/Delete buttons
- [ ] Quiz: Edit opens modal with correct data
- [ ] Quiz: Add opens modal with empty form
- [ ] Quiz: Delete shows confirmation and removes item
- [ ] Flashcards: List view shows all cards with Edit/Delete buttons
- [ ] Flashcards: Edit opens modal with correct front/back
- [ ] Flashcards: Add opens modal with empty form
- [ ] Flashcards: Delete works
- [ ] Coding: Card shows prompt/starter preview with Edit button
- [ ] Coding: Edit opens modal with all three fields
- [ ] All saves persist to backend
- [ ] Tests pass
