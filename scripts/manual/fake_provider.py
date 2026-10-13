#!/usr/bin/env python3
"""CPU-only local SSE provider fixture for manual integration checks."""
import json
import os
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import streams

MODE = os.environ.get("FIXTURE_MODE", "ok")
LOG = os.environ.get("FIXTURE_LOG", "fixture.jsonl")
DELAY = float(os.environ.get("FIXTURE_DELAY", "8"))
MARKER = os.environ.get("FIXTURE_TITLE_MARKER", "write a title of 2 to 6 words").lower()
LOCK = threading.Lock()
COUNT = {"main": 0, "title": 0}
REJECT = {"auth": (401, "synthetic auth failure"), "quota": (429, "synthetic quota exhausted"),
          "bare400": (400, "synthetic request rejected"), "bare404": (404, "synthetic request rejected")}
UNSUPPORTED = {"unsupported": "responses", "unsupported-chat": "chat"}


def is_title(body):
    head = [body.get("system"), body.get("instructions")]
    head += (body.get("messages") or body.get("input") or [])[:1]
    return MARKER in json.dumps(head).lower()


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def record(self, phase, path, role, seq, body=None):
        line = {"phase": phase, "path": path, "kind": role, "seq": seq, "time": time.time(), "body": body}
        with LOCK, open(LOG, "a", encoding="utf-8") as out:
            out.write(json.dumps(line) + "\n")

    def do_GET(self):
        if self.path == "/health":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"ok")
        elif self.path.endswith("/models"):
            self.reply(200, {"data": [{"id": "synthetic-model"}]})
        else:
            self.send_error(404)

    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0))) or b"{}")
        role = "title" if is_title(body) else "main"
        with LOCK:
            COUNT[role] += 1
            seq = COUNT[role]
        self.record("request", self.path, role, seq, body)
        kind = streams.kind_of(self.path)
        auth = self.headers.get("Authorization", "") + self.headers.get("x-api-key", "")
        if kind is None:
            return self.reply(404, {"error": {"message": "unknown synthetic endpoint"}})
        if auth and "synthetic" not in auth:
            return self.reply(401, {"error": {"message": "synthetic auth failure"}})
        if MODE in REJECT:
            status, message = REJECT[MODE]
            return self.reply(status, {"error": {"message": message}})
        if UNSUPPORTED.get(MODE) == kind:
            message = "The /" + self.path.rsplit("/", 1)[-1] + " endpoint is not supported"
            return self.reply(400, {"error": {"code": "unsupported_endpoint", "message": message}})
        if MODE == "timeout":
            time.sleep(DELAY + 10)
        if MODE == "delay-first" and role == "main" and seq == 1:
            time.sleep(DELAY)
        if MODE == "title-failure" and role == "title":
            return self.reply(500, {"error": {"message": "synthetic title failure"}})
        if MODE == "partial":
            self.sse(streams.partial(kind))
            self.close_connection = True
            return self.record("respond", self.path, role, seq)
        tool = MODE == "tool-loop" and role == "main" and seq == 1
        self.sse(streams.BUILDERS[kind](streams.TITLE if role == "title" else streams.ANSWER, tool))
        self.record("respond", self.path, role, seq)

    def reply(self, status, data):
        raw = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def sse(self, text):
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        self.wfile.write(text.encode())
        self.wfile.flush()


def make_server(host="127.0.0.1", port=0):
    return ThreadingHTTPServer((host, port), Handler)


if __name__ == "__main__":
    server = make_server(os.environ.get("FIXTURE_HOST", "127.0.0.1"), int(os.environ.get("FIXTURE_PORT", "8765")))
    print("fixture listening on %s:%d mode=%s" % (server.server_address + (MODE,)), flush=True)
    server.serve_forever()
