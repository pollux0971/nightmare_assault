#!/bin/bash
# download-audio.sh - Nightmare Assault 音频下载脚本
# 下载 Kevin MacLeod 免费音乐（CC BY 4.0）

set -e

BASE_DIR="${HOME}/.nightmare/audio"
BGM_DIR="${BASE_DIR}/bgm"
SFX_DIR="${BASE_DIR}/sfx"

echo "🎵 Nightmare Assault 音频下载器"
echo "================================"
echo ""

# 检查 wget
if ! command -v wget &> /dev/null; then
    echo "❌ 需要安装 wget"
    echo "   Ubuntu/Debian: sudo apt install wget"
    echo "   macOS: brew install wget"
    exit 1
fi

# 创建目录
echo "📁 创建目录结构..."
mkdir -p "${BGM_DIR}"
mkdir -p "${SFX_DIR}"
echo "   ✅ ${BGM_DIR}"
echo "   ✅ ${SFX_DIR}"
echo ""

echo "📥 开始下载 BGM（Kevin MacLeod - CC BY 4.0）..."
echo ""

# 下载函数
download_bgm() {
    local url="$1"
    local filename="$2"
    local description="$3"

    echo -n "  下载 ${description}... "
    if wget -q --show-progress -O "${BGM_DIR}/${filename}" "${url}" 2>&1 | grep -v "^$"; then
        echo "✅"
    else
        if [ -f "${BGM_DIR}/${filename}" ] && [ -s "${BGM_DIR}/${filename}" ]; then
            echo "✅"
        else
            echo "❌ 失败"
            rm -f "${BGM_DIR}/${filename}"
            return 1
        fi
    fi
}

# 1. 探索场景 (ambient_exploration.mp3)
# 使用 "Nerves" - 紧张的环境音乐
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Nerves.mp3" \
    "ambient_exploration.mp3" \
    "探索场景 (Nerves)"

# 2. 紧张/追逐场景 (tension_chase.mp3)
# 使用 "Run!" - 快节奏追逐音乐
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Run.mp3" \
    "tension_chase.mp3" \
    "追逐场景 (Run!)"

# 3. 安全区场景 (safe_rest.mp3)
# 使用 "Meditation Impromptu 01" - 平静的音乐
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Meditation%20Impromptu%2001.mp3" \
    "safe_rest.mp3" \
    "安全场景 (Meditation)"

# 4. 恐怖揭示场景 (horror_reveal.mp3)
# 使用 "Gathering Darkness" - 恐怖气氛
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Gathering%20Darkness.mp3" \
    "horror_reveal.mp3" \
    "恐怖场景 (Gathering Darkness)"

# 5. 谜题场景 (mystery_puzzle.mp3)
# 使用 "Dream Escape" - 神秘感
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Dream%20Escape.mp3" \
    "mystery_puzzle.mp3" \
    "谜题场景 (Dream Escape)"

# 6. 死亡/结局场景 (ending_death.mp3)
# 使用 "Grave Blow" - 悲伤的音乐
download_bgm \
    "https://incompetech.com/music/royalty-free/mp3-royaltyfree/Grave%20Blow.mp3" \
    "ending_death.mp3" \
    "死亡场景 (Grave Blow)"

echo ""
echo "📥 下载基础 SFX..."
echo ""

# 下载一些基础音效
mkdir -p "${SFX_DIR}"

# 心跳音（游戏已实现心跳系统）
download_sfx() {
    local url="$1"
    local filename="$2"
    local description="$3"

    echo -n "  下载 ${description}... "
    if wget -q -O "${SFX_DIR}/${filename}" "${url}" 2>/dev/null; then
        echo "✅"
    else
        echo "⚠️  (需手动下载)"
        touch "${SFX_DIR}/${filename}.placeholder"
    fi
}

# 注：SFX 文件暂时使用占位符，需要从 Freesound.org 手动下载
echo "  ⚠️  SFX 文件需要从 Freesound.org 手动下载"
echo "      推荐搜索关键词: heartbeat, footstep, door creak, scream"

echo ""
echo "================================"
echo "✅ BGM 下载完成！"
echo ""
echo "📁 BGM 目录: ${BGM_DIR}"
echo "   ├── ambient_exploration.mp3  (探索)"
echo "   ├── tension_chase.mp3        (追逐)"
echo "   ├── safe_rest.mp3            (安全)"
echo "   ├── horror_reveal.mp3        (恐怖)"
echo "   ├── mystery_puzzle.mp3       (谜题)"
echo "   └── ending_death.mp3         (死亡)"
echo ""
echo "📁 SFX 目录: ${SFX_DIR}"
echo ""
echo "🎮 现在可以启动游戏了："
echo "   cd ~/Desktop/nightmare-assault"
echo "   ./nightmare-assault"
echo ""
echo "📋 音乐授权信息："
echo "   所有音乐来自 Kevin MacLeod (incompetech.com)"
echo "   授权: Creative Commons: By Attribution 4.0"
echo "   http://creativecommons.org/licenses/by/4.0/"
echo ""
echo "💡 游戏内控制命令："
echo "   /bgm list    - 查看所有 BGM"
echo "   /bgm on/off  - 开启/关闭 BGM"
echo "   /bgm volume 70 - 设置音量"
echo ""
