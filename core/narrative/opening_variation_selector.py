"""core.narrative.opening_variation_selector — 變體池 + cooldown → OpeningVariationContract（P3/P4）。

職責：把「變體池 + 權重 + cooldown ledger」收斂成一份開場契約（純函式、零 LLM、可重現）。

  build_contract(...)          低階：給定 rng/run/ledger/pools → OpeningVariationContract。
  generate_opening_contract(db) 高階：接 DB 持久化的 cooldown，遞增 run 序號、抽契約、記錄使用、存回。

決定性：同一場 run（同 run_id + 同 run 序號）→ 同 seed → 同契約，方便 debug 反查。
fallback：cooldown 把候選擋光時 weighted_choice 自動退回完整池並標 cooldown_exhausted（不會無開場）。
"""
from __future__ import annotations

import hashlib
import logging
from random import Random

from core.narrative.opening_pools import VariationPools, default_pools
from core.narrative.opening_variation import (
    CooldownLedger,
    OpeningVariationContract,
    weighted_choice,
)

log = logging.getLogger("nightmare.opening_selector")


def _seed_from(run_id: str, run_index: int) -> int:
    """決定性 seed：避免依賴 PYTHONHASHSEED，用 sha256 穩定雜湊。"""
    h = hashlib.sha256(f"{run_id}:{run_index}".encode("utf-8")).hexdigest()
    return int(h, 16) % (2 ** 31)


def build_contract(
    *,
    rng: Random,
    current_run: int,
    ledger: CooldownLedger,
    pools: VariationPools | None = None,
    selector_seed: int | None = None,
) -> OpeningVariationContract:
    """依變體池 + cooldown 抽一份契約。

    - forbidden_archetypes 同時擋 motive 與 medium（兩者共用 archetype 命名空間，見 example：
      recent_archetypes 含 missing_person(motive) 與 handwritten_note(medium)）。
    - forbidden_literals 來自 ledger（近期具體字串，如 紙條/林晨）；直接寫進契約供 StoryAgent + gate。
    """
    pools = pools or default_pools()
    forbidden_literals = ledger.forbidden_literals(current_run)
    forbidden_archetypes = ledger.forbidden_archetypes(current_run)
    fa = set(forbidden_archetypes)

    motive, m_ex = weighted_choice(rng, pools.motive_weights, fa)
    medium, d_ex = weighted_choice(rng, pools.medium_weights, fa)
    anchor, a_ex = weighted_choice(rng, pools.anchor_weights, fa)
    interactable, i_ex = weighted_choice(rng, pools.interactable_weights, set())

    from core.narrative.opening_pools import ANCHOR_LABEL_HINTS, MOTIVE_GOAL_HINTS
    goal = MOTIVE_GOAL_HINTS.get(motive, "查清這裡到底發生了什麼")
    anchor_label = ANCHOR_LABEL_HINTS.get(anchor)
    objective = (
        f"你來這裡的核心理由：{goal}。"
        + (f"線索的來源是「{anchor_label}」相關的存在。" if anchor_label else "這次沒有特定的人在等你，驅動你的是眼前的異常本身。")
    )

    return OpeningVariationContract(
        motive_archetype=motive,
        personal_anchor_type=anchor,
        message_medium=medium,
        initial_goal=goal,
        first_interactable_type=interactable,
        personal_anchor_label=anchor_label,
        opening_objective_sentence=objective,
        forbidden_literals=forbidden_literals,
        forbidden_archetypes=forbidden_archetypes,
        recent_literals=list(ledger.recent_literals.keys()),
        recent_archetypes=list(ledger.recent_archetypes.keys()),
        cooldown_debug={
            "blocked_literals": forbidden_literals,
            "blocked_archetypes": forbidden_archetypes,
            "selector_seed": selector_seed,
            "cooldown_applied": bool(forbidden_literals or forbidden_archetypes),
            "cooldown_exhausted": bool(m_ex or d_ex or a_ex or i_ex),
        },
    )


def usage_for_contract(contract: OpeningVariationContract,
                       pools: VariationPools | None = None) -> tuple[list[str], list[str]]:
    """這份契約「用掉了哪些素材」→ 記進 cooldown。

    literals: 被選 medium 的表層字串家族（下一局避免重複同一載體用語）。
    archetypes: 被選的 motive 與 medium（兩者都進 archetype cooldown 命名空間）。
    """
    pools = pools or default_pools()
    literals = pools.literals_for_medium(contract.message_medium)
    archetypes = [contract.motive_archetype, contract.message_medium]
    return literals, archetypes


def generate_opening_contract(
    db,
    run_id: str,
    *,
    pools: VariationPools | None = None,
    extra_recent_literals: list[str] | None = None,
) -> OpeningVariationContract:
    """高階入口：接 DB cooldown 持久化，產生並記錄一份契約。

    extra_recent_literals：呼叫端（如已知本主題的專有名詞）可額外塞進 forbidden（例如把上一局
    出現的人名/物件名 seed 進來）；MVP 先不強制，保留擴充點。
    """
    pools = pools or default_pools()
    from core.narrative.opening_cooldown_store import open_store
    store = open_store(db)

    if store is None:                                # 無持久化 → 純記憶體一次性 ledger
        ledger = CooldownLedger()
        run_index = 1
    else:
        run_index = store.next_run_index()
        ledger = store.load_ledger()

    if extra_recent_literals:
        ledger.record(run_index - 1, literals=extra_recent_literals)

    seed = _seed_from(run_id, run_index)
    contract = build_contract(rng=Random(seed), current_run=run_index,
                              ledger=ledger, pools=pools, selector_seed=seed)

    # 記錄本局使用 → 存回（下一局才看得到 cooldown）
    literals, archetypes = usage_for_contract(contract, pools)
    ledger.record(run_index, literals=literals, archetypes=archetypes)
    if store is not None:
        try:
            store.save_ledger(ledger)
        except Exception as e:                       # 存回失敗不該擋開局
            log.warning("save_ledger failed: %s", e)

    log.info("opening contract: motive=%s medium=%s anchor=%s run=%d seed=%d "
             "(blocked_lit=%d blocked_arc=%d exhausted=%s)",
             contract.motive_archetype, contract.message_medium, contract.personal_anchor_type,
             run_index, seed, len(contract.forbidden_literals), len(contract.forbidden_archetypes),
             contract.cooldown_debug.get("cooldown_exhausted"))
    return contract
