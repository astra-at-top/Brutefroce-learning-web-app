package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed web/*
var webFS embed.FS

var db *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	os.MkdirAll(dataDir, 0755)

	var err error
	db, err = sql.Open("sqlite", filepath.Join(dataDir, "brutforse.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initSchema()
	seedData()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/courses", handleCourses)
	mux.HandleFunc("/api/lessons/", handleLessons)
	mux.HandleFunc("/api/progress", handleProgress)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/api/run", handleRunCode)
	mux.HandleFunc("/api/flashcards/", handleFlashcards)
	mux.HandleFunc("/api/powerscore", handlePowerScore)
	mux.HandleFunc("/api/sync/export", handleSyncExport)
	mux.HandleFunc("/api/sync/import", handleSyncImport)

	// Admin endpoints
	mux.HandleFunc("/api/admin/courses", handleAdminCourses)
	mux.HandleFunc("/api/admin/courses/", handleAdminCourseByID)
	mux.HandleFunc("/api/admin/lessons", handleAdminLessons)
	mux.HandleFunc("/api/admin/lessons/", handleAdminLessonByID)
	mux.HandleFunc("/api/admin/import", handleAdminImport)

	subFS, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	log.Printf("BruteForce Learning running on :%s (data: %s)", port, dataDir)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func initSchema() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS lessons (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			notes TEXT DEFAULT '',
			mnemonic TEXT DEFAULT '',
			status TEXT DEFAULT 'locked',
			progress INTEGER DEFAULT 0,
			type TEXT DEFAULT 'lesson',
			quiz_count INTEGER DEFAULT 0,
			course_id TEXT DEFAULT '',
			quiz_data TEXT DEFAULT '[]',
			coding_data TEXT DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS courses (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			parent_id TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS flashcard_reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			lesson_id TEXT NOT NULL,
			front TEXT NOT NULL,
			back TEXT NOT NULL,
			ease REAL DEFAULT 2.5,
			interval_days INTEGER DEFAULT 1,
			repetitions INTEGER DEFAULT 0,
			next_review TEXT DEFAULT '',
			last_reviewed TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS flashcard_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			lesson_id TEXT NOT NULL,
			front TEXT NOT NULL,
			back TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS progress (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS achievements (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT DEFAULT '',
			unlocked INTEGER DEFAULT 0,
			unlocked_at TEXT DEFAULT ''
		)`,
	}
	for _, s := range migrations {
		if _, err := db.Exec(s); err != nil {
			log.Fatal("schema:", err)
		}
	}

	// Migrate existing DB: add columns if missing
	for _, col := range []string{"quiz_data", "coding_data"} {
		var found int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('lessons') WHERE name=?", col).Scan(&found)
		if found == 0 {
			db.Exec("ALTER TABLE lessons ADD COLUMN " + col + " TEXT DEFAULT '[]'")
		}
	}
	var foundCourse int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('lessons') WHERE name='course_id'").Scan(&foundCourse)
	if foundCourse == 0 {
		db.Exec("ALTER TABLE lessons ADD COLUMN course_id TEXT DEFAULT ''")
	}
}

func seedData() {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&count)
	if count > 0 {
		return
	}

	courses := []struct{ ID, Name, ParentID string; Sort int }{
		{"javascript", "JavaScript", "", 0},
		{"react", "React", "", 1},
		{"git", "Git", "", 2},
		{"node", "Node.js", "", 3},
	}
	for _, c := range courses {
		db.Exec("INSERT INTO courses (id,name,parent_id,sort_order) VALUES (?,?,?,?)", c.ID, c.Name, c.ParentID, c.Sort)
	}

	var lessonCount int
	db.QueryRow("SELECT COUNT(*) FROM lessons").Scan(&lessonCount)
	if lessonCount > 0 {
		return
	}

	type lessonSeed struct {
		ID, Title, Notes, Mnemonic, Status, LType, CourseID string
		Progress, QuizCount                                 int
		QuizData, CodingData                                string
	}
	lessons := []lessonSeed{
		// ========== JavaScript ==========
		{"vars-types", "Variables & Types", "JavaScript provides three variable declaration keywords. `const` prevents reassignment but object mutations are allowed. `let` is block-scoped and reassignable. `var` is function-scoped, hoisted to undefined — avoid it in modern code. Seven primitive types: string, number, bigint, boolean, null, undefined, symbol. `typeof null` returns 'object', a legacy bug. Use `===` for strict equality to avoid type coercion surprises.", "SNOB NUN — String, Number, Object (typeof null), Boolean, Null, Undefined", "done", "lesson", "javascript", 100, 4,
			`[{"question":"Which keyword declares a block-scoped variable that cannot be reassigned?","options":["let","const","var","static"],"correct_index":1},{"question":"What does typeof null return?","options":["\"null\"","\"undefined\"","\"object\"","\"boolean\""],"correct_index":2},{"question":"Which is NOT a JavaScript primitive type?","options":["string","object","symbol","bigint"],"correct_index":1},{"question":"What is the difference between == and ===?","options":["No difference","== coerces types, === does not","=== coerces types, == does not","== is faster"],"correct_index":1}]`,
			`{"prompt":"Declare a const name with value \"Alice\". Declare a let age with 30. Write a function strictEq that checks strict equality of two args.","starter":"const name = \"Alice\";\nlet age = 30;\n","solution":"const name = \"Alice\";\nlet age = 30;\nfunction strictEq(a, b) {\n  return a === b;\n}"}`,
		},
		{"functions-scope", "Functions & Scope", "Functions are first-class citizens — assignable, passable, returnable. Three declaration forms: function declaration (hoisted), function expression, arrow function. Arrow functions lack their own `this` and `arguments` — they inherit from the enclosing scope. Three scope levels: global, function (var), block (let/const). Hoisting moves declarations to the top; `let`/`const` sit in the Temporal Dead Zone until initialisation.", "GAB — Global, Arrow (inherits this), Block scopes", "done", "lesson", "javascript", 100, 3,
			`[{"question":"What distinguishes arrow functions from regular functions?","options":["Arrow functions are faster","Arrow functions lack their own this","Arrow functions are always async","No difference"],"correct_index":1},{"question":"What gets hoisted to undefined?","options":["const x = 1","let x = 1","var x = 1","function x(){}"],"correct_index":2},{"question":"What happens if you access a let variable before its declaration?","options":["Returns undefined","Throws ReferenceError (TDZ)","Throws SyntaxError","Returns null"],"correct_index":1}]`,
			`{"prompt":"Write an arrow function multiply(a, b) that returns a*b. Write a function declaration isEven(n) that returns true if n is even.","starter":"const multiply = ","solution":"const multiply = (a, b) => a * b;\nfunction isEven(n) { return n % 2 === 0; }"}`,
		},
		{"arrays-objects", "Arrays & Objects", "Arrays are zero-indexed, dynamically-sized lists. Key methods: push/pop (stack ops), shift/unshift (queue ops), map (transform each), filter (select by predicate), reduce (fold), find, some, every. Objects are unordered key-value stores. Access with dot or bracket notation. Destructuring unpacks values: const {name, age} = person. Spread (...) clones or merges. `const` on an object prevents reassignment, not mutation.", "MAP-FIND — Map, Array methods, Push/Pop, Filter, IN, Destructure", "in_progress", "lesson", "javascript", 60, 3,
			`[{"question":"Which array method creates a new array with transformed elements?","options":["forEach()","map()","filter()","reduce()"],"correct_index":1},{"question":"How do you add an element to the end of an array?","options":["push()","pop()","shift()","unshift()"],"correct_index":0},{"question":"What does const obj = {a:1}; obj.a = 2; do?","options":["Throws TypeError","Mutates obj to {a:2}","Does nothing","Creates a new object"],"correct_index":1}]`,
			`{"prompt":"Given array nums = [1,2,3], push 4, then map to double each element. Destructure {name, age} from object person = {name:\"Alice\", age:30}.","starter":"let nums = [1, 2, 3];\nconst person = {name: \"Alice\", age: 30};","solution":"let nums = [1, 2, 3];\nnums.push(4);\nconst doubled = nums.map(n => n * 2);\nconst {name, age} = person;"}`,
		},
		{"closures", "Closures", "A closure is a function bundled with references to its outer lexical scope. When an inner function accesses variables from an outer function after that outer function has returned, you're seeing a closure in action. Practical uses: data privacy (module pattern), function factories, event handlers, memoization. Every function in JavaScript is a closure — they all capture the scope in which they were defined.", "BRIBE — Backpack, Retains, Inner, Bound, Execution context", "available", "lesson", "javascript", 0, 3,
			`[{"question":"What is a closure?","options":["A function that runs immediately","A function with access to its outer scope after the outer function returns","A function that takes no arguments","A built-in JS method"],"correct_index":1},{"question":"What does this code log? for(var i=0;i<3;i++){setTimeout(()=>console.log(i),100)}","options":["0,1,2","3,3,3","undefined","1,2,3"],"correct_index":1},{"question":"How do you fix the var loop closure bug?","options":["Use let instead of var","Add a timeout of 0","Use console.log(i) directly","Use arrow functions"],"correct_index":0}]`,
			`{"prompt":"Write a function createCounter that returns an object with increment() and getCount() methods. The count variable should be private (closure).","starter":"function createCounter() {","solution":"function createCounter() {\n  let count = 0;\n  return {\n    increment: () => ++count,\n    getCount: () => count\n  };\n}"}`,
		},
		{"js-promises", "Promises & Async", "A Promise represents a value that may be available now, later, or never. States: pending, fulfilled, rejected. Chain with `.then()` and catch errors with `.catch()`. `async` functions always return a Promise; `await` pauses execution until the Promise settles. `Promise.all` runs promises in parallel and fails fast. `Promise.allSettled` waits for all regardless of rejection. Always handle rejections — unhandled promises crash Node and cause memory leaks in browsers.", "PAC — Pending, All, Catch — Promise, Async, Chain", "locked", "lesson", "javascript", 0, 3,
			`[{"question":"What state is a Promise in after calling resolve()?","options":["pending","fulfilled","rejected","settled"],"correct_index":1},{"question":"What does async function always return?","options":["A value","A Promise","Undefined","An Observable"],"correct_index":1},{"question":"Which method runs multiple promises and fails if any one rejects?","options":["Promise.all()","Promise.allSettled()","Promise.race()","Promise.any()"],"correct_index":0}]`,
			`{"prompt":"Write an async function fetchUser(id) that fetches '/api/user/'+id and returns JSON. Handle errors with try/catch.","starter":"async function fetchUser(id) {","solution":"async function fetchUser(id) {\n  try {\n    const res = await fetch('/api/user/' + id);\n    if (!res.ok) throw new Error('fetch failed');\n    return await res.json();\n  } catch (err) {\n    console.error(err);\n    return null;\n  }\n}"}`,
		},

		// ========== React ==========
		{"jsx-components", "JSX & Components", "JSX is a JavaScript syntax extension that looks like HTML but compiles to React.createElement calls. Every component is a function returning JSX. Props are read-only arguments passed from parent to child. Children are accessed via the `children` prop. Fragment (<>) lets you return multiple elements without a wrapper div. Component names must be capitalised to distinguish from native HTML elements.", "C-FP — Capitalise, Fragment, Props", "available", "lesson", "react", 0, 3,
			`[{"question":"What does JSX compile to?","options":["HTML strings","React.createElement calls","DocumentFragments","Template literals"],"correct_index":1},{"question":"Why must component names start with a capital letter?","options":["It's a style preference","React uses lowercase for built-in HTML elements","JSX requires it for props","It's optional"],"correct_index":1},{"question":"How do you pass data from a parent to a child component?","options":["State","Props","Context","Refs"],"correct_index":1}]`,
			`{"prompt":"Write a Greeting component that takes a name prop and returns <h1>Hello, {name}!</h1>. Then write a App component that renders Greeting with name=\"World\".","starter":"function Greeting({ name }) {","solution":"function Greeting({ name }) {\n  return <h1>Hello, {name}!</h1>;\n}\nfunction App() {\n  return <Greeting name=\"World\" />;\n}"}`,
		},
		{"usestate-hook", "useState Hook", "useState is the primary React hook for adding state to functional components. It returns a pair: the current state value and a setter function. The setter can take a new value or a functional updater `prev => prev + 1`. State updates trigger re-renders. Never mutate state directly — always use the setter. For objects/arrays, spread the previous state: `setUser(prev => ({...prev, name: newName}))`.", "MRS — Mutation triggers Re-render via Setter", "locked", "lesson", "react", 0, 3,
			`[{"question":"What does useState return?","options":["A single value","An array with state and setter","An object with state and setter","A Promise"],"correct_index":1},{"question":"When should you use the functional updater pattern?","options":["Always","When new state depends on previous state","Never","Only with objects"],"correct_index":1},{"question":"What happens if you mutate state directly (state.count = 1)?","options":["Component re-renders with new value","React throws a warning but works","Nothing — state doesn't change or re-render","The app crashes"],"correct_index":2}]`,
			`{"prompt":"Write a Counter component with useState. Show count and buttons to increment/decrement. Use functional updater.","starter":"function Counter() {\n  const [count, setCount] = ","solution":"function Counter() {\n  const [count, setCount] = useState(0);\n  return (\n    <div>\n      <p>{count}</p>\n      <button onClick={() => setCount(c => c + 1)}>+</button>\n      <button onClick={() => setCount(c => c - 1)}>-</button>\n    </div>\n  );\n}"}`,
		},
		{"useeffect-hook", "useEffect Hook", "useEffect lets you perform side effects in function components: data fetching, subscriptions, DOM manipulation, timers. The first argument is the effect function; the second is the dependency array. Effects run after every render by default; empty deps `[]` runs once on mount; returning a cleanup function runs on unmount or before re-run. Missing dependencies are the #1 source of React bugs — the linter (react-hooks/exhaustive-deps) catches them.", "DMC — Dependencies, Mount, Cleanup", "locked", "lesson", "react", 0, 3,
			`[{"question":"When does useEffect run by default?","options":["Only on mount","After every render","Only on unmount","Never"],"correct_index":1},{"question":"What does the cleanup function in useEffect do?","options":["Removes the component","Runs before the effect re-runs or unmounts","Optimises performance","Is optional and rarely used"],"correct_index":1},{"question":"What happens if you omit the dependency array?","options":["Effect runs once","Effect runs after every render","Effect never runs","React throws an error"],"correct_index":1}]`,
			`{"prompt":"Write a Timer component that counts seconds using useEffect with setInterval. Clean up the interval on unmount.","starter":"function Timer() {\n  const [seconds, setSeconds] = useState(0);","solution":"function Timer() {\n  const [seconds, setSeconds] = useState(0);\n  useEffect(() => {\n    const id = setInterval(() => setSeconds(s => s + 1), 1000);\n    return () => clearInterval(id);\n  }, []);\n  return <p>{seconds}s</p>;\n}"}`,
		},
		{"data-fetching", "Data Fetching", "Fetch data in useEffect with an async function defined inside the effect — cannot pass async directly as the effect function. Handle three states: loading, error, data. Use AbortController to cancel stale requests on unmount. Extract reusable logic into custom hooks: `useFetch(url)` that returns `{data, loading, error}`. React Query / TanStack Query solves caching, refetching, and pagination out of the box for production apps.", "LEC — Loading, Error, Cancel (AbortController)", "locked", "lesson", "react", 0, 3,
			`[{"question":"How do you use async/await inside useEffect?","options":["Make the effect function async","Define an async function inside and call it","Use .then() instead","It's not allowed"],"correct_index":1},{"question":"What problem does AbortController solve?","options":["Makes fetch faster","Prevents state updates on unmounted components","Caches responses","Parallelises requests"],"correct_index":1},{"question":"What pattern returns {data, loading, error} from a fetch?","options":["HOC","Custom hook","Render prop","Context"],"correct_index":1}]`,
			`{"prompt":"Write a custom hook useFetch(url) that returns {data, loading, error}. Handle loading, success, and error states.","starter":"function useFetch(url) {","solution":"function useFetch(url) {\n  const [data, setData] = useState(null);\n  const [loading, setLoading] = useState(true);\n  const [error, setError] = useState(null);\n  useEffect(() => {\n    const ac = new AbortController();\n    setLoading(true);\n    fetch(url, {signal: ac.signal})\n      .then(r => r.json())\n      .then(setData)\n      .catch(setError)\n      .finally(() => setLoading(false));\n    return () => ac.abort();\n  }, [url]);\n  return {data, loading, error};\n}"}`,
		},

		// ========== Git ==========
		{"git-basics", "Git Basics", "Git is a distributed version control system that tracks changes as snapshots (commits). The three-tree architecture: Working Directory (files you edit), Staging Area (index — files marked for next commit), Repository (.git — committed history). Essential flow: `git init` → `git add .` → `git commit -m \"msg\"`. `git status` shows the state. `git log` shows history. `git diff` shows unstaged changes. `.gitignore` prevents tracking generated files.", "WRS — Working, Staging, Repository", "available", "lesson", "git", 0, 3,
			`[{"question":"What command moves files from working directory to staging area?","options":["git commit","git add","git push","git status"],"correct_index":1},{"question":"Where does Git store committed snapshots?","options":["Working directory","Staging area","Repository (.git)","Remote server"],"correct_index":2},{"question":"What does git diff show?","options":["Differences between commits","Unstaged changes","Staged changes","Branch differences"],"correct_index":1}]`,
			`{"prompt":"Write the sequence of commands to initialise a repo, add all files, commit with message \"initial commit\", and check status.","starter":"# initialise a git repo","solution":"git init\ngit add .\ngit commit -m \"initial commit\"\ngit status"}`,
		},
		{"git-branching", "Branching & Merging", "Branches are lightweight pointers to commits, making them cheap to create. `git branch <name>` creates; `git checkout <name>` or `git switch <name>` moves to it. `git merge <branch>` integrates changes. Fast-forward merge occurs when the target hasn't diverged; a merge commit is created when it has. Merge conflicts happen when the same file region is modified in both branches — resolve manually, then `git add` and `git commit`. `git rebase` rewrites history. `git stash` temporarily shelves changes.", "MCC — Merge, Conflict, Checkout", "locked", "lesson", "git", 0, 3,
			`[{"question":"What is a fast-forward merge?","options":["A merge that deletes the source branch","A merge where the target hasn't diverged from source","A merge with no conflicts","A merge that runs automatically"],"correct_index":1},{"question":"How do you resolve a merge conflict?","options":["Delete the conflicting file","Edit the file to resolve markers, then git add and commit","Run git conflict --resolve","Rebase instead"],"correct_index":1},{"question":"Which command shelves uncommitted changes temporarily?","options":["git save","git stash","git shelve","git hold"],"correct_index":1}]`,
			`{"prompt":"Write the sequence: create a branch 'feature', switch to it, stash current changes, commit on feature, switch back to main, merge feature.","starter":"# branch workflow","solution":"git branch feature\ngit switch feature\ngit stash\ngit add . && git commit -m \"feat: add feature\"\ngit switch main\ngit merge feature"}`,
		},

		// ========== Node.js ==========
		{"node-intro", "Node.js Basics", "Node.js is a JavaScript runtime built on Chrome's V8 engine. It uses an event-driven, non-blocking I/O model. Key built-in modules: `fs` (file system), `http` (server), `path` (path manipulation), `os` (system info). `require` vs `import`: CommonJS (require/module.exports) is the default; ES modules (import/export) require `\"type\": \"module\"` in package.json. `npm` manages dependencies. `node_modules` should never be committed — use `.gitignore`.", "NEV — Non-blocking, Event-driven, V8", "locked", "lesson", "node", 0, 2,
			`[{"question":"What does non-blocking I/O mean in Node?","options":["The CPU never waits","Operations run in the background and callback when done","Threads are never blocked","I/O operations are synchronous"],"correct_index":1},{"question":"Which module creates an HTTP server in Node?","options":["net","http","server","express"],"correct_index":1}]`,
			`{}`,
		},
	}
	for _, l := range lessons {
		db.Exec("INSERT INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			l.ID, l.Title, l.Notes, l.Mnemonic, l.Status, l.LType, l.CourseID, l.Progress, l.QuizCount, l.QuizData, l.CodingData)
	}

	type fcSeed struct{ LessonID, Front, Back string }
	fcs := []fcSeed{
		// JavaScript
		{"vars-types", "What 3 ways can you declare variables in JS?", "const, let, var"},
		{"vars-types", "How many primitive types in JS? Name them.", "7: string, number, bigint, boolean, null, undefined, symbol"},
		{"vars-types", "What does typeof null return?", "\"object\" (a legacy bug)"},
		{"vars-types", "When does JS coerce types?", "With == operator and template literals"},
		{"functions-scope", "What is hoisting?", "Declarations are moved to the top of scope during compilation"},
		{"functions-scope", "Difference between function declaration and expression?", "Declaration is hoisted; expression is not"},
		{"functions-scope", "Do arrow functions have their own 'this'?", "No — they inherit from enclosing scope"},
		{"arrays-objects", "What does .map() return?", "A new array with each element transformed"},
		{"arrays-objects", "What does .reduce() do?", "Folds an array into a single value"},
		{"arrays-objects", "How do you destructure an object?", "const { key } = obj"},
		{"closures", "What is a closure?", "A function that retains access to its outer scope after the outer function returns"},
		{"closures", "What 3 scopes does a closure access?", "Local, outer function, and global"},
		{"closures", "How do you fix the var loop closure bug?", "Use let (block scoping) instead of var"},
		{"js-promises", "What are the 3 states of a Promise?", "Pending, fulfilled, rejected"},
		{"js-promises", "What does Promise.all do?", "Runs promises in parallel; rejects fast if any rejects"},
		// React
		{"jsx-components", "What does JSX compile to?", "React.createElement calls"},
		{"jsx-components", "Can you return multiple elements from a component?", "Yes — wrap in Fragment <>...</>"},
		{"usestate-hook", "What pattern prevents stale state in useState?", "Functional updater: setCount(c => c + 1)"},
		{"usestate-hook", "Does setState merge or replace state?", "Replace — spread previous for objects"},
		{"useeffect-hook", "What is the #1 React bug source?", "Missing useEffect dependencies"},
		{"useeffect-hook", "What does the cleanup function do?", "Runs before effect re-runs or component unmounts"},
		{"data-fetching", "How to cancel a fetch on unmount?", "AbortController + signal"},
		{"data-fetching", "What 3 states should a fetch hook manage?", "Loading, error, data"},
		// Git
		{"git-basics", "What are the three Git trees?", "Working directory, staging area, repository"},
		{"git-basics", "What does .gitignore do?", "Prevents tracking specified files"},
		{"git-branching", "What is a fast-forward merge?", "Target branch hasn't diverged — pointer moves forward"},
		{"git-branching", "How to temporarily save uncommitted work?", "git stash"},
		// Node.js
		{"node-intro", "What engine powers Node.js?", "Chrome V8"},
		{"node-intro", "What does non-blocking I/O mean?", "Operations run in background, callback when done"},
	}
	for _, f := range fcs {
		db.Exec("INSERT INTO flashcard_data (lesson_id,front,back) VALUES (?,?,?)", f.LessonID, f.Front, f.Back)
		db.Exec("INSERT INTO flashcard_reviews (lesson_id,front,back,ease,interval_days,repetitions,next_review) VALUES (?,?,?,2.5,1,0,?)",
			f.LessonID, f.Front, f.Back, time.Now().Format(time.RFC3339))
	}

	db.Exec("INSERT OR IGNORE INTO progress (key,value) VALUES ('xp','320'),('streak','5'),('mastery','42'),('solved','18'),('xp_max','500'),('total_hours','12.5')")
	for _, a := range []struct{ ID, Name, Icon string }{
		{"first-quiz", "Quick Draw", "🎯"}, {"streak-3", "Heat Wave", "🔥"}, {"streak-7", "Inferno", "💥"},
		{"code-5", "Keyboard Warrior", "⌨️"}, {"quiz-10", "Socratic", "🏛️"}, {"all-topics", "Cartographer", "🗺️"},
		{"speed-demon", "Speed Demon", "⚡"}, {"contributor", "Contributor", "🤝"},
	} {
		db.Exec("INSERT OR IGNORE INTO achievements (id,name,icon,unlocked) VALUES (?,?,?,0)", a.ID, a.Name, a.Icon)
	}
}

func jsonResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func getProgress(key string) string {
	var val string
	db.QueryRow("SELECT value FROM progress WHERE key=?", key).Scan(&val)
	return val
}

func setProgress(key, val string) {
	db.Exec("INSERT OR REPLACE INTO progress (key,value) VALUES (?,?)", key, val)
}

// ======== COURSES (for Learn tab) ========
func handleCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	jsonResp(w, getCourseTree())
}

func getCourseTree() []map[string]interface{} {
	rows, err := db.Query("SELECT id, name, parent_id FROM courses WHERE parent_id='' ORDER BY sort_order")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var roots []map[string]interface{}
	for rows.Next() {
		var id, name, parentID string
		rows.Scan(&id, &name, &parentID)
		roots = append(roots, buildCourseNode(id))
	}
	return roots
}

func buildCourseNode(courseID string) map[string]interface{} {
	var id, name string
	db.QueryRow("SELECT id, name FROM courses WHERE id=?", courseID).Scan(&id, &name)

	node := map[string]interface{}{
		"id": id, "name": name, "expanded": id == "react",
	}

	// Get children courses
	childCourses, _ := db.Query("SELECT id FROM courses WHERE parent_id=? ORDER BY sort_order", courseID)
	var children []map[string]interface{}
	if childCourses != nil {
		defer childCourses.Close()
		for childCourses.Next() {
			var cid string
			childCourses.Scan(&cid)
			children = append(children, buildCourseNode(cid))
		}
	}

	// Get lessons for this course
	lessonRows, _ := db.Query("SELECT id, title, status, progress, type FROM lessons WHERE course_id=? ORDER BY id", courseID)
	if lessonRows != nil {
		defer lessonRows.Close()
		for lessonRows.Next() {
			var lid, title, status, ltype string
			var progress int
			lessonRows.Scan(&lid, &title, &status, &progress, &ltype)
			children = append(children, map[string]interface{}{
				"id": lid, "name": title, "type": ltype,
				"status": status, "progress": progress,
			})
		}
	}

	if len(children) > 0 {
		node["children"] = children
		node["status"] = calcParentStatus(children)
	} else {
		node["status"] = "locked"
		node["type"] = "lesson"
	}

	return node
}

func calcParentStatus(children []map[string]interface{}) string {
	var done, total int
	for _, c := range children {
		total++
		if s, ok := c["status"].(string); ok && s == "done" {
			done++
		}
	}
	if done == total {
		return "done"
	}
	if done > 0 {
		return "in_progress"
	}
	return "locked"
}

// ======== LESSONS (for Learn tab) ========
func handleLessons(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/lessons/")

	if strings.HasSuffix(path, "/quiz") {
		lessonID := strings.TrimSuffix(path, "/quiz")
		jsonResp(w, getQuizQuestions(lessonID))
		return
	}
	if strings.HasSuffix(path, "/coding") {
		lessonID := strings.TrimSuffix(path, "/coding")
		var codingData string
		err := db.QueryRow("SELECT coding_data FROM lessons WHERE id=?", lessonID).Scan(&codingData)
		if err != nil {
			jsonResp(w, map[string]interface{}{})
			return
		}
		var data map[string]interface{}
		json.Unmarshal([]byte(codingData), &data)
		jsonResp(w, data)
		return
	}

	var id, title, notes, mnemonic, status, ltype string
	var progress, quizCount int
	err := db.QueryRow("SELECT id,title,notes,mnemonic,status,type,progress,quiz_count FROM lessons WHERE id=?", path).
		Scan(&id, &title, &notes, &mnemonic, &status, &ltype, &progress, &quizCount)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	jsonResp(w, map[string]interface{}{
		"id": id, "title": title, "notes": notes, "mnemonic": mnemonic,
		"status": status, "type": ltype, "progress": progress, "quiz_count": quizCount,
	})
}

func getQuizQuestions(lessonID string) []map[string]interface{} {
	var quizData string
	err := db.QueryRow("SELECT quiz_data FROM lessons WHERE id=?", lessonID).Scan(&quizData)
	if err != nil {
		return []map[string]interface{}{}
	}
	var questions []map[string]interface{}
	if err := json.Unmarshal([]byte(quizData), &questions); err != nil {
		return []map[string]interface{}{}
	}
	return questions
}

// ======== ADMIN ENDPOINTS ========
func handleAdminCourses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, _ := db.Query("SELECT id, name, parent_id, sort_order FROM courses ORDER BY sort_order")
		defer rows.Close()
		var courses []map[string]interface{}
		for rows.Next() {
			var id, name, parentID string
			var sort int
			rows.Scan(&id, &name, &parentID, &sort)
			courses = append(courses, map[string]interface{}{
				"id": id, "name": name, "parent_id": parentID, "sort_order": sort,
			})
		}
		jsonResp(w, buildAdminTree(courses))
	case "POST":
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			ParentID string `json:"parent_id"`
		}
		json.Unmarshal(body, &req)
		if req.ID == "" {
			req.ID = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
		}
		_, err := db.Exec("INSERT INTO courses (id,name,parent_id) VALUES (?,?,?)", req.ID, req.Name, req.ParentID)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		jsonResp(w, map[string]interface{}{"id": req.ID, "name": req.Name})
	default:
		jsonError(w, "method not allowed", 405)
	}
}

func handleAdminCourseByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/courses/")
	if r.Method == "DELETE" {
		db.Exec("DELETE FROM courses WHERE id=?", path)
		db.Exec("DELETE FROM lessons WHERE course_id=?", path)
		jsonResp(w, map[string]string{"status": "deleted"})
		return
	}
	jsonError(w, "method not allowed", 405)
}

func handleAdminLessons(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Status   string `json:"status"`
		CourseID string `json:"course_id"`
	}
	json.Unmarshal(body, &req)
	if req.ID == "" {
		req.ID = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	}
	_, err := db.Exec("INSERT INTO lessons (id,title,type,status,course_id) VALUES (?,?,?,?,?)",
		req.ID, req.Title, req.Type, req.Status, req.CourseID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResp(w, map[string]interface{}{"id": req.ID, "title": req.Title})
}

func handleAdminLessonByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/lessons/")

	switch r.Method {
	case "GET":
		var id, title, notes, mnemonic, status, ltype, courseID, quizData, codingData string
		var progress, quizCount int
		err := db.QueryRow("SELECT id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data FROM lessons WHERE id=?", path).
			Scan(&id, &title, &notes, &mnemonic, &status, &ltype, &courseID, &progress, &quizCount, &quizData, &codingData)
		if err != nil {
			jsonError(w, "not found", 404)
			return
		}

		var quiz []map[string]interface{}
		json.Unmarshal([]byte(quizData), &quiz)
		var coding map[string]interface{}
		json.Unmarshal([]byte(codingData), &coding)

		fcRows, _ := db.Query("SELECT front, back FROM flashcard_data WHERE lesson_id=? ORDER BY id", path)
		var flashcards []map[string]string
		if fcRows != nil {
			defer fcRows.Close()
			for fcRows.Next() {
				var front, back string
				fcRows.Scan(&front, &back)
				flashcards = append(flashcards, map[string]string{"front": front, "back": back})
			}
		}

		jsonResp(w, map[string]interface{}{
			"id": id, "title": title, "notes": notes, "mnemonic": mnemonic,
			"status": status, "type": ltype, "course_id": courseID,
			"progress": progress, "quiz_count": quizCount,
			"quiz": quiz, "coding": coding, "flashcards": flashcards,
		})

	case "PUT":
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		// Determine what we're updating
		if quiz, ok := data["quiz"]; ok {
			qData, _ := json.Marshal(quiz)
			db.Exec("UPDATE lessons SET quiz_data=? WHERE id=?", string(qData), path)
			if arr, ok := quiz.([]interface{}); ok {
				db.Exec("UPDATE lessons SET quiz_count=? WHERE id=?", len(arr), path)
			}
			jsonResp(w, map[string]string{"status": "quiz saved"})
			return
		}
		if coding, ok := data["coding"]; ok {
			cData, _ := json.Marshal(coding)
			db.Exec("UPDATE lessons SET coding_data=? WHERE id=?", string(cData), path)
			jsonResp(w, map[string]string{"status": "coding saved"})
			return
		}
		if flashcards, ok := data["flashcards"]; ok {
			db.Exec("DELETE FROM flashcard_data WHERE lesson_id=?", path)
			db.Exec("DELETE FROM flashcard_reviews WHERE lesson_id=?", path)
			if arr, ok := flashcards.([]interface{}); ok {
				for _, fc := range arr {
					if m, ok := fc.(map[string]interface{}); ok {
						front, _ := m["front"].(string)
						back, _ := m["back"].(string)
						db.Exec("INSERT INTO flashcard_data (lesson_id,front,back) VALUES (?,?,?)", path, front, back)
						db.Exec("INSERT INTO flashcard_reviews (lesson_id,front,back,ease,interval_days,repetitions,next_review) VALUES (?,?,?,2.5,1,0,?)",
							path, front, back, time.Now().Format(time.RFC3339))
					}
				}
			}
			jsonResp(w, map[string]string{"status": "flashcards saved"})
			return
		}

		// Update lesson fields
		title, _ := data["title"].(string)
		ltype, _ := data["type"].(string)
		status, _ := data["status"].(string)
		notes, _ := data["notes"].(string)
		mnemonic, _ := data["mnemonic"].(string)
		courseID, _ := data["course_id"].(string)

		db.Exec("UPDATE lessons SET title=?,type=?,status=?,notes=?,mnemonic=?,course_id=? WHERE id=?",
			title, ltype, status, notes, mnemonic, courseID, path)
		jsonResp(w, map[string]string{"status": "saved"})

	case "DELETE":
		db.Exec("DELETE FROM lessons WHERE id=?", path)
		db.Exec("DELETE FROM flashcard_data WHERE lesson_id=?", path)
		db.Exec("DELETE FROM flashcard_reviews WHERE lesson_id=?", path)
		jsonResp(w, map[string]string{"status": "deleted"})

	default:
		jsonError(w, "method not allowed", 405)
	}
}

func handleAdminImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "method not allowed", 405)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), 400)
		return
	}

	// Mode 1: Import full courses array (replaces all courses + lessons)
	if coursesRaw, ok := data["courses"]; ok {
		coursesList, ok := coursesRaw.([]interface{})
		if !ok {
			jsonError(w, "courses must be an array", 400)
			return
		}

		// Wipe existing data
		db.Exec("DELETE FROM flashcard_reviews")
		db.Exec("DELETE FROM flashcard_data")
		db.Exec("DELETE FROM lessons")
		db.Exec("DELETE FROM courses")

		courseCount := 0
		lessonCount := 0

		for _, cr := range coursesList {
			c, _ := cr.(map[string]interface{})
			if c == nil {
				continue
			}
			id, _ := c["id"].(string)
			name, _ := c["name"].(string)
			parentID, _ := c["parent_id"].(string)
			sort := 0
			if s, ok := c["sort_order"].(float64); ok {
				sort = int(s)
			}
			if id == "" || name == "" {
				continue
			}

			db.Exec("INSERT INTO courses (id,name,parent_id,sort_order) VALUES (?,?,?,?)", id, name, parentID, sort)
			courseCount++

			// Import lessons for this course
			if lessonsRaw, ok := c["lessons"]; ok {
				lessonsList, ok := lessonsRaw.([]interface{})
				if !ok {
					continue
				}
				for _, lr := range lessonsList {
					lesson, _ := lr.(map[string]interface{})
					if lesson == nil {
						continue
					}
					lid, _ := lesson["id"].(string)
					title, _ := lesson["title"].(string)
					if lid == "" || title == "" {
						continue
					}

					ltype, _ := lesson["type"].(string)
					if ltype == "" {
						ltype = "lesson"
					}
					status, _ := lesson["status"].(string)
					if status == "" {
						status = "locked"
					}
					notes, _ := lesson["notes"].(string)
					mnemonic, _ := lesson["mnemonic"].(string)
					progress := 0
					if p, ok := lesson["progress"].(float64); ok {
						progress = int(p)
					}

					// Quiz data
					quizData := "[]"
					if quizRaw, ok := lesson["quiz"]; ok {
						if qb, err := json.Marshal(quizRaw); err == nil {
							quizData = string(qb)
						}
					}

					// Coding data
					codingData := "{}"
					if codingRaw, ok := lesson["coding"]; ok {
						if cb, err := json.Marshal(codingRaw); err == nil {
							codingData = string(cb)
						}
					}

					// Count quiz questions
					quizCount := 0
					if qArr, ok := lesson["quiz"].([]interface{}); ok {
						quizCount = len(qArr)
					}

					db.Exec("INSERT INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
						lid, title, notes, mnemonic, status, ltype, id, progress, quizCount, quizData, codingData)
					lessonCount++

					// Flashcards
					if fcRaw, ok := lesson["flashcards"]; ok {
						fcList, ok := fcRaw.([]interface{})
						if ok {
							for _, fcr := range fcList {
								fc, _ := fcr.(map[string]interface{})
								if fc == nil {
									continue
								}
								front, _ := fc["front"].(string)
								back, _ := fc["back"].(string)
								if front == "" || back == "" {
									continue
								}
								db.Exec("INSERT INTO flashcard_data (lesson_id,front,back) VALUES (?,?,?)", lid, front, back)
								db.Exec("INSERT INTO flashcard_reviews (lesson_id,front,back,ease,interval_days,repetitions,next_review) VALUES (?,?,?,2.5,1,0,?)",
									lid, front, back, time.Now().Format(time.RFC3339))
							}
						}
					}
				}
			}
		}

		jsonResp(w, map[string]interface{}{
			"status": "ok", "courses_imported": courseCount, "lessons_imported": lessonCount,
		})
		return
	}

	// Mode 2: Import single lesson
	if lessonRaw, ok := data["lesson"]; ok {
		lesson, _ := lessonRaw.(map[string]interface{})
		if lesson == nil {
			jsonError(w, "lesson must be an object", 400)
			return
		}

		lid, _ := lesson["id"].(string)
		title, _ := lesson["title"].(string)
		if lid == "" || title == "" {
			jsonError(w, "lesson id and title are required", 400)
			return
		}

		courseID, _ := lesson["course_id"].(string)
		if courseID == "" {
			if cid, ok := data["course_id"].(string); ok {
				courseID = cid
			}
		}

		ltype, _ := lesson["type"].(string)
		if ltype == "" {
			ltype = "lesson"
		}
		status, _ := lesson["status"].(string)
		if status == "" {
			status = "locked"
		}
		notes, _ := lesson["notes"].(string)
		mnemonic, _ := lesson["mnemonic"].(string)
		progress := 0
		if p, ok := lesson["progress"].(float64); ok {
			progress = int(p)
		}

		quizData := "[]"
		if quizRaw, ok := lesson["quiz"]; ok {
			if qb, err := json.Marshal(quizRaw); err == nil {
				quizData = string(qb)
			}
		}
		codingData := "{}"
		if codingRaw, ok := lesson["coding"]; ok {
			if cb, err := json.Marshal(codingRaw); err == nil {
				codingData = string(cb)
			}
		}
		quizCount := 0
		if qArr, ok := lesson["quiz"].([]interface{}); ok {
			quizCount = len(qArr)
		}

		db.Exec("DELETE FROM lessons WHERE id=?", lid)
		db.Exec("DELETE FROM flashcard_data WHERE lesson_id=?", lid)
		db.Exec("DELETE FROM flashcard_reviews WHERE lesson_id=?", lid)

		db.Exec("INSERT INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			lid, title, notes, mnemonic, status, ltype, courseID, progress, quizCount, quizData, codingData)

		if fcRaw, ok := lesson["flashcards"]; ok {
			fcList, ok := fcRaw.([]interface{})
			if ok {
				for _, fcr := range fcList {
					fc, _ := fcr.(map[string]interface{})
					if fc == nil {
						continue
					}
					front, _ := fc["front"].(string)
					back, _ := fc["back"].(string)
					if front == "" || back == "" {
						continue
					}
					db.Exec("INSERT INTO flashcard_data (lesson_id,front,back) VALUES (?,?,?)", lid, front, back)
					db.Exec("INSERT INTO flashcard_reviews (lesson_id,front,back,ease,interval_days,repetitions,next_review) VALUES (?,?,?,2.5,1,0,?)",
						lid, front, back, time.Now().Format(time.RFC3339))
				}
			}
		}

		jsonResp(w, map[string]interface{}{
			"status": "ok", "lesson_imported": lid, "course_id": courseID,
		})
		return
	}

	jsonError(w, "provide 'courses' (array) or 'lesson' (object) in JSON", 400)
}

func buildAdminTree(courses []map[string]interface{}) []map[string]interface{} {
	children := map[string][]map[string]interface{}{}
	for _, c := range courses {
		pid, _ := c["parent_id"].(string)
		children[pid] = append(children[pid], c)
	}

	var build func(parentID string) []map[string]interface{}
	build = func(parentID string) []map[string]interface{} {
		var nodes []map[string]interface{}
		for _, c := range children[parentID] {
			id, _ := c["id"].(string)
			name, _ := c["name"].(string)
			node := map[string]interface{}{"id": id, "name": name, "type": "course"}

			// Get lessons for this course
			lessonRows, _ := db.Query("SELECT id, title, status, type FROM lessons WHERE course_id=? ORDER BY id", id)
			if lessonRows != nil {
				var lessonNodes []map[string]interface{}
				for lessonRows.Next() {
					var lid, ltitle, lstatus, ltype string
					lessonRows.Scan(&lid, &ltitle, &lstatus, &ltype)
					lessonNodes = append(lessonNodes, map[string]interface{}{
						"id": lid, "name": ltitle, "type": ltype, "status": lstatus,
					})
				}
				lessonRows.Close()
				if len(lessonNodes) > 0 {
					node["children"] = lessonNodes
				}
			}

			subCourses := build(id)
			if len(subCourses) > 0 {
				if existing, ok := node["children"]; ok {
					existingSlice := existing.([]map[string]interface{})
					node["children"] = append(existingSlice, subCourses...)
				} else {
					node["children"] = subCourses
				}
			}

			nodes = append(nodes, node)
		}
		return nodes
	}

	return build("")
}

// ======== PROGRESS ========
func handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		for k, v := range data {
			setProgress(k, fmt.Sprintf("%v", v))
		}
	}
	xp, _ := strconv.Atoi(getProgress("xp"))
	xpMax, _ := strconv.Atoi(getProgress("xp_max"))
	streak, _ := strconv.Atoi(getProgress("streak"))
	mastery, _ := strconv.Atoi(getProgress("mastery"))
	solved, _ := strconv.Atoi(getProgress("solved"))

	jsonResp(w, map[string]interface{}{
		"problems": solved, "streak": streak, "mastery": mastery,
		"solved": solved, "xp": xp, "xp_max": xpMax,
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	streak, _ := strconv.Atoi(getProgress("streak"))
	totalHours, _ := strconv.ParseFloat(getProgress("total_hours"), 64)
	jsonResp(w, map[string]interface{}{
		"topics": []map[string]interface{}{
			{"name": "data structures", "progress": 0},
			{"name": "graphs", "progress": 40},
			{"name": "dynamic programming", "progress": 65},
			{"name": "math", "progress": 90},
			{"name": "sortings", "progress": 25},
			{"name": "bfs", "progress": 50},
		},
		"xp_history":  []int{30, 45, 20, 60, 35, 50, 40},
		"streak":      streak,
		"total_hours": totalHours,
	})
}

// ======== SM-2 ========
type SM2Result struct {
	Ease        float64 `json:"ease"`
	Interval    int     `json:"interval"`
	Repetitions int     `json:"repetitions"`
	NextReview  string  `json:"next_review"`
}

func sm2Calculate(ease float64, interval int, reps int, quality int) SM2Result {
	if quality < 3 {
		reps = 0
		interval = 1
	} else {
		if reps == 0 {
			interval = 1
		} else if reps == 1 {
			interval = 3
		} else {
			interval = int(math.Round(float64(interval) * ease))
		}
		reps++
	}
	ease = ease + (0.1 - (5-float64(quality))*(0.08+(5-float64(quality))*0.02))
	if ease < 1.3 {
		ease = 1.3
	}
	nextReview := time.Now().AddDate(0, 0, interval).Format(time.RFC3339)
	return SM2Result{Ease: ease, Interval: interval, Repetitions: reps, NextReview: nextReview}
}

func handleFlashcards(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/flashcards/")

	if strings.HasSuffix(path, "/review") && r.Method == "POST" {
		lessonID := strings.TrimSuffix(path, "/review")
		body, _ := io.ReadAll(r.Body)
		var req struct {
			CardIndex int `json:"card_index"`
			Quality   int `json:"quality"`
		}
		json.Unmarshal(body, &req)

		var id int
		var ease float64
		var interval, reps int
		err := db.QueryRow("SELECT id, ease, interval_days, repetitions FROM flashcard_reviews WHERE lesson_id=? LIMIT 1 OFFSET ?",
			lessonID, req.CardIndex).Scan(&id, &ease, &interval, &reps)
		if err != nil {
			http.Error(w, "card not found", 404)
			return
		}

		result := sm2Calculate(ease, interval, reps, req.Quality)
		db.Exec("UPDATE flashcard_reviews SET ease=?, interval_days=?, repetitions=?, next_review=?, last_reviewed=? WHERE id=?",
			result.Ease, result.Interval, result.Repetitions, result.NextReview, time.Now().Format(time.RFC3339), id)

		xp, _ := strconv.Atoi(getProgress("xp"))
		setProgress("xp", strconv.Itoa(xp+req.Quality*2))

		jsonResp(w, result)
		return
	}

	lessonID := strings.TrimSuffix(path, "/schedule")
	rows, err := db.Query("SELECT front, back, ease, interval_days, next_review FROM flashcard_reviews WHERE lesson_id=? ORDER BY id", lessonID)
	if err != nil {
		jsonResp(w, []interface{}{})
		return
	}
	defer rows.Close()

	var cards []map[string]interface{}
	for rows.Next() {
		var front, back, nextReview string
		var ease float64
		var interval int
		rows.Scan(&front, &back, &ease, &interval, &nextReview)
		cards = append(cards, map[string]interface{}{
			"front": front, "back": back, "ease": ease,
			"interval": interval, "next_review": nextReview,
		})
	}
	jsonResp(w, cards)
}

func handlePowerScore(w http.ResponseWriter, r *http.Request) {
	xp, _ := strconv.Atoi(getProgress("xp"))
	xpMax, _ := strconv.Atoi(getProgress("xp_max"))
	streak, _ := strconv.Atoi(getProgress("streak"))
	mastery, _ := strconv.Atoi(getProgress("mastery"))

	var streakMult float64
	if streak >= 30 {
		streakMult = 2.0
	} else if streak >= 14 {
		streakMult = 1.5
	} else if streak >= 7 {
		streakMult = 1.25
	} else {
		streakMult = 1.0
	}

	masteryPct := float64(mastery) / 100.0
	xpDaily := math.Min(float64(xp), float64(xpMax))
	powerScore := int(math.Round(xpDaily * masteryPct * streakMult))

	jsonResp(w, map[string]interface{}{
		"power_score":      powerScore,
		"xp":               xp,
		"xp_max":           xpMax,
		"mastery":          mastery,
		"streak":           streak,
		"streak_multiplier": streakMult,
		"formula":          fmt.Sprintf("%d × %.2f%% × %.2f = %d", xp, masteryPct*100, streakMult, powerScore),
	})
}

// ======== SYNC ========
func handleSyncExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	lessons := []map[string]interface{}{}
	rows, _ := db.Query("SELECT id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data FROM lessons")
	for rows.Next() {
		var id, title, notes, mnemonic, status, ltype, courseID, quizData, codingData string
		var progress, quizCount int
		rows.Scan(&id, &title, &notes, &mnemonic, &status, &ltype, &courseID, &progress, &quizCount, &quizData, &codingData)
		lessons = append(lessons, map[string]interface{}{
			"id": id, "title": title, "notes": notes, "mnemonic": mnemonic,
			"status": status, "type": ltype, "course_id": courseID,
			"progress": progress, "quiz_count": quizCount,
			"quiz_data": quizData, "coding_data": codingData,
		})
	}
	rows.Close()

	fcRows, _ := db.Query("SELECT lesson_id,front,back,ease,interval_days,repetitions,next_review,last_reviewed FROM flashcard_reviews")
	var fcCards []map[string]interface{}
	for fcRows.Next() {
		var lessonID, front, back, nextReview, lastReviewed string
		var ease float64
		var interval, reps int
		fcRows.Scan(&lessonID, &front, &back, &ease, &interval, &reps, &nextReview, &lastReviewed)
		fcCards = append(fcCards, map[string]interface{}{
			"lesson_id": lessonID, "front": front, "back": back,
			"ease": ease, "interval": interval, "repetitions": reps,
			"next_review": nextReview, "last_reviewed": lastReviewed,
		})
	}
	fcRows.Close()

	progRows, _ := db.Query("SELECT key, value FROM progress")
	progress := map[string]string{}
	for progRows.Next() {
		var k, v string
		progRows.Scan(&k, &v)
		progress[k] = v
	}
	progRows.Close()

	courseRows, _ := db.Query("SELECT id, name, parent_id, sort_order FROM courses")
	var courses []map[string]interface{}
	for courseRows.Next() {
		var id, name, parentID string
		var sort int
		courseRows.Scan(&id, &name, &parentID, &sort)
		courses = append(courses, map[string]interface{}{"id": id, "name": name, "parent_id": parentID, "sort_order": sort})
	}
	courseRows.Close()

	export := map[string]interface{}{
		"version": 2, "exported_at": time.Now().Format(time.RFC3339),
		"lessons": lessons, "flashcards": fcCards,
		"progress": progress, "courses": courses,
	}

	data, _ := json.Marshal(export)
	key := os.Getenv("SYNC_KEY")
	if key == "" {
		key = "brutforse-default-key"
	}
	encrypted := encrypt(data, key)
	code := generateCode()
	os.MkdirAll("./data/sync", 0755)
	os.WriteFile("./data/sync/"+code+".enc", []byte(encrypted), 0644)

	jsonResp(w, map[string]interface{}{
		"code": code, "data_size": len(data), "exported_at": export["exported_at"],
	})
}

func handleSyncImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req struct {
		Code string `json:"code"`
	}
	json.Unmarshal(body, &req)

	encData, err := os.ReadFile("./data/sync/" + req.Code + ".enc")
	if err != nil {
		http.Error(w, "sync code not found or expired", 404)
		return
	}

	key := os.Getenv("SYNC_KEY")
	if key == "" {
		key = "brutforse-default-key"
	}
	decrypted := decrypt(string(encData), key)

	var export map[string]interface{}
	json.Unmarshal([]byte(decrypted), &export)

	if courses, ok := export["courses"].([]interface{}); ok {
		for _, c := range courses {
			cm := c.(map[string]interface{})
			db.Exec("INSERT OR REPLACE INTO courses (id,name,parent_id,sort_order) VALUES (?,?,?,?)",
				cm["id"], cm["name"], cm["parent_id"], cm["sort_order"])
		}
	}

	if lessons, ok := export["lessons"].([]interface{}); ok {
		for _, l := range lessons {
			lm := l.(map[string]interface{})
			db.Exec("INSERT OR REPLACE INTO lessons (id,title,notes,mnemonic,status,type,course_id,progress,quiz_count,quiz_data,coding_data) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
				lm["id"], lm["title"], lm["notes"], lm["mnemonic"],
				lm["status"], lm["type"], lm["course_id"],
				lm["progress"], lm["quiz_count"],
				lm["quiz_data"], lm["coding_data"])
		}
	}

	if cards, ok := export["flashcards"].([]interface{}); ok {
		for _, c := range cards {
			cm := c.(map[string]interface{})
			db.Exec("INSERT INTO flashcard_reviews (lesson_id,front,back,ease,interval_days,repetitions,next_review,last_reviewed) VALUES (?,?,?,?,?,?,?,?)",
				cm["lesson_id"], cm["front"], cm["back"],
				cm["ease"], cm["interval"], cm["repetitions"],
				cm["next_review"], cm["last_reviewed"])
		}
	}

	if progress, ok := export["progress"].(map[string]interface{}); ok {
		for k, v := range progress {
			setProgress(k, fmt.Sprintf("%v", v))
		}
	}

	jsonResp(w, map[string]interface{}{"status": "imported"})
}

func encrypt(data []byte, key string) string {
	block, _ := aes.NewCipher(padKey(key))
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decrypt(encrypted string, key string) string {
	data, _ := base64.StdEncoding.DecodeString(encrypted)
	block, _ := aes.NewCipher(padKey(key))
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
	return string(plaintext)
}

func padKey(key string) []byte {
	k := []byte(key)
	if len(k) < 32 {
		padded := make([]byte, 32)
		copy(padded, k)
		return padded
	}
	return k[:32]
}

func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func handleRunCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		jsonResp(w, map[string]string{"output": "Error parsing request", "error": err.Error()})
		return
	}
	if req.Language != "javascript" && req.Language != "js" {
		jsonResp(w, map[string]string{"output": fmt.Sprintf("Language %s not supported", req.Language), "error": ""})
		return
	}
	jsonResp(w, map[string]string{"output": "Code received (sandbox execution on frontend)", "error": ""})
}
