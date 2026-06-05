"""main — Nightmare Assault 桌面入口。

啟動 pywebview 視窗，載入 ui/index.html，把 API class 暴露給前端。
需先安裝 pywebview 與系統 webview runtime（Linux: WebKitGTK）。
API key 可透過設定畫面輸入，或環境變數 OPENROUTER_API_KEY / config/config.json。
"""
import os
import sys
from pathlib import Path

import webview  # noqa: 需 pip install pywebview（+ GUI 後端）

from webview_app import API

ROOT = Path(__file__).resolve().parent


def _pick_gui():
    """選 GUI 後端：尊重 PYWEBVIEW_GUI；Linux 有 PyQt 就用 qt（避開需要 gi 的 GTK）。"""
    g = os.environ.get("PYWEBVIEW_GUI")
    if g:
        return g
    if sys.platform.startswith("linux"):
        for mod in ("PyQt6", "PyQt5"):
            try:
                __import__(mod)
                return "qt"
            except ImportError:
                continue
    return None  # macOS/Windows 用系統內建 WebView


def main():
    api = API()
    window = webview.create_window(
        "Nightmare Assault",
        str(ROOT / "ui" / "index.html"),
        js_api=api,
        width=900, height=720, min_size=(760, 600),
        background_color="#0A0A0C",
    )
    api.set_window(window)
    gui = _pick_gui()
    if gui:
        webview.start(gui=gui, debug=False)
    else:
        webview.start(debug=False)


if __name__ == "__main__":
    main()
