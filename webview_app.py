"""webview_app — pywebview 後端 API（U16）。

把前端（HTML/CSS/JS）接到 BeatLoop：前端透過 window.pywebview.api.xxx() 呼叫，
後端在背景 thread 跑 LLM、逐 token 用 window.evaluate_js("NA.xxx(...)") 推給前端。
**API key 永不到前端**：所有 OpenRouter 請求都由這裡（Python）發。

此模組刻意不在頂層 import webview，故無 GUI 環境也能 import/單元測試（main.py 才載入 webview）。
"""
from __future__ import annotations

import json
import threading
import datetime
from pathlib import Path

from core.agents.base import SkillCaller, SkillLoader
from core.blackboard import Blackboard
from core.llm.client import OpenRouterClient
from core.llm.parser import NARRATIVE_CHUNK, CONTINUE_PAUSE
from core.orchestrator_loop import BeatLoop
from core.persistence.db import Database
from core.signal import SignalBus

ROOT = Path(__file__).resolve().parent
CONFIG_PATH = ROOT / "config" / "config.json"
STORAGE = ROOT / "storage"

# 三層模型分層 → OpenRouter 模型字串（可被 config.agent_models 覆寫）
DEFAULT_AGENT_MODELS = {
    "setup": ["anthropic/claude-3.5-sonnet", "openai/gpt-4o-mini"],
    "orchestrator": ["openai/gpt-4o-mini"],
    "warden": ["openai/gpt-4o-mini"],
    "story": ["google/gemini-flash-1.5", "openai/gpt-4o-mini"],
    "compactor": ["google/gemini-flash-1.5", "openai/gpt-4o-mini"],
    "npc-chat": ["openai/gpt-4o-mini"],
    "dreaming": ["openai/gpt-4o-mini"],
    "offstage-fate": ["openai/gpt-4o-mini"],
}
DEFAULT_TEMPS = {"setup": 0.7, "story": 0.8, "warden": 0.4,
                 "orchestrator": 0.4, "compactor": 0.7}


# ── config（key 存本機，gitignore）─────────────────────────────────────────
def load_config(path: Path = CONFIG_PATH) -> dict:
    import os
    cfg: dict = {}
    if path.is_file():
        try:
            cfg = json.loads(path.read_text(encoding="utf-8"))
        except Exception:
            cfg = {}
    if not cfg.get("api_key"):
        env = os.environ.get("OPENROUTER_API_KEY")
        if env:
            cfg["api_key"] = env
    cfg.setdefault("base_url", "https://openrouter.ai/api/v1")
    cfg.setdefault("agent_models", DEFAULT_AGENT_MODELS)
    cfg.setdefault("timeout", 90)
    return cfg


def save_config(cfg: dict, path: Path = CONFIG_PATH) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(cfg, ensure_ascii=False, indent=2), encoding="utf-8")


# ── API（pywebview js_api）─────────────────────────────────────────────────
class API:
    def __init__(self, window=None, config_path: Path = CONFIG_PATH):
        self._window = window
        self._config_path = config_path
        self._config = load_config(config_path)
        self._loop: BeatLoop | None = None
        self._state = "idle"
        self._last_error = None
        self._lock = threading.Lock()
        self._transcript: list[str] = []     # 累積每個 beat 的逐字稿（供複製）
        self._chat_history: dict = {}         # MC3：每個 NPC 的聊天視窗 {npc: [{role,content}]}

    def set_window(self, window):
        self._window = window

    # ── 前端推送 ──
    def _push(self, fn: str, *args):
        payload = ",".join(json.dumps(a, ensure_ascii=False) for a in args)
        js = f"window.NA && NA.{fn}({payload})"
        if self._window is not None:
            try:
                self._window.evaluate_js(js)
            except Exception:
                pass

    def _set_state(self, state: str):
        self._state = state
        self._push("onStatus", self.get_game_state())

    _STAGE_LABELS = {
        "setup": "建構世界真相",
        "warden": "守門人裁決",
        "kernel": "推進判定",
        "orchestrator": "揭露閘門",
        "story": "編織夢魘",
    }

    def _on_progress(self, stage: str):
        self._push("onProgress", {"stage": stage,
                                  "label": self._STAGE_LABELS.get(stage, stage)})

    def _on_story_event(self, ev):
        if ev.type == NARRATIVE_CHUNK:
            self._push("appendToken", ev.text)
        elif ev.type == CONTINUE_PAUSE:
            self._push("onContinue")

    # ── 設定 ──
    def check_config(self) -> dict:
        return {"configured": bool(self._config.get("api_key"))}

    def save_config(self, cfg: dict) -> dict:
        try:
            merged = {**self._config, **cfg}
            save_config(merged, self._config_path)
            self._config = load_config(self._config_path)
            return {"ok": True}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def test_model(self, agent: str, model: str) -> dict:
        try:
            client = self._make_client()
            r = client.call(agent, "ping", "Reply with: OK", 0.0)
            return {"ok": bool(r.success), "latency_ms": r.latency_ms, "error": r.error}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    # ── 開局 / 載入 ──
    def start_game(self, opts: dict) -> dict:
        if not self._config.get("api_key"):
            return {"ok": False, "error": "尚未設定 API key"}
        self._loop = self._make_loop()
        self._transcript = []                 # 新局重置逐字稿
        threading.Thread(target=self._run_start, args=(opts or {},), daemon=True).start()
        return {"ok": True, "run_id": self._loop.run_id}

    def submit_decision(self, text: str, input_path: str = "free_text") -> dict:
        if self._loop is None:
            return {"accepted": False, "error": "尚未開局"}
        if self._state == "generating":
            return {"accepted": False, "error": "上一個分鏡尚在生成"}
        threading.Thread(target=self._run_step, args=(text, input_path), daemon=True).start()
        return {"accepted": True}

    def continue_narration(self) -> None:
        self._push("onStatus", self.get_game_state())

    def validate_custom_input(self, fields: dict) -> dict:
        # MVP：簡單非空檢查（自訂角色四欄）。完整 Light 檢查為後續。
        ok = bool(fields.get("name"))
        return {"ok": ok, "reason": "" if ok else "姓名必填"}

    # ── 查詢 ──
    def get_inventory(self) -> list:
        if self._loop is None:
            return []
        inv = self._loop.bb.snapshot().get("shared_inventory") or {}
        items = inv.get("items", []) if isinstance(inv, dict) else []
        # 不洩漏 is_key_item；held_by 可露（玩家知道道具在誰身上）
        return [{"id": i.get("id"), "name": i.get("name"), "brief": i.get("brief"),
                 "held_by": i.get("held_by")}
                for i in items if isinstance(i, dict)]

    def get_status(self) -> dict:
        if self._loop is None:
            return {}
        snap = self._loop.bb.snapshot()
        return {"protagonist": snap.get("protagonist"),
                "beat_number": self._loop.beat_number}

    def _record(self, narrative, dp, opening=None):
        parts: list[str] = []
        if opening:
            parts += [str(x) for x in opening]
        if narrative:
            parts.append(str(narrative).strip())
        if dp is not None:
            recap = getattr(dp, "situation_recap", "") or ""
            if recap:
                parts.append("〔" + recap + "〕")
            for o in getattr(dp, "suggested_options", []) or []:
                parts.append("  - " + getattr(o, "text", ""))
        if parts:
            n = len(self._transcript) + 1
            self._transcript.append(f"── 第 {n} 分鏡 ──\n" + "\n".join(parts))

    def get_transcript(self) -> dict:
        """整場故事逐字稿（開場 + 每 beat 的敘事與決策），供前端複製。"""
        return {"text": "\n\n".join(self._transcript)}

    def get_debug_state(self) -> dict:
        """progress kernel 進度檢視（scene/phase/obligations/clues/inventory/npcs），供 debug 面板。"""
        loop = self._loop
        gs = getattr(loop, "_game_state", None) if loop else None
        if gs is None:
            return {"kernel": bool(getattr(loop, "use_kernel", False)), "state": None}
        from core import progress_bridge as _bridge
        d = _bridge.debug_state(gs)
        d["kernel"] = loop.use_kernel
        return d

    def get_game_state(self) -> dict:
        loop = self._loop
        snap = loop.bb.snapshot() if loop else {}
        return {
            "run_id": loop.run_id if loop else None,
            "state": self._state,
            "busy": self._state in ("setting_up", "generating", "saving", "loading"),
            "beat_number": loop.beat_number if loop else 0,
            "current_location": (snap.get("scene_registry") or {}).get("current_location"),
            "last_error": self._last_error,
        }

    def list_saves(self) -> list:
        try:
            return self._make_db().list_runs()
        except Exception:
            return []

    def save_game_now(self, label: str) -> dict:
        if self._loop is None:
            return {"ok": False, "error": "尚未開局"}
        try:
            self._loop.db.add_save_point(self._loop.run_id, self._loop.beat_number, label or "存檔點")
            return {"ok": True}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def load_game(self, run_id: str) -> dict:
        # MVP：回報最新 beat 摘要（完整還原為後續 refine）。
        try:
            db = self._make_db()
            runs = {r["run_id"]: r for r in db.list_runs()}
            if run_id not in runs:
                return {"ok": False, "error": "找不到存檔"}
            return {"ok": True, "run": runs[run_id]}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    # ── 聊天室（MC3）──
    def list_present_npcs(self) -> list:
        """列出在場 NPC（供開聊天室）；只露公開面，不洩 secret。"""
        if self._loop is None:
            return []
        snap = self._loop.bb.snapshot()
        out = []
        for n in snap.get("npc_registry") or []:
            if isinstance(n, dict) and n.get("presence", "present") == "present":
                out.append({"name": n.get("name"), "public_face": n.get("public_face"),
                            "profession": n.get("profession")})
        return out

    def open_chatroom(self, npc: str) -> dict:
        if self._loop is None:
            return {"ok": False, "error": "尚未開局"}
        from core.agents.npc_chat import npc_is_present
        if not npc_is_present(self._loop.bb, npc):
            return {"ok": False, "error": "對方不在場"}
        self._chat_history.setdefault(npc, [])
        if self._loop.bus is not None:
            try:
                from core.constants import EVT_CHATROOM_OPENED
                self._loop.bus.publish(EVT_CHATROOM_OPENED, npc)
            except Exception:
                pass
        return {"ok": True, "npc": npc, "history": self._chat_history[npc]}

    def send_chat(self, npc: str, text: str) -> dict:
        if self._loop is None:
            return {"ok": False, "error": "尚未開局"}
        try:
            from core.agents.npc_chat import run_npc_chat_structured
            history = self._chat_history.setdefault(npc, [])
            # HC1：結構化 NPC 對話（gate + repair once + safe fallback + 消毒）
            resp = run_npc_chat_structured(self._loop.caller, self._loop.bb, npc, text, history)
            reply = resp.visible_reply
            history.append({"role": "player", "content": text})
            history.append({"role": "npc", "content": reply})
            # NR0/NR1/HC1/P-plumbing：NPC 結構化 evidence 走 gate → bridge → revealed_bible
            try:
                self._loop.bridge_npc_evidence(resp)
                self._loop.note_npc_answer(text, resp.answer_status)   # 答債：partial 償還、evasion 不償還
            except Exception:
                pass
            try:                                  # 完整紀錄入 cold（SQLite）
                self._loop.db.add_chat_log(self._loop.run_id, npc, self._loop.beat_number, "player", text)
                self._loop.db.add_chat_log(self._loop.run_id, npc, self._loop.beat_number, "npc", reply)
            except Exception:
                pass
            return {"ok": True, "reply": reply}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def close_chatroom(self, npc: str) -> dict:
        """退出聊天室：近期濃縮進 story hot context（MC2）。"""
        if self._loop is None:
            return {"ok": False, "error": "尚未開局"}
        try:
            from core.agents.npc_chat import apply_chat_exit
            history = self._chat_history.get(npc, [])
            digest = apply_chat_exit(self._loop.bb, npc, history, signal_bus=self._loop.bus)
            return {"ok": True, "digest": digest}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    # ── 配置中心 UI（P5：draft → preview → activate；preview 不呼 LLM）──
    def _config_store(self):
        return self._make_db().config_store()

    def config_overview(self) -> dict:
        """配置總覽：active profile + story 編譯 prompt hash + enabled fragments + flags。"""
        try:
            from core.config.composer import PromptComposer
            store = self._config_store()
            profile = store.active_profile()
            compiled = PromptComposer(store).preview("story", profile, {})
            flags = self.list_feature_flags(profile)
            return {"ok": True, "active_profile": profile, "default_profile": store.default_profile(),
                    "profiles": [p["profile_name"] for p in store.list_profiles()],
                    "story": {"prompt_hash": compiled.prompt_hash,
                              "enabled_fragments": compiled.enabled_fragments,
                              "model_settings": compiled.model_settings},
                    "flags": flags}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def list_prompt_fragments(self, agent: str = "story", profile: str | None = None) -> list:
        try:
            store = self._config_store()
            profile = profile or store.active_profile()
            out = []
            for fr in store.get_bound_fragments(agent, profile):
                d = store.get_fragment(fr["fragment_key"]) or {}
                draft = store.get_latest_draft(fr["fragment_key"])
                out.append({"fragment_key": fr["fragment_key"], "title": fr.get("title"),
                            "category": fr.get("category"), "content": d.get("content"),
                            "status": d.get("status"), "version": d.get("version"),
                            "sort_order": fr.get("sort_order"),
                            "has_draft": bool(draft), "draft_version": draft["version"] if draft else None})
            return out
        except Exception:
            return []

    def get_prompt_fragment(self, fragment_key: str) -> dict:
        try:
            store = self._config_store()
            frag = store.get_fragment(fragment_key) or {}
            draft = store.get_latest_draft(fragment_key)
            return {"ok": True, "fragment": frag, "latest_draft": draft}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def save_prompt_draft(self, fragment_key: str, content: str) -> dict:
        """存 draft（不影響 active run，直到 activate_prompt_draft）。"""
        try:
            version = self._config_store().save_fragment_draft(fragment_key, content)
            return {"ok": True, "version": version, "status": "draft"}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def preview_prompt(self, agent: str = "story", profile: str | None = None,
                       draft_key: str | None = None, draft_content: str | None = None) -> dict:
        """預覽編譯後 prompt（**零 LLM**）。可帶 draft 覆寫單一 fragment 而不寫 DB。"""
        try:
            from core.config.composer import PromptComposer
            store = self._config_store()
            profile = profile or store.active_profile()
            overrides = None
            if draft_key:
                content = draft_content
                if content is None:
                    d = store.get_latest_draft(draft_key)
                    content = d["content"] if d else None
                if content is not None:
                    overrides = {draft_key: content}
            compiled = PromptComposer(store).preview(agent, profile, {}, overrides=overrides)
            return {"ok": True, "llm_called": False, "profile": profile,
                    "compiled_prompt": compiled.compiled_prompt, "prompt_hash": compiled.prompt_hash,
                    "enabled_fragments": compiled.enabled_fragments, "warnings": compiled.warnings,
                    "missing_required": compiled.missing_required}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def activate_prompt_draft(self, fragment_key: str, version: int | None = None) -> dict:
        """把 draft 提升為 active（此前 active prompt 不受 draft 影響）。"""
        try:
            v = self._config_store().activate_fragment_draft(fragment_key, version)
            if v is None:
                return {"ok": False, "error": "無可啟用的 draft"}
            return {"ok": True, "activated_version": v}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def list_feature_flags(self, profile: str | None = None) -> list:
        try:
            from core.config.flags import HARDCODED_DEFAULTS, resolve_flag
            store = self._config_store()
            profile = profile or store.active_profile()
            out = []
            for name in HARDCODED_DEFAULTS:
                out.append({"name": name, "value": bool(resolve_flag(name, store, profile)),
                            "db_value": store.get_flag(name, profile)})
            return out
        except Exception:
            return []

    def set_feature_flag(self, name: str, value: bool, profile: str | None = None) -> dict:
        try:
            self._config_store().set_flag(name, 1 if value else 0, profile)
            return {"ok": True}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def get_active_profile(self) -> dict:
        try:
            store = self._config_store()
            return {"ok": True, "active_profile": store.active_profile(),
                    "profiles": [p["profile_name"] for p in store.list_profiles()]}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    def set_active_profile(self, profile_name: str) -> dict:
        try:
            self._config_store().set_active_profile(profile_name)
            return {"ok": True, "active_profile": profile_name}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    # ── 背景執行 ──
    def _run_start(self, opts: dict):
        try:
            self._set_state("setting_up")
            res = self._loop.start(
                opts, on_event=self._on_story_event, on_progress=self._on_progress,
                on_opening=lambda lines: [self._push("onOpening", ln) for ln in lines],
            )
            self._record(res["narrative"], res["decision_point"],
                         opening=res.get("opening_sequence"))
            self._push("onDecision", res["decision_point"].model_dump())
            self._push("onBeatComplete")
            self._set_state("awaiting_decision")
        except Exception as e:
            self._fail(e)

    def _run_step(self, text: str, input_path: str):
        try:
            self._set_state("generating")
            out = self._loop.step(text, input_path=input_path, on_event=self._on_story_event,
                                  on_progress=self._on_progress)
            if out.get("committed_event"):           # kernel 流程：HUD 顯示推進事件 + delta
                self._push("onProgressInfo", {"event": out["committed_event"],
                                              "delta": out.get("progress_delta", [])})
            w = out.get("warden")                    # UB1：破格技能被封頂 → UI 提示侷限
            if w is not None and getattr(w, "skill_claim", None):
                self._push("onSkillLimit", {"skill": w.skill_claim,
                                            "limitation": w.skill_limitation or ""})
            self._record(out["narrative"], out["decision_point"])
            self._push("onDecision", out["decision_point"].model_dump())
            self._push("onBeatComplete")
            if out.get("ended"):
                self._push("onEnding", out.get("ending") or {})
                self._set_state("idle")
            else:
                self._set_state("awaiting_decision")
        except Exception as e:
            self._fail(e)

    def _fail(self, e: Exception):
        self._last_error = str(e)
        self._set_state("error")
        self._push("onError", {"message": str(e), "recoverable": True})

    # ── 工廠 ──
    def _make_client(self) -> OpenRouterClient:
        c = self._config
        return OpenRouterClient({
            "api_key": c["api_key"], "base_url": c["base_url"],
            "agent_models": c.get("agent_models", DEFAULT_AGENT_MODELS),
            "timeout": c.get("timeout", 90),
        })

    def _make_db(self) -> Database:
        STORAGE.mkdir(parents=True, exist_ok=True)
        return Database(str(STORAGE / "game.db"))

    def _make_loop(self) -> BeatLoop:
        caller = SkillCaller(self._make_client(), SkillLoader(str(ROOT / "skills")),
                             temperature_by_agent=DEFAULT_TEMPS)
        run_id = "run-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
        return BeatLoop(caller, Blackboard(), self._make_db(), SignalBus(), run_id=run_id)
