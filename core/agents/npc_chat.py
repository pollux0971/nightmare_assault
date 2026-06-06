"""core.agents.npc_chat — NPC-Chat Agent（MC1）。

玩家對**在場 NPC** 多輪對話。組「認知卡」只投影 NPC 該知道的事——公開面 + 公開 evolving +
voice_sample + 該 NPC 在場時的近期事 + 職業 knows_about。**絕不放 real_bible / secret_core**
（結構性防暴雷，同 story C2/E2）；連扮演此 NPC 的 LLM 也拿不到秘密內容，因此無法洩漏。
玩家輸入以 `<player_action>` 包住（C3 injection 防護）。
"""
from __future__ import annotations

from typing import Any

# 認知卡可見的公開欄位（白名單；secret_core/self_aware 內幕不在其中——self_aware 只當行為旗標傳，不傳秘密內容）
_PUBLIC_FIELDS = ("name", "profession", "personality", "voice_sample", "public_face", "appearance")
_PUBLIC_EVOLVING = ("emotional_state", "relationship", "intent")
_PRESENT = ("present",)


def _get(o: Any, k: str, d=None):
    return o.get(k, d) if isinstance(o, dict) else getattr(o, k, d)


def _find_npc(npc_registry: list, name: str):
    for n in npc_registry or []:
        if _get(n, "name") == name:
            return n
    return None


def _cognition_card(npc: Any, recent: list) -> dict:
    """把單一 NPC 投影成認知卡（剝除 secret_core；self_aware 只當旗標）。"""
    card: dict = {}
    for f in _PUBLIC_FIELDS:
        v = _get(npc, f)
        if v is not None:
            card[f] = v
    card["self_aware"] = bool(_get(npc, "self_aware", False))   # 行為旗標（非秘密內容）
    ev = _get(npc, "evolving")
    if ev is not None:
        pub = {}
        for f in _PUBLIC_EVOLVING:
            val = _get(ev, f)
            if val is not None:
                pub[f] = val
        if pub:
            card["evolving"] = pub
    card["recent_when_present"] = recent
    return card


def build_npc_chat_context(blackboard: Any, npc_name: str, player_message: str,
                           history: list | None = None,
                           control_ctx: Any = None) -> dict:
    """組 npc-chat 的結構化 context（防暴雷：無 real_bible / secret_core）。

    control_ctx（NR1，optional）：NPCChatControlContext——啟用敘事控制時注入母題上限/答債/動機，
    讓 npc-chat 收與 Story Agent 同一份敘事約束（仍不含任何秘密內容）。
    """
    snap = blackboard.snapshot()
    npc = _find_npc(snap.get("npc_registry") or [], npc_name)
    recent = [b.get("narrative", "") if isinstance(b, dict) else str(b)
              for b in (snap.get("beat_window") or [])[-2:]]
    knows = (snap.get("revealed_bible") or {}).get("known_atmosphere", [])
    ctx = {
        "cognition_card": _cognition_card(npc, recent) if npc is not None else {"name": npc_name},
        "knows_about": knows,                  # 職業 + 已揭露讓他知道的事（無 real_bible）
        "rolling_summary": snap.get("rolling_summary", ""),
        "history": list(history or [])[-8:],   # 該聊天視窗近 8 輪
        # C3：玩家輸入永遠是角色的遊戲內台詞，用標籤隔離、verbatim
        "player_action": "<player_action>\n" + str(player_message) + "\n</player_action>",
    }
    # P1/P2/P4：NPC onboarding——個性語氣（恆給）+ 首次接觸要求（unintroduced 時）
    try:
        from core.world.actor_profile import (
            get_npc_profile, build_first_contact_context, UNINTRODUCED)
        prof = get_npc_profile(blackboard, npc_name)
        ctx["npc_profile"] = {
            "intro_state": prof.intro_state, "display_label": prof.display_label,
            "known_role": prof.known_role, "surface_motive": prof.surface_motive,
            "personality_description": prof.personality_description,
            "speech_style": prof.speech_style,
            "rule": "用 personality_description 與 speech_style 塑造語氣；個性只改說話風格，"
                    "**不**改變你能不能講真相（仍受 reveal/gate 約束）。",
        }
        if prof.intro_state == UNINTRODUCED:
            ctx["first_contact"] = build_first_contact_context(prof, player_message)
    except Exception:
        pass
    # NR1：敘事控制約束（只給可用母題/上限/答債/動機；無秘密內容）
    if control_ctx is not None:
        ctx["narrative_control"] = {
            "allowed_motifs": sorted(getattr(control_ctx, "allowed_terms", set()) or []),
            "forbidden_motifs": sorted(getattr(control_ctx, "forbidden_terms", set()) or []),
            "reveal_ceiling": getattr(control_ctx, "reveal_ceiling", "hinted"),
            "answer_debt_level": getattr(control_ctx, "answer_debt_level", 0),
            "active_player_motive": getattr(control_ctx, "active_motive", ""),
            "rule": ("不要發明新的組織/協定/核心機制/怪物類型；只用 allowed_motifs；"
                     "可迴避或部分回答，但 answer_debt_level≥2 時須給部分答案或具理由拒答；"
                     "揭露不得超過 reveal_ceiling。"),
        }
        # P-plumbing：NPC 可安全引用的 truth_id 白名單（NPC 只能從這裡選，不可自己猜/掃關鍵詞）
        truth_ctx = (getattr(blackboard, "game_meta", {}) or {}).get("npc_allowed_truth_refs")
        if truth_ctx:
            ctx["npc_chat_truth_context"] = {
                "allowed_truth_refs": truth_ctx.get("allowed_truth_refs", []),
                "forbidden_truth_refs": truth_ctx.get("forbidden_truth_refs", []),
                "rule": ("evidence_events 只能引用 allowed_truth_refs 裡的 truth_id；"
                         "level 不可超過該 ref 的 max_level；想講其他真相一律改為不確定語氣、不附 truth_id。"),
            }
    return ctx


def npc_is_present(blackboard: Any, npc_name: str) -> bool:
    npc = _find_npc(blackboard.snapshot().get("npc_registry") or [], npc_name)
    return npc is not None and _get(npc, "presence", "present") in _PRESENT


def _resolve_control_ctx(blackboard: Any, player_message: str):
    """NR1/NR2：從 game_meta 組 npc-chat 控制 context（含本句答債等級）。flag OFF → None。"""
    try:
        from core.constants import ENABLE_NARRATIVE_CONTROL
        if not ENABLE_NARRATIVE_CONTROL:
            return None
        from core.narrative.npc_chat_control import control_context_from_meta
        meta = getattr(blackboard, "game_meta", {}) or {}
        debt_level = 0
        try:
            from core.narrative.answer_debt import classify_question
            topic = classify_question(player_message)
            if topic:
                debt_level = int((meta.get("answer_debt") or {}).get(topic, 0))
        except Exception:
            debt_level = 0
        return control_context_from_meta(meta, answer_debt_level=debt_level)
    except Exception:
        return None


def _parse_npc_response(raw: str):
    """HC1：把 LLM 輸出解析成 NPCChatResponse。JSON → 結構化；否則包成純文字（向後相容）。"""
    import json
    from core.narrative.npc_chat_control import NPCChatResponse
    text = (raw or "").strip()
    # 去 markdown fence
    if text.startswith("```"):
        text = text.strip("`")
        if text[:4].lower() == "json":
            text = text[4:]
        text = text.strip()
    try:
        data = json.loads(text)
        if isinstance(data, dict) and (data.get("reply") or data.get("visible_reply")):
            ed = data.get("entity_delta")              # 容錯：非 list → []（不讓對話解析失敗）
            return NPCChatResponse(
                visible_reply=str(data.get("reply") or data.get("visible_reply") or "").strip(),
                answer_status=str(data.get("answer_status", "partial")),
                evidence_events=list(data.get("evidence_events") or []),
                new_lore_terms=list(data.get("new_lore_terms") or []),
                used_truth_ids=list(data.get("used_truth_ids") or []),
                blocked_or_uncertain_claims=list(data.get("blocked_or_uncertain_claims") or []),
                entity_delta=ed if isinstance(ed, list) else [])
    except Exception:
        pass
    # 純文字 fallback：無法判定 new lore（設空，靠 sanitizer 把關）。
    # answer_status 不可一律當 partial——明顯迴避不該被記為「已付答債」或洩證據。
    if not text:
        return NPCChatResponse(visible_reply=text, answer_status="none")
    _EVASIVE = ("不知道", "不確定", "不能說", "不清楚", "沒辦法說", "無可奉告",
                "別問", "我不知", "說不清", "不記得", "不告訴", "問別人")
    status = "evasion" if any(e in text for e in _EVASIVE) else "partial"
    return NPCChatResponse(visible_reply=text, answer_status=status)


def run_npc_chat_structured(caller: Any, blackboard: Any, npc_name: str, player_message: str,
                            history: list | None = None):
    """HC1：結構化 NPC 對話。回 NPCChatResponse；經 NPCChatControlGate（repair once → safe fallback）。

    flag OFF / 無 contract → 仍回 NPCChatResponse（純文字包裝），行為等同 MC1。
    """
    from core.narrative.npc_chat_control import (
        NPCChatControlGate, safe_fallback_reply, cap_evidence_to_ceiling)
    control_ctx = _resolve_control_ctx(blackboard, player_message)
    ctx = build_npc_chat_context(blackboard, npc_name, player_message, history,
                                 control_ctx=control_ctx)
    resp = _parse_npc_response(caller.call("npc-chat", ctx))

    if control_ctx is not None:
        gate = NPCChatControlGate()
        flags = gate.validate(control_ctx, resp)
        if flags:                                   # 違規 → repair once
            try:
                repair_ctx = {**ctx, "repair_instruction":
                              "移除未授權的新世界觀名詞，只用既有敘事契約回答；保留情緒與部分答案。"}
                resp2 = _parse_npc_response(caller.call("npc-chat", repair_ctx))
                resp = resp2 if not gate.validate(control_ctx, resp2) else safe_fallback_reply(player_message)
            except Exception:
                resp = safe_fallback_reply(player_message)
        # evidence 揭露上限封頂（防 NPC 跳級暴雷）
        cap_evidence_to_ceiling(resp, getattr(control_ctx, "reveal_ceiling", "hinted"))

    # NR7：表層消毒
    try:
        from core.narrative.sanitizer import sanitize_text
        allowed = getattr(control_ctx, "allowed_terms", None) if control_ctx else None
        resp.visible_reply = sanitize_text(resp.visible_reply, allowed_terms=allowed)
    except Exception:
        pass
    # P2：首次接觸成功（有實際回覆）→ intro_state unintroduced → introduced
    try:
        if (resp.visible_reply or "").strip():
            from core.world.actor_profile import mark_introduced
            mark_introduced(blackboard, npc_name)
    except Exception:
        pass
    return resp


def run_npc_chat(caller: Any, blackboard: Any, npc_name: str, player_message: str,
                 history: list | None = None) -> str:
    """跑一輪 NPC 對話，回傳 NPC 回覆文字（向後相容；內部走結構化 HC1 路徑）。"""
    return run_npc_chat_structured(caller, blackboard, npc_name, player_message, history).visible_reply


# ── MC2：聊天退出三向分流（完整 → cold；近期濃縮 3–4 句 → story hot context）─────
def _trim(s: str, n: int = 60) -> str:
    s = (s or "").strip().replace("\n", " ")
    return s if len(s) <= n else s[:n] + "…"


def condense_chat(npc_name: str, history: list | None) -> str:
    """把一場聊天濃縮成 3–4 句摘要（規則版，零 LLM；供進 story hot context）。"""
    hist = history or []
    npc_lines = [h.get("content", "") for h in hist if h.get("role") == "npc"][-3:]
    player_lines = [h.get("content", "") for h in hist if h.get("role") == "player"]
    if not npc_lines and not player_lines:
        return f"你與{npc_name}交談了幾句，但沒問出什麼。"
    parts = [f"你與{npc_name}談過。"]
    if player_lines:
        parts.append(f"你問了關於「{_trim(player_lines[-1], 24)}」的事。")
    for line in npc_lines[:3]:
        if line.strip():
            parts.append(f"他說：{_trim(line)}")
    return " ".join(parts[:4])                      # 至多 4 句


def apply_chat_exit(blackboard: Any, npc_name: str, history: list | None,
                    signal_bus: Any = None) -> str:
    """退出聊天室：近期濃縮寫進 recent_chat_digest（story 下個 beat 可見），發 CHATROOM_CLOSED。"""
    from core.constants import EVT_CHATROOM_CLOSED
    digest = condense_chat(npc_name, history)
    try:
        blackboard.write("compactor", "recent_chat_digest", digest)
    except Exception:
        blackboard.recent_chat_digest = digest      # 保底：直接設
    if signal_bus is not None:
        try:
            signal_bus.publish(EVT_CHATROOM_CLOSED, npc_name)
        except Exception:
            pass
    return digest
