"""core.agents.story — Story Agent + 串流（U13，防暴雷承重牆）。

story 是唯一每個 beat 都運作的生成 agent。它逐 beat 把世界、NPC、玩家決定編織成
恐怖分鏡並串流輸出，寫到決策點停筆。本模組的核心職責是**結構性防暴雷（C2/E2）**：

  build_story_context 只從 blackboard.snapshot() 取 **revealed_bible**（已揭露子集）
  與安全欄位，**絕不**把 real_bible / 任何 secret_core 放進 story 看得到的 context。
  story 結構上拿不到完整真相 → 它沒有東西可暴雷（不是靠 prompt 自律，是靠資料隔離）。

另外把玩家輸入以 <player_action> 標籤 verbatim 包住（C3 prompt-injection 防護、C4 不摘要）。
"""
from __future__ import annotations

from typing import Any

from core.llm.parser import StreamParser
from core.models import DecisionPoint


# ── NPC 在場時，story 可見的「公開面」白名單 ──────────────────────────────────
# 嚴格白名單：只有這些鍵會被放進 story context。
# secret_core / self_aware / emergent_lies / personal_arc 等內幕一律不在其中。
_NPC_PUBLIC_FIELDS = ("name", "profession", "personality", "voice_sample",
                      "public_face", "appearance", "presence", "alignment")

# evolving 中可公開給 story 的子集（NPC 當前對外表現出的狀態）。
# revealed_layers / emergent_lies / personal_arc 屬內幕，排除。
_NPC_PUBLIC_EVOLVING_FIELDS = ("emotional_state", "relationship", "intent")

# 在場判定：這些 presence 視為「在場景中」，story 才看得到該 NPC。
_PRESENT_VALUES = ("present",)


def _public_npc_view(npc: Any) -> dict:
    """把單一 NPC 投影成 story 可見的公開面（剝除 secret_core 等內幕）。"""
    if isinstance(npc, dict):
        get = npc.get
    else:
        get = lambda k, d=None: getattr(npc, k, d)  # noqa: E731

    view: dict = {}
    for field in _NPC_PUBLIC_FIELDS:
        val = get(field, None)
        if val is not None:
            view[field] = val

    evolving = get("evolving", None)
    if evolving is not None:
        if isinstance(evolving, dict):
            eget = evolving.get
        else:
            eget = lambda k, d=None: getattr(evolving, k, d)  # noqa: E731
        pub_evolving: dict = {}
        for field in _NPC_PUBLIC_EVOLVING_FIELDS:
            ev = eget(field, None)
            if ev is not None:
                pub_evolving[field] = ev
        if pub_evolving:
            view["evolving"] = pub_evolving
    return view


def _present_npcs(npc_registry: list) -> list[dict]:
    """挑出在場 NPC，並各自投影成公開面。"""
    out: list[dict] = []
    for npc in npc_registry or []:
        presence = npc.get("presence") if isinstance(npc, dict) else getattr(npc, "presence", None)
        if presence in _PRESENT_VALUES:
            out.append(_public_npc_view(npc))
    return out


def build_story_context(
    blackboard: Any,
    player_decision: str,
    directive: Any = None,
    newly_revealed: Any = None,
) -> dict:
    """組 story 的結構化 context（防暴雷結構性保證 C2/E2）。

    **只**從 snapshot() 取已揭露/安全欄位：
      revealed_bible、rolling_summary、ledger、beat_window、在場 NPC 的公開面、
      scene 當前位置、directive、newly_revealed。

    **絕對不放** real_bible 或任何 secret_core —— story 結構上看不到完整真相，
    因此沒有東西可暴雷。這是靠資料隔離而非 prompt 自律。

    玩家輸入以 <player_action> 標籤 verbatim 包住（C3 injection 防護、C4 不摘要）。
    """
    snap = blackboard.snapshot()

    # scene 當前位置（只取座標，不洩漏其餘世界結構）
    scene = snap.get("scene_registry") or {}
    if isinstance(scene, dict):
        current_location = scene.get("current_location")
    else:
        current_location = getattr(scene, "current_location", None)

    context: dict = {
        # 只讀已揭露子集，絕不放 real_bible
        "revealed_bible": snap.get("revealed_bible", {}),
        "rolling_summary": snap.get("rolling_summary", ""),
        "ledger": snap.get("ledger", []),
        "beat_window": snap.get("beat_window", []),
        "npcs_present": _present_npcs(snap.get("npc_registry", [])),
        "current_location": current_location,
        "directive": directive,
        "newly_revealed": newly_revealed,
        # C3：玩家輸入永遠是「角色的遊戲內行動/台詞」，用標籤隔離、verbatim 不摘要（C4）
        "player_action": "<player_action>\n" + str(player_decision) + "\n</player_action>",
    }
    return context


def run_story(
    caller: Any,
    blackboard: Any,
    player_decision: str,
    beat_number: int,
    directive: Any = None,
    newly_revealed: Any = None,
    expect_narration: bool = False,
    on_event: Any = None,
    context_override: Any = None,
    system_override: Any = None,
) -> tuple[str, DecisionPoint]:
    """執行一個 beat 的 story 串流，回傳 (narrative, DecisionPoint)。

    流程：
      1. build_story_context（防暴雷投影）。
      2. StreamParser 逐 token 餵入 caller.stream("story", context) 的輸出。
      3. finalize(expect_narration) → DecisionPoint（內含三級 repair，保證可玩）。
      4. 把本 beat 寫入 blackboard：beat_window（story 可 append）、
         turn_context.narrative_output（story 可寫）。
      5. 回 (parser.narrative, dp)。

    on_event：可選 callable(ParseEvent)，逐事件回呼（前端串流：NARRATIVE_CHUNK→吐字、
    CONTINUE_PAUSE→暫停）。
    """
    # kernel 模式：直接用 ContextBuilder 的最小檢視（仍只含 revealed，結構性防暴雷）；
    # 否則用既有防暴雷投影。
    if context_override is not None:
        context = dict(context_override)
        context.setdefault("player_action",
                           "<player_action>\n" + (player_decision or "") + "\n</player_action>")
    else:
        context = build_story_context(
            blackboard,
            player_decision,
            directive=directive,
            newly_revealed=newly_revealed,
        )

    # system_override（配置中心 P4）：非 None 才傳，避免破壞不接受該參數的測試 caller。
    stream_kwargs = {} if system_override is None else {"system_override": system_override}
    parser = StreamParser(beat_number=beat_number)
    for token in caller.stream("story", context, **stream_kwargs):
        for ev in parser.feed(token):
            if on_event is not None:
                on_event(ev)
    dp = parser.finalize(expect_narration=expect_narration)

    narrative = parser.narrative

    # ── 寫回 blackboard（story 權限：beat_window append / turn_context）──────
    beat_record = {
        "beat_number": beat_number,
        "narrative": narrative,
        "is_narration_only": dp.is_narration_only,
    }
    blackboard.write("story", "beat_window", beat_record)
    blackboard.write("story", "turn_context.narrative_output", narrative)

    return narrative, dp


# ── 反重複（P3）：forbidden id → 不得再被當成「選項」重新提出的措辭 ──────────────
# 把 forbidden_repeats 的事件 id 對映成中文/英文的「再次提供同一行動」措辭，供斷言比對。
_FORBIDDEN_PHRASE_MAP: dict[str, tuple[str, ...]] = {
    "open_door": ("開門", "開那扇門", "推開門", "打開門", "打開那扇門", "踹開門",
                  "轉動門把", "open the door", "open door"),
}


def _forbidden_phrases(forbidden_id: str) -> list[str]:
    """由 forbidden 事件 id 推出「不得再次作為選項」的措辭集合（已知映射 + token 衍生）。"""
    norm = "".join(ch for ch in str(forbidden_id).lower() if not ch.isdigit()).strip("_")
    phrases: list[str] = []
    for key, words in _FORBIDDEN_PHRASE_MAP.items():
        if key in norm:
            phrases.extend(words)
    # token 衍生（去掉常見意圖前綴），讓未登錄 id 也有基本防護
    tokens = [t for t in norm.split("_") if t and t not in ("ask", "again", "re")]
    if len(tokens) >= 2:
        phrases.append(" ".join(tokens))      # e.g. "open door"
    return phrases


def _option_texts(decision_point: Any) -> list[str]:
    opts = getattr(decision_point, "suggested_options", None) or []
    out: list[str] = []
    for o in opts:
        t = o.get("text") if isinstance(o, dict) else getattr(o, "text", None)
        if t:
            out.append(str(t))
    return out


def assert_respects_forbidden(decision_point: Any, forbidden_repeats: list[str]) -> None:
    """反重複斷言（P3）：DecisionPoint 的建議選項不得**重新提供**任何 forbidden_repeats 行動。

    例：forbidden_repeats=[ask_open_door_101] → 選項不得再出現「開門/推開門/轉動門把…」。
    命中 → raise AssertionError（供測試與執行期 E-class 守門；P6 回歸沿用）。
    """
    options = [t.lower() for t in _option_texts(decision_point)]
    for fid in forbidden_repeats or []:
        for phrase in _forbidden_phrases(fid):
            p = phrase.lower()
            for opt in options:
                if p and p in opt:
                    raise AssertionError(
                        f"反重複違規：forbidden_repeats 含「{fid}」，但選項仍重新提供「{phrase}」")


def assert_no_spoiler(narrative: str, forbidden_contents: list[str]) -> None:
    """E2 防暴雷斷言：narrative 不得含任何未揭露碎片的 content 子字串。

    供測試與執行期使用。命中任一未揭露碎片 → raise AssertionError。
    """
    text = narrative or ""
    for fragment in forbidden_contents or []:
        if fragment and fragment in text:
            raise AssertionError(
                f"防暴雷違規：narrative 含未揭露碎片內容「{fragment}」"
            )
