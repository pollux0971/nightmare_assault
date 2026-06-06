"""core.narrative.opening_variation_prompt — 把 OpeningVariationContract 接到 story 開場（P5/P6）。

三件事（全部抽象、無具體名詞錨點）：

  apply_contract_to_context  把契約注入序幕 ctx（story 的 user 訊息會看到），並補一條「只能執行契約、
                             不可改回紙條/找人/固定姓名」的義務。
  repair_instruction         開場違規時，產生一段要求重寫的指令（列出違規 + 重申契約）。
  fallback_opening           repair 後仍違規 → 產生**決定性**開場 + 乾淨 DecisionPoint（保證不含違規素材）。

設計邊界：只組 prompt / 組決定性開場，不改世界、不推 reveal、不收束劇情。
"""
from __future__ import annotations

from core.models import BeatMeta, DecisionPoint, Option
from core.narrative.opening_pools import VariationPools, default_pools
from core.narrative.opening_variation import OpeningVariationContract


# 抽象 medium → 表層名詞（fallback / 提示用；不含人名/固定物名）。
_MEDIUM_SURFACE: dict[str, str] = {
    "handwritten_note": "一張字條",
    "voice_message": "一段語音留言",
    "corrupted_log": "一段對不上時間的錯誤紀錄",
    "access_record": "一筆不該存在的門禁紀錄",
    "radio_burst": "一段斷續的無線電訊號",
    "cctv_frame": "一格監視器畫面",
    "printed_receipt": "一張剛吐出的列印單",
    "device_status": "一盞不正常的狀態燈與錯誤碼",
    "body_mark": "你手腕上不記得的編號",
    "object_placement": "一樣被刻意擺在門口的東西",
    "environmental_trace": "一道還沒乾的濕腳印",
    "npc_claim": "某個人壓低聲音說的一句話",
    "inventory_anomaly": "你口袋裡多出來、不認得的東西",
    "schedule_entry": "值班表上一行被劃掉的字",
    "map_annotation": "地圖上一個被圈起來的點",
    "photo_artifact": "一張反光裡多出一個人的照片",
    "terminal_prompt": "終端機上一行還沒清掉的指令",
    "emergency_broadcast": "一段反覆播放的廣播",
}

_INTERACTABLE_SURFACE: dict[str, str] = {
    "terminal_or_console": "一台還亮著的終端機",
    "door_or_hatch": "一扇半掩的門",
    "container_or_locker": "一個沒鎖好的櫃子",
    "control_panel": "一面閃著燈的控制盤",
    "personal_device": "一台不是你的個人裝置",
    "fixed_screen_or_monitor": "一面卡住畫面的螢幕",
    "wall_marking_or_sign": "牆上一處新刮出的記號",
    "intercom_or_speaker": "一具還通著電的對講機",
    "window_or_aperture": "一道看不清外面的窗",
    "discarded_object": "一樣被丟在角落的東西",
    "switch_or_lever": "一個扳到一半的開關",
    "body_or_remains": "一處你不敢細看的痕跡",
}


def _medium_surface(contract: OpeningVariationContract,
                    pools: VariationPools | None = None) -> str:
    """挑一個不在 forbidden_literals 內的 medium 表層名詞（避免 fallback 自我違規）。"""
    pools = pools or default_pools()
    forbidden = set(contract.forbidden_literals or [])
    phrase = _MEDIUM_SURFACE.get(contract.message_medium, "一個說不通的異常")
    if any(f and f in phrase for f in forbidden):
        return "一個說不通的異常"
    return phrase


def apply_contract_to_context(ctx: dict, contract: OpeningVariationContract,
                              pools: VariationPools | None = None) -> dict:
    """把契約注入序幕 ctx（回新 dict）。story 必須照 contract 寫開場核心素材。"""
    out = dict(ctx or {})
    out["opening_variation_contract"] = contract.to_dict()
    obligations = [
        "【開場核心素材契約】開場的「動機 / 人物錨點 / 第一則訊息載體 / 第一個可互動物」由 "
        "opening_variation_contract 決定；你只能用表層敘事把它們寫自然，不可自行改成別的素材。",
        f"動機方向：{contract.initial_goal}。",
        (f"線索載體是 message_medium=「{contract.message_medium}」——"
         f"請寫成{_medium_surface(contract, pools)}之類，"
         "絕不可改寫成手寫紙條／便條／手寫留言（除非 medium 本身就是 handwritten_note）。"),
    ]
    if contract.personal_anchor_type == "no_person_anchor":
        obligations.append("這次沒有特定的人在等你：不要硬塞一個失蹤/牽掛的人，由眼前的異常本身驅動。")
    elif contract.personal_anchor_label:
        obligations.append(f"人物錨點是抽象角色「{contract.personal_anchor_label}」——"
                           "用這個關係去寫，不要替他取一個會反覆出現的固定姓名。")
    if contract.forbidden_literals:
        obligations.append("以下字串本局禁止出現（近期已重複過）："
                           + "、".join(contract.forbidden_literals) + "。")
    if "missing_person" in (contract.forbidden_archetypes or []):
        obligations.append("本局禁止把開場寫成『尋找失蹤的人』。")

    ob = list(out.get("narrative_obligations") or [])
    out["narrative_obligations"] = obligations + ob
    base = out.get("instruction", "")
    out["instruction"] = (base + " " if base else "") + (
        "嚴格遵守 opening_variation_contract 的開場核心素材；不得偷換回被擋下的素材或常見的偷懶開場。")
    return out


def repair_instruction(contract: OpeningVariationContract, violations: list) -> str:
    """產生 repair 指令（附加到 ctx.instruction，要求重寫違規開場）。"""
    lines = ["你上一版開場違反了開場核心素材契約，請完整重寫一個開場，修正以下問題："]
    for v in violations:
        vtype = getattr(v, "type", None) or (v.get("type") if isinstance(v, dict) else "")
        vval = getattr(v, "value", None) or (v.get("value") if isinstance(v, dict) else "")
        if vtype == "forbidden_literal":
            lines.append(f"- 不得出現字串「{vval}」（近期已重複過）。")
        elif vtype == "forbidden_archetype":
            if vval == "missing_person":
                lines.append("- 不得把開場寫成『尋找失蹤的人』。")
            else:
                lines.append(f"- 不得使用 archetype「{vval}」。")
        elif vtype == "message_medium_mismatch":
            lines.append(f"- 訊息載體必須是 {contract.message_medium}，不得寫成手寫紙條/便條/手寫留言。")
    lines.append(f"重申：動機方向「{contract.initial_goal}」，訊息載體 {contract.message_medium}。"
                 "其餘照常自由發揮，最後停在第一個可行動選擇。")
    return "\n".join(lines)


def fallback_opening(contract: OpeningVariationContract,
                     pools: VariationPools | None = None) -> tuple[str, DecisionPoint]:
    """決定性 fallback 開場（保證不含違規素材）+ 乾淨 DecisionPoint。"""
    pools = pools or default_pools()
    surface = _medium_surface(contract, pools)
    interactable = _INTERACTABLE_SURFACE.get(contract.first_interactable_type, "一樣引起你注意的東西")
    anchor = contract.personal_anchor_label
    anchor_line = (f"是{anchor}留下的痕跡，把你拽到這裡。" if anchor
                   else "沒有誰在等你——是這個地方本身出了問題。")

    narrative = (
        "燈光不穩，空氣裡有一股說不上來的味道。你站在原地，先讓呼吸慢下來。\n\n"
        f"最先抓住你的，是{surface}。它不該在這裡，內容也對不上你記得的任何事。{anchor_line}\n\n"
        f"你心裡很清楚自己為什麼還沒走：{contract.initial_goal}。\n\n"
        f"幾步之外，{interactable}正等著你決定要不要碰它。再不動，情況只會更糟。"
    )
    dp = DecisionPoint(
        situation_recap=f"你面前有{interactable}，而{surface}讓一切都不對勁。",
        decision_type="action",
        suggested_options=[
            Option(text=f"靠近並檢查{interactable}", tone="cautious"),
            Option(text="先退一步，觀察四周有沒有別的出路", tone="evasive"),
            Option(text="直接出聲，看看這裡還有沒有別人", tone="bold"),
        ],
        free_input_hint="或描述你想做的事…",
        beat_meta=BeatMeta(beat_number=0, pacing="rising"),
        is_narration_only=False,
    )
    return narrative, dp
