#!/usr/bin/env python3
"""Corte duro por límite de tool calls o reintentos del mismo enfoque."""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path
from typing import Any

MAX_TOOL_CALLS = 20
MAX_SAME_APPROACH_ATTEMPTS = 4  # 1 intento inicial + 3 reintentos

STATE_DIR = Path(__file__).resolve().parent / "state"
HOOKS_DIR = Path(__file__).resolve().parent


def session_key(data: dict[str, Any]) -> str:
    return data.get("conversation_id") or data.get("session_id") or "unknown"


def generation_key(data: dict[str, Any]) -> str:
    return data.get("generation_id") or session_key(data)


def state_path(session_id: str) -> Path:
    safe = "".join(c if c.isalnum() or c in "-_" else "_" for c in session_id)
    return STATE_DIR / f"{safe}.json"


def summarize_tool_call(data: dict[str, Any]) -> str:
    tool_name = data.get("tool_name", "unknown")
    tool_input = data.get("tool_input") or {}

    if tool_name == "Shell":
        command = tool_input.get("command", "")
        cwd = tool_input.get("working_directory") or data.get("cwd") or ""
        return f"{tool_name}: {command}" + (f" (cwd: {cwd})" if cwd else "")

    if tool_name in {"Read", "Write", "StrReplace", "Delete"}:
        return f"{tool_name}: {tool_input.get('path', '')}"

    if tool_name == "Grep":
        pattern = tool_input.get("pattern", "")
        path = tool_input.get("path", "")
        return f"Grep: {pattern!r}" + (f" en {path}" if path else "")

    if tool_name == "Task":
        description = tool_input.get("description") or tool_input.get("prompt", "")[:120]
        return f"Task: {description}"

    if tool_name.startswith("MCP"):
        return f"{tool_name}: {json.dumps(tool_input, ensure_ascii=False)[:160]}"

    return f"{tool_name}: {json.dumps(tool_input, ensure_ascii=False, sort_keys=True)[:160]}"


def fingerprint(data: dict[str, Any]) -> str:
    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input") or {}
    parts = [tool_name]

    if tool_name == "Shell":
        parts.extend(
            [
                tool_input.get("command", ""),
                tool_input.get("working_directory", "") or data.get("cwd", ""),
            ]
        )
    elif tool_name in {"Read", "Write", "StrReplace", "Delete"}:
        parts.append(tool_input.get("path", ""))
        if tool_name == "StrReplace":
            parts.append(tool_input.get("old_string", "")[:200])
    elif tool_name == "Grep":
        parts.extend([tool_input.get("pattern", ""), tool_input.get("path", "")])
    elif tool_name == "Task":
        parts.extend(
            [
                tool_input.get("description", ""),
                tool_input.get("prompt", "")[:300],
                tool_input.get("subagent_type", ""),
            ]
        )
    else:
        parts.append(json.dumps(tool_input, ensure_ascii=False, sort_keys=True))

    raw = "|".join(str(part) for part in parts)
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]


def load_state(session_id: str) -> dict[str, Any]:
    path = state_path(session_id)
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return {}


def save_state(session_id: str, state: dict[str, Any]) -> None:
    STATE_DIR.mkdir(parents=True, exist_ok=True)
    state_path(session_id).write_text(
        json.dumps(state, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )


def reset_for_generation(state: dict[str, Any], generation_id: str) -> dict[str, Any]:
    if state.get("generation_id") == generation_id:
        return state
    return {
        "generation_id": generation_id,
        "tool_calls": 0,
        "fingerprints": {},
        "history": [],
        "cutoff": False,
        "cutoff_reason": None,
    }


def build_history_lines(history: list[str], limit: int = 12) -> str:
    if not history:
        return "- (sin acciones registradas)"
    shown = history[-limit:]
    lines = [f"- {item}" for item in shown]
    if len(history) > limit:
        lines.insert(0, f"- ... ({len(history) - limit} acciones anteriores omitidas)")
    return "\n".join(lines)


def register_cutoff_error(
    data: dict[str, Any],
    reason: str,
    history: list[str],
) -> None:
    try:
        import sys

        if str(HOOKS_DIR) not in sys.path:
            sys.path.insert(0, str(HOOKS_DIR))
        from cursor_errores_lib import build_record, infer_proyecto, monorepo_root, write_record

        root = monorepo_root(HOOKS_DIR)
        cwd = data.get("cwd") or (data.get("tool_input") or {}).get("working_directory")
        record = build_record(
            origen="hook",
            tipo="limite_tokens",
            resumen=reason,
            detalle=summarize_tool_call(data),
            error=reason,
            proyecto=infer_proyecto(root, cwd),
            session_id=session_key(data),
            generation_id=generation_key(data),
            tool_name=data.get("tool_name", ""),
            acciones=history[-12:],
            workspace_root=root,
            extra={"historial": build_history_lines(history, limit=12)},
        )
        write_record(record, root)
    except Exception:
        pass


def deny_response(reason: str, history: list[str]) -> dict[str, str]:
    history_text = build_history_lines(history)
    user_message = (
        f"Tarea cortada automáticamente: {reason}\n\n"
        f"Acciones registradas en esta tarea:\n{history_text}\n\n"
        "El agente debe responder solo con texto explicando qué pasó y qué se necesita para continuar."
    )
    agent_message = (
        f"CORTE DURO ACTIVADO: {reason}\n\n"
        "No uses más herramientas. Responde al usuario en español con:\n"
        "1) Qué pasó (límite de 20 tool calls o 3 reintentos del mismo enfoque).\n"
        "2) Qué se intentó.\n"
        "3) Dónde se trabó.\n"
        "4) Qué falta del usuario.\n"
        "5) Próximos pasos concretos.\n\n"
        f"Historial reciente:\n{history_text}"
    )
    return {
        "permission": "deny",
        "user_message": user_message,
        "agent_message": agent_message,
    }


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        print(json.dumps({"permission": "allow"}))
        return 0

    session_id = session_key(data)
    generation_id = generation_key(data)
    state = reset_for_generation(load_state(session_id), generation_id)
    history: list[str] = list(state.get("history", []))
    fingerprints: dict[str, int] = dict(state.get("fingerprints", {}))

    if state.get("cutoff"):
        print(json.dumps(deny_response(state.get("cutoff_reason", "límite previo"), history)))
        return 0

    fp = fingerprint(data)
    previous_attempts = fingerprints.get(fp, 0)
    if previous_attempts >= MAX_SAME_APPROACH_ATTEMPTS:
        reason = (
            "se alcanzaron 3 reintentos del mismo enfoque "
            f"(4 intentos en total: {summarize_tool_call(data)})"
        )
        state.update(
            {
                "cutoff": True,
                "cutoff_reason": reason,
                "history": history,
                "fingerprints": fingerprints,
            }
        )
        save_state(session_id, state)
        register_cutoff_error(data, reason, history)
        print(json.dumps(deny_response(reason, history)))
        return 0

    tool_calls = int(state.get("tool_calls", 0))
    if tool_calls >= MAX_TOOL_CALLS:
        reason = f"se alcanzaron {MAX_TOOL_CALLS} tool calls en esta tarea"
        state.update(
            {
                "cutoff": True,
                "cutoff_reason": reason,
                "history": history,
                "fingerprints": fingerprints,
            }
        )
        save_state(session_id, state)
        register_cutoff_error(data, reason, history)
        print(json.dumps(deny_response(reason, history)))
        return 0

    summary = summarize_tool_call(data)
    history.append(summary)
    fingerprints[fp] = previous_attempts + 1
    state.update(
        {
            "tool_calls": tool_calls + 1,
            "fingerprints": fingerprints,
            "history": history,
            "cutoff": False,
            "cutoff_reason": None,
        }
    )
    save_state(session_id, state)
    print(json.dumps({"permission": "allow"}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
