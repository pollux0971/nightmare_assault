"""core.llm.parser — StreamParser（U08，承重牆）。

story 串流輸出 = `敘事文字 <<<DECISION>>> { JSON }`，或旁白型以 `<<<CONTINUE>>>` 暫停。
本 parser 是**唯一解析者**（前端不解析）。逐 token 餵入，滑動視窗偵測分隔符
（分隔符可能被拆成多 token），分離 narrative 與 decision JSON，三級 repair 保證遊戲不掛。

事件型別：NARRATIVE_CHUNK / CONTINUE_PAUSE / DECISION_READY / NARRATION_END
"""
from __future__ import annotations

import json
import re
from collections import namedtuple

from core.constants import DELIM_CONTINUE, DELIM_DECISION
from core.models import BeatMeta, DecisionPoint, Option

ParseEvent = namedtuple("ParseEvent", ["type", "text"])

# ── LLM 列舉漂移容錯（L1 repair 的一環）─────────────────────────────────────
# 真實模型常吐合理但不在列舉內的 tone（analytical/direct/observant…）或 pacing/audio_cue。
# 這些只是裝飾性語氣/節奏，不該讓整個 DecisionPoint 被拒。在驗證前正規化到合法值。
_TONE_VALID = {"cautious", "bold", "evasive", "aggressive"}
_TONE_MAP = {
    "direct": "bold", "decisive": "bold", "assertive": "bold", "confident": "bold",
    "brave": "bold", "determined": "bold", "firm": "bold", "proactive": "bold",
    "forceful": "aggressive", "confrontational": "aggressive", "hostile": "aggressive",
    "reckless": "aggressive", "violent": "aggressive", "angry": "aggressive",
    "avoidant": "evasive", "defensive": "evasive", "wary": "evasive", "retreat": "evasive",
    "flee": "evasive", "guarded": "evasive", "suspicious": "evasive",
    # 其餘（analytical/observant/methodical/curious/calm…）→ cautious
}
_PACING_VALID = {"calm", "rising", "peak"}
_AUDIO_VALID = {"normal", "silence", "sting", "swell"}
_DECISION_VALID = {"action", "dialogue"}


def _coerce_enums(data: dict) -> dict:
    """把 LLM 輸出中漂移的列舉值正規化到合法值（容錯，不改 DecisionPoint 契約）。"""
    dt = str(data.get("decision_type", "")).strip().lower()
    if dt not in _DECISION_VALID:
        data["decision_type"] = "action"
    opts = data.get("suggested_options")
    if isinstance(opts, list):
        for o in opts:
            if isinstance(o, dict):
                t = str(o.get("tone", "")).strip().lower()
                o["tone"] = t if t in _TONE_VALID else _TONE_MAP.get(t, "cautious")
    bm = data.get("beat_meta")
    if isinstance(bm, dict):
        p = str(bm.get("pacing", "")).strip().lower()
        if p and p not in _PACING_VALID:
            bm["pacing"] = "rising" if p in ("tense", "climax", "urgent") else "calm"
        a = str(bm.get("audio_cue", "")).strip().lower()
        if a and a not in _AUDIO_VALID:
            bm["audio_cue"] = "normal"
    return data

NARRATIVE_CHUNK = "NARRATIVE_CHUNK"
CONTINUE_PAUSE = "CONTINUE_PAUSE"
DECISION_READY = "DECISION_READY"
NARRATION_END = "NARRATION_END"

_HOLD = max(len(DELIM_DECISION), len(DELIM_CONTINUE)) - 1  # 尾端可能是半個分隔符，先扣住


def _ev(t, text=None):
    return ParseEvent(t, text)


class StreamParser:
    DELIM_CONTINUE = DELIM_CONTINUE
    DELIM_DECISION = DELIM_DECISION

    def __init__(self, llm_repair=None, beat_number: int = 0):
        self._llm_repair = llm_repair      # callable(broken_json, schema_hint) -> str | None
        self._beat_number = beat_number
        self._buf = ""                     # narrative 模式未提交緩衝
        self._decision_buf = ""            # decision 模式累積的 JSON 文字
        self._narrative = ""               # 已確定為 narrative 的累積（已吐，不回收）
        self._mode = "narrative"           # narrative | decision

    # ── 唯讀檢視 ──
    @property
    def narrative(self) -> str:
        return self._narrative

    @property
    def saw_decision(self) -> bool:
        return self._mode == "decision"

    # ── 串流餵入 ──
    def feed(self, token: str) -> list:
        events: list = []
        if self._mode == "decision":
            self._decision_buf += token
            return events

        self._buf += token
        # 反覆掃描，處理一個 token 內可能有多個分隔符的情況
        while True:
            di = self._buf.find(self.DELIM_DECISION)
            ci = self._buf.find(self.DELIM_CONTINUE)
            present = [i for i in (di, ci) if i != -1]
            if not present:
                break
            first = min(present)
            if di != -1 and first == di:
                pre = self._buf[:di]
                if pre:
                    self._narrative += pre
                    events.append(_ev(NARRATIVE_CHUNK, pre))
                self._decision_buf = self._buf[di + len(self.DELIM_DECISION):]
                self._buf = ""
                self._mode = "decision"
                events.append(_ev(DECISION_READY))
                return events
            else:  # CONTINUE
                pre = self._buf[:ci]
                if pre:
                    self._narrative += pre
                    events.append(_ev(NARRATIVE_CHUNK, pre))
                events.append(_ev(CONTINUE_PAUSE))
                self._buf = self._buf[ci + len(self.DELIM_CONTINUE):]
                continue

        # 無完整分隔符：吐出安全前綴，尾端 _HOLD 字留著（可能是半個分隔符）
        if len(self._buf) > _HOLD:
            emit = self._buf[:len(self._buf) - _HOLD]
            if emit:
                self._narrative += emit
                events.append(_ev(NARRATIVE_CHUNK, emit))
                self._buf = self._buf[len(emit):]
        return events

    # ── 結束 ──
    def finalize(self, expect_narration: bool = False) -> DecisionPoint:
        """串流結束時呼叫，回傳驗證後的 DecisionPoint（內含三級 repair）。"""
        if self._mode == "decision":
            return self._parse_decision()
        # 沒看到 DECISION：把尾端緩衝沖刷成 narrative
        if self._buf:
            self._narrative += self._buf
            self._buf = ""
        if expect_narration:
            return self._narration_only()
        return self._fallback()  # 該有決策卻沒有 → 保底

    # ── 三級 repair ──
    def _parse_decision(self) -> DecisionPoint:
        raw = self._decision_buf
        # L1 程式碼修復
        dp = self._try(self._l1_repair(raw)) or self._try(raw)
        if dp is not None:
            return dp
        # L2 LLM repair
        if self._llm_repair is not None:
            try:
                repaired = self._llm_repair(raw, self._schema_hint())
            except Exception:
                repaired = None
            if repaired:
                dp = self._try(self._l1_repair(repaired)) or self._try(repaired)
                if dp is not None:
                    return dp
        # L3 fallback（保證可玩）
        return self._fallback()

    def _try(self, s: str):
        if not s:
            return None
        try:
            data = json.loads(s)
        except Exception:
            return None
        if not isinstance(data, dict):
            return None
        data.setdefault("beat_meta", {"beat_number": self._beat_number})
        if isinstance(data.get("beat_meta"), dict):
            data["beat_meta"].setdefault("beat_number", self._beat_number)
        _coerce_enums(data)            # LLM 列舉漂移容錯（tone/pacing/audio_cue/decision_type）
        try:
            return DecisionPoint.model_validate(data)
        except Exception:
            return None

    @staticmethod
    def _l1_repair(s: str) -> str:
        if not s:
            return s
        # 取第一個 { 到最後一個 }
        i, j = s.find("{"), s.rfind("}")
        if i != -1 and j != -1 and j > i:
            s = s[i:j + 1]
        # 智慧引號 → 直引號
        s = (s.replace("“", '"').replace("”", '"')
              .replace("‘", "'").replace("’", "'"))
        # 去尾逗號  ,}  ,]
        s = re.sub(r",\s*([}\]])", r"\1", s)
        return s

    def _schema_hint(self) -> str:
        return ('DecisionPoint{situation_recap:str, decision_type:"action"|"dialogue", '
                'suggested_options:[{text:str,tone:"cautious"|"bold"|"evasive"|"aggressive"}], '
                'beat_meta:{beat_number:int}, is_narration_only:bool}')

    def _fallback(self) -> DecisionPoint:
        return DecisionPoint(
            situation_recap="你站在原地，四周的黑暗像是短暫失去了形狀。你必須做出選擇。",
            decision_type="action",
            suggested_options=[
                Option(text="繼續觀察", tone="cautious"),
                Option(text="往前走", tone="bold"),
                Option(text="呼喚附近的人", tone="evasive"),
            ],
            beat_meta=BeatMeta(beat_number=self._beat_number),
            is_narration_only=False,
        )

    def _narration_only(self) -> DecisionPoint:
        return DecisionPoint(
            situation_recap="",
            decision_type="action",
            suggested_options=[],
            beat_meta=BeatMeta(beat_number=self._beat_number),
            is_narration_only=True,
        )
