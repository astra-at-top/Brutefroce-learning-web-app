#!/usr/bin/env python3
import http.server
import os
import time

PORT = int(os.environ.get('PORT', 8000))
DIR = os.path.dirname(os.path.abspath(__file__))
WATCHED = {}

class ReloadHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIR, **kwargs)

    def do_GET(self):
        if self.path == '/reload':
            return self._handle_sse()
        path = self.translate_path(self.path)
        if os.path.isfile(path) and path.endswith('.html'):
            WATCHED[path] = os.path.getmtime(path)
            with open(path, 'rb') as f:
                content = f.read()
            inject = b'<script>new EventSource("/reload").onmessage=()=>location.reload()</script>'
            content = content.replace(b'</body>', inject + b'</body>')
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.send_header('Content-Length', str(len(content)))
            self.end_headers()
            self.wfile.write(content)
        else:
            super().do_GET()

    def _handle_sse(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/event-stream')
        self.send_header('Cache-Control', 'no-cache')
        self.send_header('Connection', 'keep-alive')
        self.end_headers()
        while True:
            for path, last_mtime in list(WATCHED.items()):
                try:
                    if os.path.getmtime(path) != last_mtime:
                        self.wfile.write(b'data: reload\n\n')
                        self.wfile.flush()
                        WATCHED[path] = os.path.getmtime(path)
                except OSError:
                    pass
            try:
                time.sleep(0.3)
            except (BrokenPipeError, ConnectionResetError):
                break

if __name__ == '__main__':
    server = http.server.ThreadingHTTPServer(('', PORT), ReloadHandler)
    print(f'Dev server: http://localhost:{PORT}  (reloads on HTML changes)')
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        server.shutdown()
