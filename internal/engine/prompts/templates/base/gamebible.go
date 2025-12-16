package base

// TemplateBaseBible contains the core rules for story generation (T_BASE_BIBLE).
// This is the "game bible" that defines all fundamental mechanics.
const TemplateBaseBible = `# Game Bible v1.0

## Core Mechanics
- HP: Physical health (0-100). Reaches 0 = death.
- SAN: Sanity (0-100). Affects perception and choices.
  - 80-100: Clear-headed
  - 50-79: Anxious, minor hallucinations
  - 20-49: Panicked, reality distortion
  - 0-19: Insanity, loss of control

## Narrative Rules
- Every story beat ends with 2-3 choices (maximum 3)
- Each choice must be concise: maximum 15 Traditional Chinese characters
- DO NOT include consequence hints in parentheses (e.g., "（可能損傷HP）")
- Choices must have clear risk/reward implications
- At least one choice per beat should affect HP or SAN
- Maintain Lovecraftian cosmic horror atmosphere
- Use environmental storytelling over exposition

## Hidden Seeds
- Plant 2-3 subtle clues in opening (e.g., locked door, strange symbol)
- Seeds trigger callbacks 3-5 beats later
- Difficulty affects seed impact severity

## Writing Style
- Second person narrative ("You walk into...")
- Present tense for immediacy
- Short paragraphs for tension
- Use sensory details (sounds, smells, textures)
- Avoid explicit gore unless 18+ mode enabled

## Hidden Rules (潛規則)
- The game has hidden rules that the player must discover
- Breaking rules causes HP/SAN damage or instant death
- Rules are NEVER explicitly stated to the player
- Instead, plant subtle clues in the narrative
- Player should be able to deduce rules from context and hints
- Rule types: Scenario, Time, Behavior, Object, Status

## NPC Teammates (隊友系統)
- Stories may include 1-3 AI-generated teammate characters
- Each teammate has distinct personality, background, and skills
- Teammates are introduced through SHOW DON'T TELL:
  * Show personality through actions, dialogue, and possessions
  * NEVER directly list personality traits (e.g., "李明是個理性的人")
  * Good example: "李明推了推眼鏡，掏出筆記本開始記錄牆上的符號"
- Six archetype templates:
  * Victim (受害者型): Easily panics, needs protection, emotional
  * Unreliable (不可靠型): Hides secrets, acts strange, suspicious
  * Logic (理性型): Analytical, calm, provides deduction
  * Intuition (直覺型): Senses danger, provides warnings, perceptive
  * Informer (情報型): Knows background lore, provides clues
  * Possessed (被附身型): Influenced by evil, may betray
- Maintain character consistency using established traits and speech patterns
- Teammates can die, become injured, or change based on story events
`
