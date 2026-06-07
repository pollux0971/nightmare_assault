"""core.world.actor_profile — NPC onboarding 檔案（patch v0.7 P1）。

讓 NPC 像「世界裡的人」而非資訊端點：每個 actor 帶 intro_state / 表層身分 / 別名 / 個性語氣。
**主題無關、不含隱藏真相**：player-visible 欄位絕不放 secret_core；true_name 可內部存、未 known 不顯示。
個性只改語氣，**不改任何 gate 權限**（不繞 TruthEvidenceGate、不推 reveal、不建地圖實體）。

存放：`blackboard.game_meta["npc_profiles"][npc_name]`（npc-chat 可直接讀，無需 WorldModel）。
對應 docs/02-npc-onboarding-policy.md、04-personality-context-policy.md。
"""
from __future__ import annotations

from dataclasses import dataclass, field, asdict

# intro 狀態（docs/02）
UNINTRODUCED = "unintroduced"   # 玩家還沒建立「這人看起來是誰」
INTRODUCED = "introduced"       # 有足夠表層 context，可正常互動
KNOWN = "known"                 # 反覆互動/被其他證據佐證，可用穩定 label

# personality（setup 的 NPCBible.personality）→ 表層個性描述 + 語氣（只影響說話風格）
_PERSONALITY_DESC = {
    "leader": "說話帶決斷、習慣主導；傾向先給方向，但保留自己的盤算。",
    "nervous": "語句短促、容易重複某些詞；緊張時壓低聲音，給的資訊可能不太穩。",
    "analytical": "說話精確、先問清楚再回答；偏好給可驗證的細節，不愛空泛承諾。",
    "optimistic": "語氣較緩、願意安撫人；但可能淡化危險、把話說得比實情輕。",
    "mysterious": "說話留白、常以反問回應；不輕易表態，像在掂量你。",
}
_PERSONALITY_TONE = {
    "leader": "calm", "nervous": "nervous", "analytical": "calm",
    "optimistic": "helpful", "mysterious": "evasive",
}
_PERSONALITY_POSTURE = {
    "leader": "guarded", "nervous": "fearful", "analytical": "cooperative",
    "optimistic": "cooperative", "mysterious": "manipulative",
}


@dataclass
class SpeechStyle:
    directness: str = "medium"      # low | medium | high
    emotional_tone: str = "nervous"  # calm | nervous | hostile | evasive | helpful
    verbosity: str = "balanced"     # short | balanced | detailed
    trust_posture: str = "guarded"  # guarded | cooperative | manipulative | fearful


@dataclass
class ActorProfile:
    intro_state: str = UNINTRODUCED
    display_label: str = "陌生人"   # 玩家面顯示名（未 known 時可為描述性 label）
    true_name: str | None = None    # 內部真名；未 known 不顯示給玩家
    aliases: list = field(default_factory=list)
    known_role: str = ""            # 表層身分（職業推測），**不得含隱藏真相**
    first_seen_area: str | None = None
    first_seen_context: str = ""
    surface_motive: str = ""
    relationship_to_current_goal: str = ""
    trust_hint: str = "unknown"     # guarded | cooperative | hostile | fearful | unknown
    personality_description: str = ""
    speech_style: dict = field(default_factory=lambda: asdict(SpeechStyle()))

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, d: dict | None) -> "ActorProfile":
        d = dict(d or {})
        ss = d.get("speech_style")
        if isinstance(ss, dict):
            d["speech_style"] = {**asdict(SpeechStyle()), **ss}
        return cls(**{k: v for k, v in d.items() if k in cls.__dataclass_fields__})


def _g(o, k, default=None):
    return o.get(k, default) if isinstance(o, dict) else getattr(o, k, default)


def profile_from_npc(npc, *, first_seen_area: str | None = None,
                     first_seen_context: str = "") -> ActorProfile:
    """從 NPC registry 的**公開面**建表層 profile（絕不放 secret_core / 隱藏真相）。"""
    name = _g(npc, "name") or "陌生人"
    personality = (_g(npc, "personality") or "").strip()
    profession = (_g(npc, "profession") or "").strip()
    public_face = (_g(npc, "public_face") or "").strip()
    appearance = (_g(npc, "appearance") or "").strip()
    desc = _PERSONALITY_DESC.get(personality, "說話方式一般，視情況回應。")
    style = SpeechStyle(
        emotional_tone=_PERSONALITY_TONE.get(personality, "nervous"),
        trust_posture=_PERSONALITY_POSTURE.get(personality, "guarded"))
    return ActorProfile(
        intro_state=UNINTRODUCED,
        display_label=name,                       # 本作 NPC 自帶名；未知時可改描述性 label
        true_name=name,
        aliases=[name] + ([appearance] if appearance else []),
        known_role=profession or "身分不明的人",   # 表層職業推測，非隱藏真相
        first_seen_area=first_seen_area,
        first_seen_context=first_seen_context or public_face,
        surface_motive=public_face,               # 公開面當表層動機（非 secret_core）
        relationship_to_current_goal="",
        trust_hint=_PERSONALITY_POSTURE.get(personality, "unknown"),
        personality_description=desc,
        speech_style=asdict(style))


# ── game_meta 存取（npc-chat 只有 blackboard，故 profile 存 game_meta）──────────
def _find_npc(blackboard, name: str):
    for n in (blackboard.snapshot().get("npc_registry") or []):
        if _g(n, "name") == name:
            return n
    return None


def get_npc_profile(blackboard, name: str) -> ActorProfile:
    """取（或從 registry 建立並持久化）某 NPC 的 profile。"""
    store = dict((getattr(blackboard, "game_meta", {}) or {}).get("npc_profiles") or {})
    if name in store:
        return ActorProfile.from_dict(store[name])
    npc = _find_npc(blackboard, name)
    prof = profile_from_npc(npc, ) if npc is not None else ActorProfile(
        display_label=name, true_name=name, aliases=[name])
    set_npc_profile(blackboard, name, prof)
    return prof


def set_npc_profile(blackboard, name: str, profile: ActorProfile):
    store = dict((getattr(blackboard, "game_meta", {}) or {}).get("npc_profiles") or {})
    store[name] = profile.to_dict()
    blackboard.game_meta = {**(getattr(blackboard, "game_meta", {}) or {}),
                            "npc_profiles": store}


def mark_introduced(blackboard, name: str) -> ActorProfile:
    prof = get_npc_profile(blackboard, name)
    if prof.intro_state == UNINTRODUCED:
        prof.intro_state = INTRODUCED
        set_npc_profile(blackboard, name, prof)
    return prof


# ── WorldModel actor entity 一致性（NPC registry/focus ↔ WorldModel actor）─────
_INTRO_RANK = {UNINTRODUCED: 0, INTRODUCED: 1, KNOWN: 2}


def ensure_actor_entity_from_npc_registry(world, blackboard, name: str, *,
                                          origin: str = "npc_registry"):
    """確保 WorldModel 有對應該 NPC 的 actor entity（解決 focus/registry 有但 world 無 entity）。

    欄位：id=actor.<slug> / kind=actor / label=name / props{aliases, intro_state, known_role,
    origin_kind}。已存在 → **只 merge aliases / 補 profile（不覆蓋已知狀態、intro_state 不降級）**。
    **只放公開 profile 欄位**（known_role=表層職業推測）；**不新增 fact/area/exit、不推 reveal、
    不寫 hidden truth（secret_core 永不進來）**。回傳該 actor Entity（world/name 缺 → None）。
    """
    if world is None or not name:
        return None
    from core.world.model import ACTOR, slug
    # 從公開 profile 取 aliases / intro_state / known_role（profile_from_npc 已剝除 secret_core）
    aliases, intro_state, known_role = [], UNINTRODUCED, ""
    if blackboard is not None:
        try:
            prof = get_npc_profile(blackboard, name)
            aliases = [a for a in (prof.aliases or []) if a and a != name]
            intro_state = prof.intro_state or UNINTRODUCED
            known_role = prof.known_role or ""
        except Exception:
            pass
    # 找既有 actor entity（label 雙向子字串 或 既定 id）
    e = world.find(name, kind=ACTOR) or world.get(f"actor.{slug(name)}")
    if e is not None and e.kind != ACTOR:
        e = None
    if e is None:                                        # 新建（只公開欄位、不含真相）
        props = {"origin_kind": origin}
        if aliases:
            props["aliases"] = list(aliases)
        props["intro_state"] = intro_state
        if known_role:
            props["known_role"] = known_role
        return world.register(ACTOR, name, id=f"actor.{slug(name)}", props=props, origin="npc")
    # 已存在 → merge aliases（不覆蓋）、intro_state 只升不降、known_role 缺才補
    cur = list(e.props.get("aliases") or [])
    cur_set = {a for a in cur}
    for a in aliases:
        if a not in cur_set:
            cur.append(a); cur_set.add(a)
    if cur:
        e.props["aliases"] = cur
    if _INTRO_RANK.get(intro_state, 0) > _INTRO_RANK.get(e.props.get("intro_state", UNINTRODUCED), 0):
        e.props["intro_state"] = intro_state
    if known_role and not e.props.get("known_role"):
        e.props["known_role"] = known_role
    e.props.setdefault("origin_kind", origin)
    return e


# ── First-contact 要求（docs/02 §first-contact；注入 npc-chat context）──────────
def build_first_contact_context(profile: ActorProfile, player_question: str = "") -> dict:
    """unintroduced NPC 首次接觸：回應須先自然帶出 context，再給部分答案——而非資訊端點。"""
    return {
        "is_first_contact": True,
        "must_include": [
            "npc 此刻的位置 / 正在做什麼",
            "一個表層身分線索（外觀 / 動作 / 職業感）",
            "對玩家的態度 / 姿態（戒備 / 配合 / 敵意…）",
            "若玩家問了具體問題，給**部分**答案（不必全盤托出）",
        ],
        "must_not_include": [
            "隱藏真相 / 未揭露 bible 內容",
            "把任何主張說成已確認真相",
            "新的 area / exit",
        ],
        "display_label": profile.display_label,
        "known_role": profile.known_role,
        "first_seen_context": profile.first_seen_context,
        "surface_motive": profile.surface_motive,
        "personality_description": profile.personality_description,
        "speech_style": profile.speech_style,
        "player_question": player_question,
        "rule": ("這名 NPC 還沒被介紹過：第一句回應要先把『他是誰、在哪、在做什麼、對你什麼態度』"
                 "自然帶出來（用你的個性與語氣），再給部分答案；**不要**像查資料的工具一樣只丟資訊。"
                 "介紹 NPC **不**推進真相、**不**確認任何主張為真。"),
    }
