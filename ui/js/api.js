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

  // ── 配置中心 UI（P5：draft → preview → activate；preview 不呼 LLM）──
  const ConfigUI = {
    _key: null,
    async open() {
      if (!api()) return;
      const ov = await api().config_overview();
      const sel = q("cfg-profile"); sel.innerHTML = "";
      ((ov && ov.profiles) || []).forEach((p) => {
        const o = document.createElement("option");
        o.value = p; o.textContent = p; if (p === ov.active_profile) o.selected = true;
        sel.appendChild(o);
      });
      sel.onchange = async () => { await api().set_active_profile(sel.value); this.open(); };
      q("cfg-active-hash").textContent = ov && ov.story ? "active hash: " + ov.story.prompt_hash : "";
      await this.loadFragments();
      await this.loadFlags();
      q("cfg-msg").textContent = ""; q("cfg-preview").textContent = "";
      q("cfg-preview-hash").textContent = "";
      q("dlg-config").showModal();
    },
    async loadFragments() {
      const list = await api().list_prompt_fragments("story");
      const ul = q("cfg-frag-list"); ul.innerHTML = "";
      (list || []).forEach((f) => {
        const li = document.createElement("li");
        li.style.cssText = "padding:.3rem .4rem;cursor:pointer;border-bottom:1px solid var(--border)";
        li.textContent = (f.title || f.fragment_key) + (f.has_draft ? " ✎draft" : "");
        li.onclick = () => this.select(f);
        ul.appendChild(li);
      });
    },
    select(f) {
      this._key = f.fragment_key;
      q("cfg-frag-key").textContent = f.fragment_key;
      q("cfg-frag-editor").value = f.content || "";
      q("cfg-msg").textContent = "";
    },
    async saveDraft() {
      if (!this._key) { q("cfg-msg").textContent = "先選一個 fragment"; return; }
      const r = await api().save_prompt_draft(this._key, q("cfg-frag-editor").value);
      q("cfg-msg").textContent = r && r.ok
        ? "已存草稿 v" + r.version + "（未影響進行中遊戲，按「啟用草稿」才生效）"
        : "存草稿失敗：" + ((r && r.error) || "");
      this.loadFragments();
    },
    async preview() {
      if (!this._key) { q("cfg-msg").textContent = "先選一個 fragment"; return; }
      const r = await api().preview_prompt("story", null, this._key, q("cfg-frag-editor").value);
      if (r && r.ok) {
        q("cfg-preview").textContent = r.compiled_prompt;
        q("cfg-preview-hash").textContent = "preview hash: " + r.prompt_hash
          + (r.llm_called ? "" : " · 零 LLM");
        if (r.missing_required && r.missing_required.length)
          q("cfg-msg").textContent = "⚠ 缺必填變數：" + r.missing_required.join(", ");
      } else {
        q("cfg-msg").textContent = "預覽失敗：" + ((r && r.error) || "");
      }
    },
    async activate() {
      if (!this._key) { q("cfg-msg").textContent = "先選一個 fragment"; return; }
      const r = await api().activate_prompt_draft(this._key);
      q("cfg-msg").textContent = r && r.ok
        ? "✓ 已啟用 v" + r.activated_version + "（現在起 active prompt 生效）"
        : "啟用失敗：" + ((r && r.error) || "");
      this.loadFragments();
      const ov = await api().config_overview();
      if (ov && ov.story) q("cfg-active-hash").textContent = "active hash: " + ov.story.prompt_hash;
    },
    async loadFlags() {
      const flags = await api().list_feature_flags();
      const ul = q("cfg-flags"); ul.innerHTML = "";
      (flags || []).forEach((f) => {
        const li = document.createElement("li");
        const cb = document.createElement("input");
        cb.type = "checkbox"; cb.checked = !!f.value; cb.id = "flag-" + f.name;
        cb.onchange = async () => { await api().set_feature_flag(f.name, cb.checked); };
        const lbl = document.createElement("label");
        lbl.htmlFor = cb.id; lbl.textContent = " " + f.name; lbl.style.cursor = "pointer";
        li.appendChild(cb); li.appendChild(lbl); ul.appendChild(li);
      });
    },
  };
  window.ConfigUI = ConfigUI;

  function bootConfig() {
    const d = q("cfg-btn-draft"), p = q("cfg-btn-preview"), a = q("cfg-btn-activate");
    if (d) d.onclick = () => ConfigUI.saveDraft();
    if (p) p.onclick = () => ConfigUI.preview();
    if (a) a.onclick = () => ConfigUI.activate();
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
