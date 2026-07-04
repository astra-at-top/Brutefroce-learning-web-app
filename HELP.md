# BruteForce Learning — Help

## Run on a different port

### Go server (production)
```sh
PORT=9090 go run main.go
# or
export PORT=9090
./brutforselearning
```
Default port: `8080`

### Python dev server (hot-reload)
```sh
PORT=9090 python3 server.py
```
Default port: `8000`

## AI Code Review

The Code Playground tab has an **AI Review** button that uses opencode to review student code.

**Requirements:**
- [opencode](https://opencode.ai) must be installed and in PATH
- The Go server automatically starts `opencode serve` on port 4099 as a sidecar
- Uses `opencode/north-mini-code-free` model (free)

**How it works:**
1. Write code in the editor
2. Click **AI Review**
3. opencode reviews your code against the expected solution
4. Feedback appears below the editor
