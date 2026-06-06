// api.js — 按鈕接線 + 呼叫後端 window.pywebview.api（U16/U18/U19）
(function () {
  const q = (id) => document.getElementById(id);
  const api = () => (window.pywebview && window.pywebview.api) || null;

  // pywebview 注入 api 前的保護
  function ready(cb) {
    if (api()) return cb();
    window.addEventListener("pywebviewready", cb);
    setTimeout(() => api() && cb(), 400); // 後備
  }

  const App = {
    async submit(text, inputPath) {
      if (!text || !api()) return;
      Streaming.clearNarrative();
      q("options").innerHTML = "";
      q("recap").textContent = "";
      const free = q("free-text"); if (free) free.value = "";
      Streaming.hudReset();      // 顯示生成管線 HUD
      Streaming.showWaiting();   // 故事還沒吐字前一直跑等待動畫
      await api().submit_decision(text, inputPath || "free_text");
    },
  };
  window.App = App;

  function bootSettings() {
    q("btn-save-cfg").onclick = async () => {
      const key = q("cfg-key").value.trim();
      if (!key) { q("cfg-msg").textContent = "請輸入 API key"; return; }
      const r = await api().save_config({ api_key: key });
      if (r && r.ok) Views.show("menu");
      else q("cfg-msg").textContent = (r && r.error) || "儲存失敗";
    };
  }

  function bootMenu() {
    q("btn-new").onclick = () => Views.show("new");
    q("btn-settings").onclick = () => Views.show("settings");
    const cfgBtn = q("btn-config");
    if (cfgBtn) cfgBtn.onclick = () => ConfigUI.open();
    q("btn-continue").onclick = async () => {
      const saves = await api().list_saves();
      if (saves && saves.length) { Views.show("game"); }
      else { q("game-status") && (q("game-status").textContent = "沒有存檔"); }
    };
  }

  function bootNew() {
    q("btn-back-menu").onclick = () => Views.show("menu");
    q("btn-start").onclick = async () => {
      const theme = q("new-theme").value.trim();
      if (!theme) { q("new-msg").textContent = "請輸入主題"; return; }
      const opts = { theme,
        protagonist_name: q("new-name").value.trim() || "林默",
        npc_count: parseInt(q("new-npc").value, 10) || 2 };
      Views.show("loading");
      ritual();
      Streaming.clearNarrative();
      Streaming.hudReset();      // setup→揭露→story 進度顯示於 HUD
      const r = await api().start_game(opts);
      // 切到遊戲畫面由 onOpening/appendToken 觸發（switchToGameIfLoading）
      if (!r || !r.ok) { Views.show("new"); q("new-msg").textContent = (r && r.error) || "開局失敗"; }
    };
  }

  function bootGame() {
    q("btn-submit").onclick = () => App.submit(q("free-text").value.trim(), "free_text");
    q("free-text").addEventListener("keydown", (e) => {
      if (e.key === "Enter") App.submit(q("free-text").value.trim(), "free_text");
    });
    q("btn-continue-narration").onclick = () => {
      q("btn-continue-narration").classList.add("hidden");
      api() && api().continue_narration();
    };
    q("btn-menu").onclick = () => Views.show("menu");
    q("btn-debug").onclick = async () => {
      const d = await api().get_debug_state();
      q("debug-json").textContent = JSON.stringify(d, null, 2);
      q("dlg-debug").showModal();
    };
    q("btn-transcript").onclick = async () => {
      const r = await api().get_transcript();
      q("transcript-text").value = (r && r.text) || "（還沒有內容）";
      q("copy-msg").textContent = "";
      q("dlg-transcript").showModal();
    };
    q("btn-copy-transcript").onclick = async () => {
      const ta = q("transcript-text");
      let ok = false;
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(ta.value); ok = true;
        }
      } catch (e) { ok = false; }
      if (!ok) {                      // Qt WebEngine fallback
        ta.focus(); ta.select();
        try { ok = document.execCommand("copy"); } catch (e) { ok = false; }
        window.getSelection && window.getSelection().removeAllRanges();
      }
      q("copy-msg").textContent = ok ? "✓ 已複製整場故事到剪貼簿" : "複製失敗，請手動全選複製";
    };
    const chatBtn = q("btn-chat");
    if (chatBtn) chatBtn.onclick = () => ChatUI.open();
    q("btn-inv").onclick = async () => {
      const items = await api().get_inventory();
      const ul = q("inv-list"); ul.innerHTML = "";
      (items || []).forEach((it) => {
        const li = document.createElement("li");
        li.textContent = (it.name || "?") + (it.brief ? " — " + it.brief : "");
        ul.appendChild(li);
      });
      if (!items || !items.length) ul.innerHTML = "<li>（空）</li>";
      q("dlg-inv").showModal();
    };
    q("btn-save").onclick = async () => {
      const ul = q("save-list"); ul.innerHTML = "";
      const saves = await api().list_saves();
      (saves || []).forEach((s) => {
        const li = document.createElement("li");
        li.textContent = s.run_id + " · 第 " + s.current_beat + " 分鏡";
        ul.appendChild(li);
      });
      q("dlg-save").showModal();
    };
    q("btn-do-save").onclick = async () => {
      await api().save_game_now(q("save-label").value.trim());
      q("save-label").value = "";
      q("btn-save").click();
    };
  }

  function bootEnding() {
    q("btn-ending-menu").onclick = () => Views.show("menu");
  }

  // ── 聊天室（MC3）──
  const ChatUI = {
    _npc: null,
    async open() {
      if (!api()) return;
      this._npc = null;
      const npcs = await api().list_present_npcs();
      const ul = q("chat-npc-list"); ul.innerHTML = "";
      q("chat-log").style.display = "none";
      q("chat-input-row").style.display = "none";
      q("chat-with").textContent = "";
      if (!npcs || !npcs.length) {
        ul.innerHTML = "<li class='hint'>（這裡沒有人可以交談）</li>";
      } else {
        npcs.forEach((n) => {
          const li = document.createElement("li");
          li.style.cssText = "padding:.4rem;cursor:pointer;border-bottom:1px solid var(--border,#2a2a34)";
          li.textContent = (n.name || "?") + (n.profession ? "（" + n.profession + "）" : "");
          li.onclick = () => this.enter(n.name);
          ul.appendChild(li);
        });
      }
      q("dlg-chat").showModal();
    },
    async enter(npc) {
      const r = await api().open_chatroom(npc);
      if (!r || !r.ok) { q("chat-with").textContent = (r && r.error) || "無法交談"; return; }
      this._npc = npc;
      q("chat-npc-list").style.display = "none";
      q("chat-with").textContent = "· " + npc;
      const log = q("chat-log"); log.style.display = "flex"; log.innerHTML = "";
      (r.history || []).forEach((m) => this._append(m.role, m.content));
      q("chat-input-row").style.display = "flex";
      q("chat-text").value = ""; q("chat-text").focus();
    },
    _append(role, content) {
      const log = q("chat-log");
      const d = document.createElement("div");
      d.style.cssText = "font-size:.9rem;line-height:1.6;" +
        (role === "player" ? "color:var(--text,#e6e1d6);text-align:right" : "color:var(--blood-lt,#b83232)");
      d.textContent = (role === "player" ? "你：" : "") + content;
      log.appendChild(d); log.scrollTop = log.scrollHeight;
    },
    async send() {
      const t = q("chat-text").value.trim();
      if (!t || !this._npc) return;
      this._append("player", t);
      q("chat-text").value = "";
      const r = await api().send_chat(this._npc, t);
      this._append("npc", (r && r.ok) ? r.reply : "（……）");
    },
    async back() {
      if (this._npc) await api().close_chatroom(this._npc);   // 退出→濃縮進 story context
      this._npc = null;
      q("chat-npc-list").style.display = "block";
      this.open();
    },
  };
  window.ChatUI = ChatUI;

  function bootChat() {
    const send = q("chat-send"), back = q("chat-back"), txt = q("chat-text");
    if (send) send.onclick = () => ChatUI.send();
    if (back) back.onclick = () => ChatUI.back();
    if (txt) txt.addEventListener("keydown", (e) => { if (e.key === "Enter") ChatUI.send(); });
    const dlg = q("dlg-chat");
    if (dlg) dlg.addEventListener("close", () => {     // 關閉視窗也觸發退出濃縮
      if (ChatUI._npc && api()) { api().close_chatroom(ChatUI._npc); ChatUI._npc = null; }
    });
  }

  // ── 配置中心 UI（v0.8 Agent Configuration Center；preview 零 LLM）──────────
  const ConfigUI = {
    _key: null, _dirty: false, _blocksAgent: "story", _agents: [],

    _setDirty(v) { this._dirty = v; const d = q("cfg-dirty"); if (d) d.textContent = v ? "● 未儲存" : ""; },
    _toast(msg) {
      const t = q("cfg-toast"); if (!t) return;
      t.textContent = msg; t.classList.add("show");
      clearTimeout(this._tt); this._tt = setTimeout(() => t.classList.remove("show"), 2200);
    },
    _msg(s) { const m = q("cfg-msg"); if (m) m.textContent = s || ""; },
    _opt(sel, items, cur) {
      sel.innerHTML = "";
      items.forEach((v) => { const o = document.createElement("option");
        o.value = v; o.textContent = v; if (v === cur) o.selected = true; sel.appendChild(o); });
    },

    async open() {
      if (!api()) return;
      this._setDirty(false); this._key = null;
      const ov = await api().config_overview();
      this._opt(q("cfg-profile"), (ov && ov.profiles) || [], ov && ov.active_profile);
      q("cfg-profile").onchange = async (e) => { await api().set_active_profile(e.target.value); this.open(); };
      q("cfg-active-hash").textContent = ov && ov.story ? "story hash " + ov.story.prompt_hash : "";
      // agent 清單（給 Prompt Blocks / Preview / Test 的 agent 選擇器）
      const mo = await api().agent_models_overview();
      this._agents = ((mo && mo.agents) || []).map((a) => a.agent);
      [["cfg-blocks-agent", "story"], ["cfg-preview-agent", "story"], ["cfg-test-agent", "story"]]
        .forEach(([id, def]) => { const s = q(id); if (s) {
          this._opt(s, this._agents.length ? this._agents : [def], this._blocksAgent || def); } });
      q("cfg-blocks-agent").onchange = (e) => { this._blocksAgent = e.target.value; this.loadBlocks(); };
      q("cfg-preview-agent").onchange = () => this.preview();
      this.renderModels(mo);
      await this.loadBlocks();
      await this.preview();
      await this.loadFlags();
      this._msg(""); this.tab("models");
      q("dlg-config").showModal();
    },

    tab(name) {
      document.querySelectorAll(".cfg-tab").forEach((b) => b.classList.toggle("on", b.dataset.tab === name));
      document.querySelectorAll(".cfg-panel").forEach((p) => p.classList.toggle("on", p.dataset.panel === name));
    },

    // ── P0 Agent Models ──────────────────────────────────────────────────
    renderModels(mo) {
      const body = q("cfg-models-body"); body.innerHTML = "";
      if (!mo || !mo.ok) { body.innerHTML = "<tr><td colspan=7 class=hint>讀取失敗</td></tr>"; return; }
      (mo.agents || []).forEach((a) => {
        const tr = document.createElement("tr");
        const cell = (html) => { const td = document.createElement("td"); td.innerHTML = html; return td; };
        const tdA = cell(""); tdA.textContent = a.agent;
        const tdP = cell("<input type=text>"); tdP.querySelector("input").value = a.primary || "";
        const tdF = cell("<input type=text>"); tdF.querySelector("input").value = (a.fallbacks || []).join(", ");
        const tdT = cell("<input type=number step=0.1 min=0 max=2>"); tdT.className = "num";
        tdT.querySelector("input").value = (a.temperature ?? "");
        const tdM = cell("<input type=number min=1>"); tdM.className = "num";
        tdM.querySelector("input").value = (a.max_tokens ?? "");
        const tdE = cell("<input type=checkbox>"); tdE.querySelector("input").checked = !!a.enabled;
        const tdTest = cell("<button class=ghost style='padding:.2rem .5rem'>Test</button>");
        tr.append(tdA, tdP, tdF, tdT, tdM, tdE, tdTest);
        tr.dataset.agent = a.agent;
        tr.querySelectorAll("input").forEach((i) => i.oninput = () => this._setDirty(true));
        tdTest.querySelector("button").onclick = async () => {
          const m = tdP.querySelector("input").value.trim();
          this._toast("測試中…");
          const r = await api().test_model(a.agent, m);
          this._toast(r && r.ok ? `✓ ${a.agent} ${r.latency_ms}ms` : `✗ ${(r && r.error) || "失敗"}`);
        };
        body.appendChild(tr);
      });
    },
    _collectModels() {
      return [...q("cfg-models-body").querySelectorAll("tr")].filter((tr) => tr.dataset.agent).map((tr) => {
        const inp = tr.querySelectorAll("input");
        const fb = inp[1].value.split(",").map((s) => s.trim()).filter(Boolean);
        return { agent: tr.dataset.agent, primary: inp[0].value.trim(), fallbacks: fb,
          temperature: inp[2].value === "" ? null : parseFloat(inp[2].value),
          max_tokens: inp[3].value === "" ? null : parseInt(inp[3].value, 10),
          enabled: inp[4].checked };
      });
    },
    async saveModels() {
      const r = await api().save_agent_models(this._collectModels());
      if (r && r.ok) { this._setDirty(false); this._toast("✓ 已儲存模型設定"); this._msg(""); }
      else { this._msg("儲存失敗：" + ((r && r.error) || "")); }
    },

    // ── P1 Prompt Blocks ─────────────────────────────────────────────────
    async loadBlocks() {
      const agent = this._blocksAgent || "story";
      const r = await api().list_prompt_blocks(agent);
      const body = q("cfg-blocks-body"); body.innerHTML = "";
      const blocks = (r && r.blocks) || [];
      if (!blocks.length) { body.innerHTML = "<tr><td colspan=10 class=hint>此 agent/profile 無 prompt block</td></tr>"; return; }
      blocks.forEach((b) => {
        const tr = document.createElement("tr");
        const td = (txt) => { const c = document.createElement("td"); c.textContent = txt == null ? "—" : txt; return c; };
        const tdOn = document.createElement("td");
        const cb = document.createElement("input"); cb.type = "checkbox"; cb.checked = !!b.enabled;
        cb.onchange = async () => {
          const res = await api().set_fragment_enabled(agent, b.fragment_key, cb.checked);
          this._toast(res && res.ok ? (cb.checked ? "已啟用 block" : "已停用 block") : "切換失敗");
        };
        tdOn.appendChild(cb);
        const tdEdit = document.createElement("td");
        const eb = document.createElement("button"); eb.className = "ghost"; eb.style.padding = ".2rem .5rem";
        eb.textContent = "Edit"; eb.onclick = () => this.edit(b);
        tdEdit.appendChild(eb);
        const upd = b.updated_at ? String(b.updated_at).slice(0, 16).replace("T", " ") : "—";
        tr.append(td(b.sort_order), tdOn, td(b.fragment_key), td(b.title), td(b.category),
          td((b.status || "") + (b.has_draft ? " ✎" : "")), td(b.version), td(upd), td(b.preview), tdEdit);
        body.appendChild(tr);
      });
    },
    async edit(b) {
      this._key = b.fragment_key;
      const r = await api().get_prompt_fragment(b.fragment_key);
      const content = (r && r.ok && r.fragment && r.fragment.content) || "";
      q("cfg-frag-key").textContent = b.fragment_key;
      q("cfg-frag-editor").value = content;
      q("cfg-editor-wrap").style.display = "";
      this._msg("");
    },
    async saveDraft() {
      if (!this._key) { this._msg("先在表格點 Edit 選一個 block"); return; }
      const r = await api().save_prompt_draft(this._key, q("cfg-frag-editor").value);
      this._msg(r && r.ok ? "已存草稿 v" + r.version + "（按「啟用草稿」才生效）" : "存草稿失敗：" + ((r && r.error) || ""));
      this.loadBlocks();
    },
    async activate() {
      if (!this._key) { this._msg("先點 Edit 選一個 block"); return; }
      const r = await api().activate_prompt_draft(this._key);
      if (r && r.ok) { this._toast("✓ 已啟用 v" + r.activated_version); this._msg(""); }
      else { this._msg("啟用失敗：" + ((r && r.error) || "")); }
      this.loadBlocks();
      const ov = await api().config_overview();
      if (ov && ov.story) q("cfg-active-hash").textContent = "story hash " + ov.story.prompt_hash;
    },

    // ── P2 Compiled Preview（零 LLM）──────────────────────────────────────
    async preview() {
      const agent = (q("cfg-preview-agent") && q("cfg-preview-agent").value) || "story";
      const r = await api().preview_prompt(agent, null);
      const pre = q("cfg-preview"), meta = q("cfg-preview-meta");
      if (r && r.ok) {
        const cp = r.compiled_prompt || "";
        pre.textContent = cp;
        const chars = cp.length, tok = Math.round(chars / 1.8);
        q("cfg-preview-hash").textContent = "hash " + r.prompt_hash + (r.llm_called ? "" : " · 零 LLM");
        meta.textContent = `enabled fragments: ${(r.enabled_fragments || []).length}　|　`
          + `${chars} chars　~${tok} tokens (est)`
          + (r.missing_required && r.missing_required.length ? "　⚠ 缺必填：" + r.missing_required.join(", ") : "");
      } else { pre.textContent = ""; meta.textContent = "預覽失敗：" + ((r && r.error) || ""); }
    },

    async loadFlags() {
      const flags = await api().list_feature_flags();
      const ul = q("cfg-flags"); ul.innerHTML = "";
      (flags || []).forEach((f) => {
        const li = document.createElement("li"); li.style.padding = ".25rem 0";
        const cb = document.createElement("input");
        cb.type = "checkbox"; cb.checked = !!f.value; cb.id = "flag-" + f.name;
        cb.onchange = async () => { await api().set_feature_flag(f.name, cb.checked); this._toast("flag 已更新"); };
        const lbl = document.createElement("label");
        lbl.htmlFor = cb.id; lbl.textContent = " " + f.name; lbl.style.cursor = "pointer";
        li.appendChild(cb); li.appendChild(lbl); ul.appendChild(li);
      });
    },

    close() {
      if (this._dirty && !confirm("Agent Models 有未儲存的修改，確定離開？")) return;
      this._setDirty(false); q("dlg-config").close();
    },
  };
  window.ConfigUI = ConfigUI;

  function bootConfig() {
    document.querySelectorAll(".cfg-tab").forEach((b) => b.onclick = () => ConfigUI.tab(b.dataset.tab));
    const wire = (id, fn) => { const el = q(id); if (el) el.onclick = fn; };
    wire("cfg-models-save", () => ConfigUI.saveModels());
    wire("cfg-btn-draft", () => ConfigUI.saveDraft());
    wire("cfg-btn-activate", () => ConfigUI.activate());
    wire("cfg-btn-preview", () => ConfigUI.preview());
    wire("cfg-close", () => ConfigUI.close());
  }

  // 序章儀式：恐怖台詞緩慢浮現又淡去
  const RITUAL = [
    "你聽見自己的心跳。",
    "黑暗在你睜眼之前就已存在。",
    "有些門，打開就再也關不上。",
    "深呼吸。然後，進入。",
  ];
  function ritual() {
    const el = q("loading-line"); let i = 0;
    function step() {
      if (NA._view !== "loading") return;
      el.textContent = RITUAL[i % RITUAL.length]; el.style.opacity = "1";
      setTimeout(() => { el.style.opacity = "0"; }, 2200);
      i++; setTimeout(step, 3600);
    }
    step();
  }

  ready(() => {
    bootSettings(); bootMenu(); bootNew(); bootGame(); bootEnding(); bootConfig(); bootChat();
    if (api()) {
      api().check_config().then((r) => {
        Views.show(r && r.configured ? "menu" : "settings");
      });
    } else {
      Views.show("settings");
    }
  });
})();
