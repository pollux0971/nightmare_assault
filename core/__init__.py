"""Nightmare Assault — 後端 core 套件。

分層：state（Blackboard/WorldBible/…）· memory（三層記憶/快照）· llm（OpenRouter client）
· agents（SkillCaller 與各 agent 封裝）· persistence（SQLite/JSON）。詳見 nightmare-assault-design/01。
"""
__version__ = "0.0.0"
__all__: list[str] = []
