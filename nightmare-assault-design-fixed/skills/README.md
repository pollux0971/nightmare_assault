# Skills 索引

每個 agent 的 prompt 外部化為一個 `SKILL.md`，可熱重載、可版本控制。loader 在啟動時載入，prompt 調整不需重編程式碼。

| Skill 檔 | 對應 Agent | 模型分層 | 頻率 | 演算法章節 | Context 規格 |
|----------|-----------|---------|------|-----------|-------------|
| `setup/SKILL.md` | setup-agent | Heavy | 一次性 | 02 §三、五 | 03 setup |
| `orchestrator/SKILL.md` | orchestrator-agent | Light | 每 beat | 02 §三 | 03 orchestrator |
| `story/SKILL.md` | story-agent | Medium（串流） | 每 beat | 02 §一、二、四 | 03 story |
| `warden/SKILL.md` | warden-agent | Light | 每 beat（玩家） | 02 §五、六、十 | 03 warden |
| `npc-chat/SKILL.md` | npc-chat-agent | Light | 隨選 | 02 §（聊天） | 03 npc-chat + 認知卡 |
| `dreaming/SKILL.md` | dreaming-agent | Light | 非同步 | 02 §七 | 03 dreaming |
| `offstage-fate/SKILL.md` | offstage-fate-agent | Light | 非同步（命運觸發） | 02 §十一 | 03 offstage-fate |
| `compactor/SKILL.md` | compactor-agent | Medium | 非同步 | 02 §八 | 03 compactor |

## 呼叫順序（一個標準 beat）

```
玩家輸入
  → warden（判玩家動作）
  → orchestrator（決定揭露哪些碎片）
  → story（串流 beat，停在決策點）
  ⇣ 非同步（玩家讀字時）
  → compactor（若使用率過門檻）
  → dreaming（在場 NPC，若排程到）
  → offstage-fate（離場 NPC，若命運 tick 觸發）
```

## 設計原則（所有 skill 共通）

1. **結構化餵 context，指令極短**——個性/行為用範例（voice_sample）錨定，不用形容詞堆砌。
2. **權限邊界**——dreaming/offstage-fate 碰不到 secret_core；story 讀不到 real_bible。錨點靠權限而非自律保護。
3. **嚴格 JSON 輸出**（除 story/npc-chat 的敘述部分串流）——便於程式碼解析。
4. **熱重載**——改 prompt 不重編程式碼。

## 模型字串（OpenRouter，見 01）

- Heavy：`claude-sonnet-4`（fallback：claude-3.5-sonnet → gemini-flash）
- Medium：`gemini-flash`（fallback：claude-haiku-3.5 → gpt-4.1-mini）
- Light：`gpt-4.1-mini`（fallback：gemini-flash-lite → 本地邏輯）
