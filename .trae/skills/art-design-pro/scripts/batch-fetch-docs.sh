#!/bin/bash
# Art Design Pro 官方文档批量下载脚本
# 使用 Claude Code 的 MCP 工具获取所有文档并保存到本地

DOCS_DIR="/home/aijiaye/hcg/.claude/skills/art-design-pro/docs"
mkdir -p "$DOCS_DIR"

echo "====================================="
echo "Art Design Pro 文档批量下载工具"
echo "====================================="
echo ""

# 文档列表（URL -> 文件名）
declare -A DOCS=(
  # 开始
  ["https://www.artd.pro/docs/zh/guide/introduce.html"]="01-introduce.md"
  ["https://www.artd.pro/docs/zh/guide/quick-start.html"]="02-quick-start.md"
  ["https://www.artd.pro/docs/zh/guide/lite-version.html"]="03-lite-version.md"
  ["https://www.artd.pro/docs/zh/guide/must-read.html"]="04-must-read.md"
  ["https://www.artd.pro/docs/zh/guide/update.html"]="05-update.md"

  # 基础
  ["https://www.artd.pro/docs/zh/guide/essentials/project-introduce.html"]="06-project-introduce.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/element-plus.html"]="07-element-plus.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/route.html"]="08-route.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/settings.html"]="09-settings.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/theme.html"]="10-theme.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/icon.html"]="11-icon.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/env-variables.html"]="12-env-variables.md"
  ["https://www.artd.pro/docs/zh/guide/essentials/build.html"]="13-build.md"

  # 深入
  ["https://www.artd.pro/docs/zh/guide/in-depth/locale.html"]="14-locale.md"
  ["https://www.artd.pro/docs/zh/guide/in-depth/permission.html"]="15-permission.md"
  ["https://www.artd.pro/docs/zh/guide/in-depth/script.html"]="16-script.md"

  # Hooks
  ["https://www.artd.pro/docs/zh/guide/hooks/use-table.html"]="17-use-table.md"

  # 组件
  ["https://www.artd.pro/docs/zh/guide/components/art-search-bar.html"]="18-art-search-bar.md"

  # 工程化
  ["https://www.artd.pro/docs/zh/guide/project/standard.html"]="19-standard.md"
)

echo "准备下载 ${#DOCS[@]} 个文档页面..."
echo ""

# 由于这个脚本需要 MCP 工具支持，我们生成一个命令列表
# 让 Claude Code 可以逐个执行

COMMAND_FILE="$DOCS_DIR/fetch-commands.txt"
echo "# Art Design Pro 文档获取命令" > "$COMMAND_FILE"
echo "# 在 Claude Code 中逐个执行这些命令" >> "$COMMAND_FILE"
echo "" >> "$COMMAND_FILE"

for url in "${!DOCS[@]}"; do
  filename="${DOCS[$url]}"
  echo "mcp__web-reader__webReader(url=\"$url\")" >> "$COMMAND_FILE"
done

echo "✅ 命令文件已生成: $COMMAND_FILE"
echo ""
echo "⚠️  注意：由于需要 MCP 工具支持，请在 Claude Code 中查看命令文件"
echo "   并逐个执行 mcp__web-reader__webReader 命令"
echo ""
echo "📋 文档列表："
count=0
for url in "${!DOCS[@]}"; do
  count=$((count + 1))
  filename="${DOCS[$url]}"
  echo "  $count. $filename"
  echo "     URL: $url"
done

echo ""
echo "====================================="
echo "📊 总计: ${#DOCS[@]} 个文档"
echo "====================================="
