"""dev/demo_config_ui — 實機啟動 main.py 的視窗，自動導到「配置中心」並截圖。

無人值守示範：開真實 pywebview 視窗（與 main.py 同 API），用 evaluate_js 驅動前端
（開選單 → 開配置中心 → 選 fragment → 預覽），每步用 gnome-screenshot 截全螢幕。
產出 dev/_shots/*.png。跑完自動關窗。
"""
import subprocess
import sys
import time
from pathlib import Path

import webview

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from webview_app import API
SHOTS = ROOT / "dev" / "_shots"
SHOTS.mkdir(exist_ok=True)


def shot(name: str):
    out = SHOTS / f"{name}.png"
    subprocess.run(["gnome-screenshot", "-f", str(out)], check=False)
    print(f"[shot] {out}")


def driver(window):
    def js(code, wait=0.6):
        try:
            window.evaluate_js(code)
        except Exception as e:
            print("[js-err]", e)
        time.sleep(wait)

    time.sleep(3.0)                       # 等 WebKit/Qt 載入 + pywebview.api 注入
    js("Views.show('menu')", 1.0)
    shot("01-menu")                       # 主選單（含新按鈕「配置中心」）

    js("ConfigUI.open()", 1.6)            # 開配置中心（載 fragments + flags）
    shot("02-config-open")

    # 選 no_repetition fragment（清單第 4 項），載入內容到編輯器
    js("document.querySelectorAll('#cfg-frag-list li')[3].click()", 0.8)
    shot("03-fragment-selected")

    # 預覽編譯後 prompt（零 LLM）
    js("document.getElementById('cfg-btn-preview').click()", 1.2)
    shot("04-preview")

    time.sleep(0.5)
    try:
        window.destroy()
    except Exception:
        pass


def main():
    api = API()
    window = webview.create_window(
        "Nightmare Assault", str(ROOT / "ui" / "index.html"),
        js_api=api, width=900, height=720, background_color="#0A0A0C",
    )
    api.set_window(window)
    webview.start(driver, window, gui="qt", debug=False)


if __name__ == "__main__":
    main()
