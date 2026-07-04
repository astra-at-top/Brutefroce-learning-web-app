package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	dir, err := os.MkdirTemp("", "bf-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	dbPath := filepath.Join(dir, "test.db")
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	initSchema()
	seedData()
}

func TestSm2Calculate(t *testing.T) {
	tests := []struct {
		name         string
		ease         float64
		interval     int
		reps         int
		quality      int
		wantEaseMin  float64
		wantInterval int
	}{
		{"first review good", 2.5, 1, 0, 3, 1.3, 1},
		{"first review easy", 2.5, 1, 0, 5, 2.5, 1},
		{"first review hard", 2.5, 1, 0, 1, 1.3, 1},
		{"repeat good", 2.5, 1, 1, 3, 1.3, 3},
		{"repeat easy", 2.5, 3, 2, 5, 2.5, 8},
		{"quality 0 resets", 2.5, 30, 5, 0, 1.3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm2Calculate(tt.ease, tt.interval, tt.reps, tt.quality)
			if result.Ease < tt.wantEaseMin {
				t.Errorf("sm2Calculate() ease = %v, want >= %v", result.Ease, tt.wantEaseMin)
			}
			if result.Interval != tt.wantInterval {
				t.Errorf("sm2Calculate() interval = %v, want %v", result.Interval, tt.wantInterval)
			}
		})
	}
}

func TestFlashcardsAPI_Schedule(t *testing.T) {
	setupTestDB(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/flashcards/", handleFlashcards)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("returns flashcards for known lesson", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/flashcards/vars-types/schedule")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var cards []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
			t.Fatal(err)
		}
		if len(cards) == 0 {
			t.Fatal("expected flashcards for vars-types, got empty array")
		}

		card := cards[0]
		if _, ok := card["front"]; !ok {
			t.Error("card missing 'front' field")
		}
		if _, ok := card["back"]; !ok {
			t.Error("card missing 'back' field")
		}
		if _, ok := card["ease"]; !ok {
			t.Error("card missing 'ease' field")
		}
	})

	t.Run("returns empty array for unknown lesson", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/flashcards/nonexistent/schedule")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var cards []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
			t.Fatal(err)
		}
		if len(cards) != 0 {
			t.Errorf("expected empty array, got %d items", len(cards))
		}
	})
}

func TestFlashcardsAPI_Review(t *testing.T) {
	setupTestDB(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/flashcards/", handleFlashcards)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	body := `{"card_index":0,"quality":3}`
	resp, err := http.Post(ts.URL+"/api/flashcards/vars-types/review",
		"application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if _, ok := result["ease"]; !ok {
		t.Error("review response missing 'ease'")
	}
	if _, ok := result["interval"]; !ok {
		t.Error("review response missing 'interval'")
	}
	if _, ok := result["next_review"]; !ok {
		t.Error("review response missing 'next_review'")
	}
}

func TestLessonsAPI(t *testing.T) {
	setupTestDB(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/lessons/", handleLessons)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("returns lesson data", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/lessons/vars-types")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var lesson map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&lesson); err != nil {
			t.Fatal(err)
		}
		if lesson["id"] != "vars-types" {
			t.Errorf("expected id=vars-types, got %v", lesson["id"])
		}
	})

	t.Run("returns quiz questions", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/lessons/vars-types/quiz")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var questions []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&questions); err != nil {
			t.Fatal(err)
		}
		if len(questions) == 0 {
			t.Fatal("expected quiz questions for vars-types, got empty")
		}
	})

	t.Run("returns coding data", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/lessons/vars-types/coding")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal(err)
		}
		if _, ok := data["prompt"]; !ok {
			t.Error("coding data missing 'prompt'")
		}
	})
}

func TestFrontend_HasRequiredFunctions(t *testing.T) {
	htmlContent, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlContent)

	requiredFunctions := []string{
		"function renderModuleTabs",
		"function bindLearnEvents",
		"function loadLearnModule",
		"function bindModuleActions",
		"function initFcSchedule",
	}

	for _, fn := range requiredFunctions {
		if !strings.Contains(html, fn) {
			t.Errorf("Missing required function: %s", fn)
		}
	}
}

func TestFrontend_ModuleTabActiveClassUpdate(t *testing.T) {
	htmlContent, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlContent)

	if strings.Contains(html, "learnState.activeModule = el.dataset.module") {
		idx := strings.Index(html, "learnState.activeModule = el.dataset.module")
		chunk := html[idx : idx+350]
		// The word "active" alone is a false positive (matches activeModule).
		// We need classList.add/remove/toggle, className, or classList.toggle('active')
		hasDOMClassToggle := strings.Contains(chunk, "classList.toggle") ||
			strings.Contains(chunk, "classList.add") ||
			strings.Contains(chunk, "className") ||
			strings.Contains(chunk, "classList.remove")
		if !hasDOMClassToggle {
			t.Error("FAIL: Module tab click handler does NOT update the 'active' CSS class on tab DOM elements.\n" +
				"Bug: clicking a module tab updates learnState.activeModule but never toggles the 'active' CSS class\n" +
				"on .module-tab elements, so the visual active state never changes.")
		}
	} else {
		t.Error("Could not find module tab click handler")
	}
}
