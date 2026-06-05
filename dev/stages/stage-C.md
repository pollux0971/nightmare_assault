# 階段 C · MVP-C（聊天室 + 離場命運）

- **進入條件**：MVP-A + 階段 S/P/B/N done。
- **目標**：接通兩個既有但未接迴圈的 agent，讓世界真正活起來、離場有後果——玩家能跟在場 NPC 多輪對話（npc-chat），離場 NPC 在看不見處繼續自己的命運（offstage-fate）。
- **整合驗收**：聊天多輪不暴雷（無 real_bible）/ 退出濃縮進 story context / 離場四種命運機率可控且只寫 npc/scene / 命運對玩家隱藏直到重逢或搜屍才揭曉。

## 套用順序

```text
MC1 → MC2 → MC3   （聊天室：agent → 濃縮 → UI）
MC4 → MC5         （離場命運：agent → 隱藏/重逢揭曉）
```

## 本階段工單

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `MC1` | D | NPC-Chat agent + chat_logs 持久化（認知卡投影、無 real_bible） | U07, U05 |
| `MC2` | D | 聊天退出濃縮（3–4 句進 story hot context） | MC1, U14 |
| `MC3` | E | 聊天室 UI + API | MC1, U18 |
| `MC4` | D | Offstage-Fate agent + 命運 tick（程式碼擲骰 + LLM 寫血肉） | U03, U04, U10 |
| `MC5` | D | 離場命運隱藏 + 重逢揭曉 | MC4, U10 |

> 紀律：npc-chat 結構性防暴雷（同 story，認知卡無 real_bible）；offstage-fate 只寫 npc/scene（權限邊界，碰不到 secret_core）；
> 離場命運隱藏直到重逢/搜屍才過揭露閘門（路線 A，不主動劇透）；各 agent 失敗皆 graceful（不崩主迴圈）。
> canonical：`skills/npc-chat/SKILL.md`、`skills/offstage-fate/SKILL.md`、00 §六 B5/B6/離場雙軸。契約見 dev/CONTRACTS.md §十三。
