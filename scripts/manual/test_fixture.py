#!/usr/bin/env python3
"""Offline checks of the fixture server. Uses an in-process server on a free port."""
import http.client
import json
import os
import tempfile
import threading
import time
import unittest

os.environ["FIXTURE_LOG"] = os.path.join(tempfile.mkdtemp(prefix="fixture-test-"), "log.jsonl")
os.environ["FIXTURE_DELAY"] = "1"
import fake_provider as fp  # noqa: E402
import streams  # noqa: E402

PATHS = {"responses": "/v1/responses", "chat": "/v1/chat/completions", "anthropic": "/v1/messages"}
KEY = {"Authorization": "Bearer synthetic-key", "x-api-key": "synthetic-key", "Content-Type": "application/json"}


def events(raw):
    out = []
    for block in raw.split("\n\n"):
        data = [l[5:].strip() for l in block.split("\n") if l.startswith("data:")]
        if data:
            out.append(data[0] if data[0] == "[DONE]" else json.loads(data[0]))
    return out


def decode(kind, raw):
    """Return (text, tool_args, completed) like the product parsers."""
    ev, text, args, done = events(raw), "", "", False
    for e in ev:
        if kind == "chat":
            done = done or e == "[DONE]"
            for c in (e.get("choices", []) if e != "[DONE]" else []):
                text += c["delta"].get("content", "")
                for t in c["delta"].get("tool_calls", []):
                    args += t["function"]["arguments"]
        elif kind == "responses":
            if e["type"] == "response.output_text.delta":
                text += e["delta"]
            if e["type"] == "response.completed":
                done = True
                for it in e["response"]["output"]:
                    args += it.get("arguments", "")
                    text = text or "".join(p["text"] for p in it.get("content", []))
        else:
            done = done or e["type"] == "message_stop"
            if e["type"] == "content_block_delta":
                text += e["delta"].get("text", "")
                args += e["delta"].get("partial_json", "")
    return text, args, done


TITLE_PROMPT = "Write a title of 2 to 6 words."


class Fixture(unittest.TestCase):
    def setUp(self):
        open(os.environ["FIXTURE_LOG"], "w").close()
        fp.COUNT.update(main=0, title=0)
        fp.MODE = "ok"
        self.server = fp.make_server()
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.port = self.server.server_address[1]

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()

    def post(self, kind, prompt="synthetic question", headers=KEY, tools=False):
        c = http.client.HTTPConnection("127.0.0.1", self.port, timeout=10)
        turns = [{"role": "user", "content": prompt}]
        body = {"model": "synthetic-model", "stream": True, "messages": turns}
        if kind == "responses":
            body = {"model": "synthetic-model", "stream": True, "input": turns}
        c.request("POST", PATHS[kind], json.dumps(body), headers)
        r = c.getresponse()
        return r, r.read().decode()

    def log(self):
        with open(os.environ["FIXTURE_LOG"]) as f:
            return [json.loads(l) for l in f if l.strip()]

    def test_valid_streams(self):
        for kind in PATHS:
            r, raw = self.post(kind)
            self.assertEqual(r.status, 200, kind)
            self.assertEqual(decode(kind, raw), ("Synthetic answer.", "", True), kind)

    def test_tool_streams(self):
        fp.MODE = "tool-loop"
        for kind in PATHS:
            fp.COUNT.update(main=0, title=0)
            _, raw = self.post(kind)
            _, args, done = decode(kind, raw)
            self.assertEqual(json.loads(args), json.loads(streams.TOOL_ARGS), kind)
            self.assertTrue(done, kind)
            _, raw = self.post(kind)
            self.assertEqual(decode(kind, raw)[0], "Synthetic answer.", kind)

    def test_protocol_rejection(self):
        fp.MODE = "unsupported"
        r, raw = self.post("responses")
        self.assertEqual(r.status, 400)
        self.assertIn("/responses endpoint is not supported", raw)
        self.assertIn("unsupported_endpoint", raw)
        self.assertEqual(self.post("chat")[0].status, 200)
        fp.MODE = "unsupported-chat"
        self.assertEqual(self.post("chat")[0].status, 400)
        self.assertEqual(self.post("responses")[0].status, 200)

    def test_bare_errors_have_no_protocol_evidence(self):
        for mode, status in (("bare400", 400), ("bare404", 404), ("auth", 401), ("quota", 429)):
            fp.MODE = mode
            r, raw = self.post("responses")
            self.assertEqual(r.status, status)
            self.assertNotIn("endpoint", raw.lower())
            self.assertNotIn("support", raw.lower())

    def test_title_dispatch(self):
        self.post("chat")
        self.post("chat", TITLE_PROMPT)
        roles = [(e["kind"], e["seq"]) for e in self.log() if e["phase"] == "request"]
        self.assertEqual(roles, [("main", 1), ("title", 1)])

    def test_delay_first_title_before_answer(self):
        fp.MODE = "delay-first"
        done = {}

        def main():
            self.post("chat")
            done["main"] = time.time()

        t = threading.Thread(target=main)
        t.start()
        time.sleep(0.3)
        r, raw = self.post("chat", TITLE_PROMPT)
        done["title"] = time.time()
        t.join()
        self.assertEqual(decode("chat", raw)[0], "Synthetic title")
        self.assertLess(done["title"], done["main"])
        start = time.time()
        self.post("chat")
        self.assertLess(time.time() - start, 0.9)

    def test_title_failure_only_title(self):
        fp.MODE = "title-failure"
        self.assertEqual(self.post("chat", TITLE_PROMPT)[0].status, 500)
        self.assertEqual(self.post("chat")[0].status, 200)

    def test_partial_has_no_terminator(self):
        fp.MODE = "partial"
        for kind in PATHS:
            _, raw = self.post(kind)
            self.assertFalse(decode(kind, raw)[2], kind)

    def test_foreign_key_rejected(self):
        r, _ = self.post("chat", headers={"Authorization": "Bearer real-key", "Content-Type": "application/json"})
        self.assertEqual(r.status, 401)

    def test_binds_loopback_only(self):
        self.assertEqual(self.server.server_address[0], "127.0.0.1")


if __name__ == "__main__":
    unittest.main()
