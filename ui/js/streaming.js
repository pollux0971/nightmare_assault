// streaming.js — NA.* 前端串流接收與渲染（U17，前端承重牆）
// 後端逐 token 推 NA.appendToken；前端控速吐字、關鍵詞血紅、決策淡入。
// 前端「不解析」分隔符或 JSON——那全在後端 StreamParser 完成。
(function () {
  const KEY_RE = /(死|血|屍|尖叫|黑暗|消失|背叛|逃|別動|不要|救命|怪|噬|腐|冷)/;
  const q = (id) => document.getElementById(id);

  // 控速吐字佇列（避免一次 LLM token 過長卡頓）
  let queue = [];
  let draining = false;
  function enqueue(text) {
    for (const ch of text) queue.push(ch);
    drain();
  }
  function drain() {
    if (draining) return;
    draining = true;
    const narr = q("narrative");
    function tick() {
      if (queue.length === 0) { draining = false; return; }
      const ch = queue.shift();
      const span = document.createElement("span");
      span.className = "tok" + (KEY_RE.test(ch) ? " blood" : "");
      span.textContent = ch;
      narr.appendChild(span);
      narr.scrollTop = narr.scrollHeight;
      // 句末停頓更久，營造呼吸
      const pause = /[。！？…\n]/.test(ch) ? 220 : (KEY_RE.test(ch) ? 70 : 26);
      setTimeout(tick, pause);
    }
    tick();
  }

  function newBeatBlock() {
    const narr = q("narrative");
    const hr = document.createElement("div");
    hr.style.cssText = "height:1.2rem";
    narr.appendChild(hr);
  }

  function switchToGameIfLoading() {
    if (NA._view === "loading") Views.show("game");
  }

  // ── 等待動畫（注入 #narrative，story 還沒吐字前一直跑）──
  const WAIT_LINES = ["夢魘正在成形…", "黑暗在重新排列…", "它聽見了你的選擇…",
    "空氣變得黏稠…", "某個東西正在靠近…", "別回頭。"];
  let waitTimer = null, waitIdx = 0, waitingShown = false;
  function showWaiting() {
    const narr = q("narrative");
    if (!document.getElementById("wait-core")) {
      const w = document.createElement("div");
      w.id = "wait-core"; w.className = "wait-core";
      w.innerHTML = '<div class="wait-dots"><span></span><span></span><span></span></div>'
        + '<p id="wait-line" class="wait-line"></p>';
      narr.appendChild(w);
    }
    waitingShown = true;
    const set = () => { const el = document.getElementById("wait-line");
      if (el) { el.style.opacity = "0";
        setTimeout(() => { const e2 = document.getElementById("wait-line");
          if (e2) { e2.textContent = WAIT_LINES[waitIdx++ % WAIT_LINES.length]; e2.style.opacity = "1"; } }, 300); } };
    const el = document.getElementById("wait-line");
    if (el) { el.textContent = WAIT_LINES[waitIdx++ % WAIT_LINES.length]; el.style.opacity = "1"; }
    clearInterval(waitTimer); waitTimer = setInterval(set, 2800);
  }
  function hideWaiting() {
    clearInterval(waitTimer); waitingShown = false;
    const w = document.getElementById("wait-core"); if (w) w.remove();
  }

  // ── HUD：生成管線進度 ──
  function hudReset() { q("hud-steps").innerHTML = ""; q("hud").classList.remove("hidden"); }
  function hudMarkDone(ul) {
    ul.querySelectorAll("li.active").forEach((li) => {
      li.className = "done"; const m = li.querySelector(".mark"); if (m) m.textContent = "✓";
    });
  }
  function hudProgress(label) {
    const ul = q("hud-steps");
    hudMarkDone(ul);
    const li = document.createElement("li"); li.className = "active";
    li.innerHTML = '<span class="mark">●</span><span>' + label + "</span>";
    ul.appendChild(li);
  }
  function hudDone() {
    hudMarkDone(q("hud-steps"));
    setTimeout(() => q("hud").classList.add("hidden"), 1100);
  }

  const NA = {
    _view: "settings",
    _busy: false,

    onOpening(line) {
      switchToGameIfLoading();
      enqueue(line + "\n\n");
    },
    appendToken(tok) {
      if (waitingShown) hideWaiting();
      switchToGameIfLoading();
      enqueue(tok);
    },
    onProgress(p) {
      hudProgress((p && p.label) || "處理中");
    },
    onProgressInfo(info) {
      // kernel 推進事件 + delta 顯示於 HUD
      const ul = q("hud-steps");
      if (!ul || !info) return;
      const li = document.createElement("li");
      li.className = "done";
      const delta = (info.delta || []).join("·");
      li.innerHTML = '<span class="mark">✦</span><span>' + (info.event || "")
        + (delta ? "（" + delta + "）" : "") + "</span>";
      ul.appendChild(li);
    },
    onContinue() {
      const b = q("btn-continue-narration");
      if (b) b.classList.remove("hidden");
    },
    onSkillLimit(info) {
      // UB1：破格技能被封頂 → 顯示侷限提示（淡入後自動消失）
      if (!info || !info.limitation) return;
      const bar = q("skill-limit");
      if (!bar) return;
      bar.textContent = "⚠ 能力受限：" + info.limitation;
      bar.classList.add("show");
      clearTimeout(bar._t);
      bar._t = setTimeout(() => bar.classList.remove("show"), 9000);
    },
    onDecision(dp) {
      // 等吐字佇列排空再呈現決策（敘述吐完才淡入選項）
      const waitDrain = () => {
        if (queue.length > 0 || draining) { setTimeout(waitDrain, 120); return; }
        renderDecision(dp);
      };
      waitDrain();
    },
    onBeatComplete() {
      hudDone();
      newBeatBlock();
    },
    onStatus(st) {
      NA._busy = !!(st && st.busy);
      const loc = q("loc-name"); if (loc && st && st.current_location) loc.textContent = st.current_location;
      const bn = q("beat-no"); if (bn && st) bn.textContent = st.beat_number;
      const sb = q("game-status");
      if (sb) sb.textContent = st && st.busy ? "（夢魘正在成形…）" : "";
      const free = q("free-text"), btn = q("btn-submit");
      if (free) free.disabled = NA._busy;
      if (btn) btn.disabled = NA._busy;
    },
    onError(err) {
      hideWaiting();
      const msg = (err && err.message) ? err.message : "發生錯誤";
      if (NA._view === "loading") {            // 開局失敗 → 回新局畫面提示
        Views.show("new");
        const m = q("new-msg"); if (m) m.textContent = msg;
        return;
      }
      const sb = q("game-status");
      if (sb) sb.textContent = "⚠ " + msg;
    },
    onEnding(info) {
      hideWaiting();
      info = info || {};
      const t = q("ending-title"), b = q("ending-body");
      if (t) t.textContent = info.title || "結局";
      // UB7 masked：純敘述收尾 + 已確認(已發現全文) + 未確認(遮罩標題) + 重玩鉤子
      const parts = [];
      if (info.closing) parts.push(info.closing);
      const rc = info.recap || {}, tr = info.truth || {};
      const total = rc.total_count || 0, found = rc.found_count || 0;
      const ratio = total ? found / total : 0;
      const cont = (x) => (x && typeof x === "object") ? (x.content || "") : (x || "");
      const titl = (x) => (x && typeof x === "object") ? (x.title || "未解的線索") : "未解的線索";
      if (total) {
        parts.push("\n── 復盤 ──");
        parts.push("你在這場夢魘裡發現了 " + found + "/" + total + " 個真相碎片。");
        if (rc.discovered && rc.discovered.length) {
          parts.push("已確認：");
          rc.discovered.forEach((d) => parts.push("・" + cont(d)));
        }
        if (rc.missed && rc.missed.length) {
          parts.push("未確認：");
          rc.missed.forEach((m) => parts.push("・" + titl(m) + "：？？？"));
        }
      }
      if (tr.what_really_happened && total && ratio >= 0.6) {
        parts.push("\n── 真相 ──");
        parts.push("真正發生的事：" + tr.what_really_happened);
        if (tr.the_threat_is) parts.push("真正的威脅：" + tr.the_threat_is);
        if (tr.deadly_rule) parts.push("那條致命規則：" + tr.deadly_rule);
      }
      if (rc.missed && rc.missed.length) parts.push("\n有些答案，你還沒走到它面前。");
      if (b) b.textContent = parts.join("\n") || q("narrative").textContent.slice(-600);
      Views.show("ending");
    },
  };
  window.NA = NA;

  function clearNarrative() { q("narrative").innerHTML = ""; queue = []; draining = false; }

  function renderDecision(dp) {
    const recap = q("recap");
    recap.textContent = dp.situation_recap || "";
    recap.classList.remove("show"); void recap.offsetWidth; recap.classList.add("show");
    const opts = q("options");
    opts.innerHTML = "";
    (dp.suggested_options || []).forEach((o, i) => {
      const btn = document.createElement("button");
      btn.className = "opt" + (o.tone === "aggressive" || o.tone === "bold" ? " danger" : "");
      btn.textContent = o.text;
      btn.style.animationDelay = (i * 0.28 + 0.3) + "s";
      btn.onclick = () => window.App && App.submit(o.text, "option");
      opts.appendChild(btn);
    });
    q("btn-continue-narration").classList.add("hidden");
    const free = q("free-text");
    if (free) { free.disabled = false; free.focus(); }
  }

  window.Streaming = { clearNarrative, showWaiting, hideWaiting, hudReset };
})();
