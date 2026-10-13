"""Valid synthetic stream bodies for Chat Completions, Responses, and Anthropic."""
import json

ANSWER = "Synthetic answer."
TITLE = "Synthetic title"
TOOL_ARGS = json.dumps({"question": "Choose synthetic option?"})


def sse(data, name=None):
    head = "event: " + name + "\n" if name else ""
    return head + "data: " + json.dumps(data, separators=(",", ":")) + "\n\n"


def kind_of(path):
    for suffix, kind in (("/responses", "responses"), ("/chat/completions", "chat"), ("/messages", "anthropic")):
        if path.endswith(suffix):
            return kind
    return None


def chat(text, tool=False):
    delta = {"role": "assistant", "content": text}
    if tool:
        call = {"index": 0, "id": "fixture-tool", "type": "function", "function": {"name": "ask_user", "arguments": TOOL_ARGS}}
        delta = {"role": "assistant", "tool_calls": [call]}
    usage = {"choices": [], "usage": {"prompt_tokens": 7, "completion_tokens": 3}}
    return sse({"choices": [{"delta": delta}]}) + sse(usage) + "data: [DONE]\n\n"


def responses(text, tool=False):
    item = {"type": "message", "role": "assistant", "content": [{"type": "output_text", "text": text}]}
    if tool:
        item = {"type": "function_call", "call_id": "fixture-tool", "name": "ask_user", "arguments": TOOL_ARGS}
    done = {"type": "response.completed", "response": {"output": [item], "usage": {"input_tokens": 7, "output_tokens": 3}}}
    delta = [] if tool else [sse({"type": "response.output_text.delta", "delta": text})]
    return "".join(delta) + sse(done)


def anthropic(text, tool=False):
    start = {"type": "message_start", "message": {"role": "assistant", "usage": {"input_tokens": 7}}}
    block = {"type": "text", "text": ""}
    delta = {"type": "text_delta", "text": text}
    if tool:
        block = {"type": "tool_use", "id": "fixture-tool", "name": "ask_user", "input": {}}
        delta = {"type": "input_json_delta", "partial_json": TOOL_ARGS}
    events = [
        ("message_start", start),
        ("content_block_start", {"type": "content_block_start", "index": 0, "content_block": block}),
        ("content_block_delta", {"type": "content_block_delta", "index": 0, "delta": delta}),
        ("content_block_stop", {"type": "content_block_stop", "index": 0}),
        ("message_delta", {"type": "message_delta", "delta": {"stop_reason": "tool_use" if tool else "end_turn"}, "usage": {"output_tokens": 3}}),
        ("message_stop", {"type": "message_stop"}),
    ]
    return "".join(sse(data, name) for name, data in events)


def partial(kind):
    if kind == "responses":
        return sse({"type": "response.output_text.delta", "delta": "partial synthetic answer"})
    if kind == "anthropic":
        return anthropic("partial synthetic answer").split("event: content_block_stop")[0]
    return sse({"choices": [{"delta": {"role": "assistant", "content": "partial synthetic answer"}}]})


BUILDERS = {"chat": chat, "responses": responses, "anthropic": anthropic}
