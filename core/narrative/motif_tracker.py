"""core.narrative.motif_tracker — 母題冷卻 + 動機心跳（NR5 / NR6，敘事控制 v0.2）。

- MotifTracker（NR5）：逐場景追蹤母題使用次數；超 max_uses_per_scene → 下次須演化/揭露/進冷卻。
- MotiveHeartbeat（NR6）：追蹤距上次提及主角動機的 beat 數；逾 N beat → 加動機提醒義務。

零 LLM、規則版。對應 dev/CONTRACTS.md §十四（MotifCooldown / MotiveHeartbeat）。
"""
from __future__ import annotations

from dataclasses import dataclass, field

# 預設母題詞庫（docs/06）：motif_key → 偵測用中文關鍵詞
_DEFAULT_MOTIF_VOCAB: dict[str, list[str]] = {
    "stopped_clock": ["掛鐘", "時鐘", "停在", "整點", "指針", "11:55", "報時"],
    "water_reflection": ["水窪", "倒影", "水面", "映出", "鏡"],
    "salt_and_rust_smell": ["鏽", "鐵鏽", "鹽", "腥味", "潮濕"],
    "metal_scraping": ["刮擦", "金屬", "拖行", "刮過", "摩擦"],
    "red_light": ["紅光", "紅燈", "暗紅", "血紅"],
    "distorted_npc_face": ["扭曲", "沒有臉", "變形的臉", "面具"],
}


def extract_motifs(text: str, vocab: dict[str, list[str]] | None = None) -> set[str]:
    """從敘事文字偵測出現的母題 key（粗略關鍵詞比對）。"""
    text = text or ""
    vocab = vocab or _DEFAULT_MOTIF_VOCAB
    return {key for key, terms in vocab.items() if any(t in text for t in terms)}


@dataclass
class MotifTracker:
    """逐場景母題使用計數。超用 → 該母題進「須演化」名單。"""
    max_uses_per_scene: int = 2
    uses: dict[str, int] = field(default_factory=dict)
    history: list[set] = field(default_factory=list)     # 每 beat 出現的母題集合（供停滯偵測）

    def register_beat(self, motifs) -> None:
        s = set(motifs or [])
        for m in s:
            self.uses[m] = self.uses.get(m, 0) + 1
        self.history.append(s)

    def is_overused(self, motif: str) -> bool:
        return self.uses.get(motif, 0) >= self.max_uses_per_scene

    def build_blocked_motifs(self) -> list[str]:
        """回傳已達上限、下次須演化/揭露/暫停的母題。"""
        return sorted(m for m, c in self.uses.items() if c >= self.max_uses_per_scene)

    def reset_scene(self) -> None:
        """換場景：清空計數與歷史（母題冷卻以場景為單位）。"""
        self.uses.clear()
        self.history.clear()

    def stagnant_motifs(self, window: int = 3) -> list[str]:
        """連續 window 個 beat 都出現的母題（停滯，QualityGate 告警用）。"""
        if len(self.history) < window:
            return []
        recent = self.history[-window:]
        common = set(recent[0])
        for s in recent[1:]:
            common &= s
        return sorted(common)


@dataclass
class MotiveHeartbeat:
    """動機心跳（NR6）：距上次提及主角動機的 beat 數；逾 max → 須提醒。"""
    max_beats_without_motive: int = 3
    beats_since_motive: int = 0

    def register_beat(self, referenced_motive: bool) -> None:
        self.beats_since_motive = 0 if referenced_motive else self.beats_since_motive + 1

    def required(self) -> bool:
        return self.beats_since_motive >= self.max_beats_without_motive

    def to_dict(self) -> dict:
        return {"beats_since_motive": self.beats_since_motive,
                "max": self.max_beats_without_motive}


def motif_block_instruction(blocked: list[str]) -> str:
    return ("以下母題本場景已重複多次，這次若要再用，**必須**讓它揭露新資訊、改變狀態或變成可行動線索，"
            f"否則換別的意象：{', '.join(blocked)}。")
