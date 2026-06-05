"""底座冒煙測試（E0-S0）：確認 core 套件可匯入、pytest 框架可運作。

這是專案第一個測試；它存在的意義是讓「pytest 綠燈」成為真實的 exit 0
（pytest 在 0 個測試時會回 exit 5，故底座需至少一個通過的測試）。
"""
import importlib


def test_core_importable():
    core = importlib.import_module("core")
    assert core.__version__ == "0.0.0"


def test_core_subpackages_importable():
    for sub in ("agents", "state", "memory", "llm", "persistence"):
        importlib.import_module(f"core.{sub}")


def test_harness_alive():
    assert True
