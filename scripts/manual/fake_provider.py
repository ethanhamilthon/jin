#!/usr/bin/env python3
"""CPU-only local SSE provider fixture for manual integration checks."""
import json
import os
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

MODE = os.environ.get("FIXTURE_MODE", "ok")
LOG = os.environ.get("FIXTURE_LOG", "fixture.jsonl")
DELAY = float(os.environ.get("FIXTURE_DELAY", "8"))


def event(data):
    return "data: " + json.dumps(data, separators=(",", ":")) + "\n\n"


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def record(self, path, body):
        with open(LOG, "a", encoding="utf-8") as out:
            out.write(json.dumps({"path": path, "body": body, "time": time.time()}) + "\n")

    def do_GET(self):
        if self.path == "/health":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"ok")
        else:
            self.send_error(404)

    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0))) or b"{}")
        self.record(self.path, body)
        auth = self.headers.get("Authorization", "") + self.headers.get("x-api-key", "")
        if auth and "synthetic" not in auth:
            return self.reply(401, {"error": {"message": "synthetic auth failure"}})
        if MODE == "auth":
            return self.reply(401, {"error": {"message": "synthetic auth failure"}})
        if MODE == "quota":
            return self.reply(429, {"error": {"message": "synthetic quota exhausted"}})
        if MODE in ("bare400", "bare404"):
            return self.reply(int(MODE[-3:]), {"error": {"message": "synthetic request rejected"}})
        if MODE == "unsupported" and self.path.endswith("/responses"):
            return self.reply(400, {"error": {"code": "unsupported_endpoint", "message": "The /responses endpoint is not supported"}})
        if MODE == "timeout":
            time.sleep(DELAY + 10)
        if MODE == "delay-first" and len(open(LOG, encoding="utf-8").readlines()) == 1:
            time.sleep(DELAY)
        if MODE == "title-failure" and ("title" in json.dumps(body).lower() or "name this session" in json.dumps(body).lower()):
            return self.reply(500, {"error": {"message": "synthetic title failure"}})
        if MODE == "partial":
            return self.partial(self.path)
        if MODE == "tool-loop" and len(open(LOG, encoding="utf-8").readlines()) == 1:
            return self.tool_call(self.path)
        self.stream(self.path)

    def reply(self, status, data):
        raw = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def partial(self, path):
        payload = 'data: {"choices":[{"delta":{"content":"partial synthetic answer"}}]}\n\n'
        if path.endswith("/responses"):
            payload = event({"type": "response.output_text.delta", "delta": "partial synthetic answer"})
        if path.endswith("/messages"):
            payload = event({"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "partial synthetic answer"}})
        self.sse(payload)
        self.close_connection = True

    def tool_call(self, path):
        if path.endswith("/responses"):
            text = event({"type": "response.output_item.added", "item": {"type": "function_call", "call_id": "fixture-tool", "name": "ask_user", "arguments": "{\\\"question\\\":\\\"Choose synthetic option?\\\"}"}}) + event({"type": "response.completed", "response": {"output": [{"type": "function_call", "call_id": "fixture-tool", "name": "ask_user", "arguments": "{\\\"question\\\":\\\"Choose synthetic option?\\\"}"}]}})
        elif path.endswith("/messages"):
            text = event({"type": "content_block_start", "index": 0, "content_block": {"type": "tool_use", "id": "fixture-tool", "name": "ask_user", "input": {"question": "Choose synthetic option?"}}}) + event({"type": "message_stop"})
        else:
            text = 'data: {"choices":[{"delta":{"tool_calls":[{"id":"fixture-tool","function":{"name":"ask_user","arguments":"{\\\"question\\\":\\\"Choose synthetic option?\\\"}"}}]}}]}\\n\\n' + "data: [DONE]\\n\\n"
        self.sse(text)

    def stream(self, path):
        if path.endswith("/responses"):
            chunks = [event({"type": "response.output_text.delta", "delta": "Synthetic answer."}), event({"type": "response.completed", "response": {"output": [], "usage": {"input_tokens": 7, "output_tokens": 3}}})]
        elif path.endswith("/messages"):
            chunks = [event({"type": "message_start", "message": {"usage": {"input_tokens": 7}}}), event({"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Synthetic answer."}}), event({"type": "message_stop"})]
        else:
            chunks = ['data: {"choices":[{"delta":{"content":"Synthetic answer."}}]}\n\n', "data: [DONE]\n\n"]
        self.sse("".join(chunks))

    def sse(self, text):
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        self.wfile.write(text.encode())
        self.wfile.flush()


if __name__ == "__main__":
    host = os.environ.get("FIXTURE_HOST", "127.0.0.1")
    port = int(os.environ.get("FIXTURE_PORT", "8765"))
    ThreadingHTTPServer((host, port), Handler).serve_forever()
