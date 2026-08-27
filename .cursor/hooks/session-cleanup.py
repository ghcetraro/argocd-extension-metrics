#!/usr/bin/env python3
"""Limpia el estado de límites al cerrar una conversación."""

from __future__ import annotations

import json
import sys
from pathlib import Path

STATE_DIR = Path(__file__).resolve().parent / "state"


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        print("{}")
        return 0

    session_id = data.get("session_id") or data.get("conversation_id")
    if session_id:
        safe = "".join(c if c.isalnum() or c in "-_" else "_" for c in session_id)
        state_file = STATE_DIR / f"{safe}.json"
        if state_file.exists():
            state_file.unlink()

    print("{}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
