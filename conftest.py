"""pytest 根設定（E0-S0）。

把專案根放上 import path，讓測試能 `import core` 等；並提供共用 fixtures 的放置點，
後續 story（狀態模組、agent 測試）在此擴充。
"""
from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

# ── 共用 fixtures 佔位（後續 story 擴充，例如 blackboard / mock LLM）──
# import pytest
# @pytest.fixture
# def blackboard():
#     ...
