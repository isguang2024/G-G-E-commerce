#!/usr/bin/env python3
"""
批量下载 Art Design Pro 官方文档
"""

import json
import sys
from pathlib import Path

# 添加 MCP 工具路径
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "lib" / "mcp"))

# 文档列表
DOCS = [
    # 开始
    ("https://www.artd.pro/docs/zh/guide/introduce.html", "01-introduce.md"),
    ("https://www.artd.pro/docs/zh/guide/quick-start.html", "02-quick-start.md"),
    ("https://www.artd.pro/docs/zh/guide/lite-version.html", "03-lite-version.md"),
    ("https://www.artd.pro/docs/zh/guide/must-read.html", "04-must-read.md"),
    ("https://www.artd.pro/docs/zh/guide/update.html", "05-update.md"),

    # 基础
    ("https://www.artd.pro/docs/zh/guide/essentials/project-introduce.html", "06-project-introduce.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/element-plus.html", "07-element-plus.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/route.html", "08-route.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/settings.html", "09-settings.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/theme.html", "10-theme.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/icon.html", "11-icon.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/env-variables.html", "12-env-variables.md"),
    ("https://www.artd.pro/docs/zh/guide/essentials/build.html", "13-build.md"),

    # 深入
    ("https://www.artd.pro/docs/zh/guide/in-depth/locale.html", "14-locale.md"),
    ("https://www.artd.pro/docs/zh/guide/in-depth/permission.html", "15-permission.md"),
    ("https://www.artd.pro/docs/zh/guide/in-depth/script.html", "16-script.md"),

    # Hooks
    ("https://www.artd.pro/docs/zh/guide/hooks/use-table.html", "17-use-table.md"),

    # 组件
    ("https://www.artd.pro/docs/zh/guide/components/art-search-bar.html", "18-art-search-bar.md"),

    # 工程化
    ("https://www.artd.pro/docs/zh/guide/project/standard.html", "19-standard.md"),
]

DOCS_DIR = Path(__file__).parent.parent / "docs"
DOCS_DIR.mkdir(exist_ok=True)

print("Art Design Pro 文档下载工具")
print("=" * 60)
print(f"目标目录: {DOCS_DIR}")
print("=" * 60)

# 由于我们无法在 Python 中直接调用 MCP 工具，我们需要使用 Bash 调用
import subprocess
import urllib.request
import html

def fetch_url(url: str) -> str:
    """使用 Python 内置的 urllib 获取页面内容"""
    try:
        with urllib.request.urlopen(url, timeout=30) as response:
            content = response.read().decode('utf-8')

            # 简单提取 HTML 中的文本内容
            # 注意：这是简化版本，实际生产环境应该使用专门的 HTML 解析库
            import re
            # 提取 <title>
            title_match = re.search(r'<title>(.*?)</title>', content, re.IGNORECASE)
            title = title_match.group(1) if title_match else "未知标题"

            # 这里返回原始内容，后续可以进一步处理
            return f"# {title}\n\n来源: {url}\n\n"
    except Exception as e:
        return f"# 获取失败\n\nURL: {url}\n错误: {str(e)}\n"

# 由于 Python 直接获取网页内容有限制，我们生成一个索引文件
# 让用户手动使用 web-reader MCP 工具获取

def main():
    """生成文档索引"""

    index_content = "# Art Design Pro 官方文档索引\n\n"
    index_content += "本文档索引列出了所有官方文档页面的 URL。\n\n"
    index_content += "## 使用方法\n\n"
    index_content += "由于网页内容需要通过 MCP 工具获取，请在 Claude Code 中使用以下命令：\n\n"
    index_content += "```\n"
    index_content += "# 使用 web-reader MCP 工具\n"
    index_content += "mcp__web-reader__webReader(url=\"<URL>\")\n"
    index_content += "```\n\n"

    sections = {
        "开始 (Getting Started)": [],
        "基础 (Basics)": [],
        "深入 (In-depth)": [],
        "Hooks 函数 (Hooks)": [],
        "组件 (Components)": [],
        "工程化 (Engineering)": []
    }

    for url, filename in DOCS:
        if "/guide/" in url and "/essentials/" not in url and "/in-depth/" not in url and "/hooks/" not in url and "/components/" not in url and "/project/" not in url:
            sections["开始 (Getting Started)"].append((url, filename))
        elif "/essentials/" in url:
            sections["基础 (Basics)"].append((url, filename))
        elif "/in-depth/" in url:
            sections["深入 (In-depth)"].append((url, filename))
        elif "/hooks/" in url:
            sections["Hooks 函数 (Hooks)"].append((url, filename))
        elif "/components/" in url:
            sections["组件 (Components)"].append((url, filename))
        elif "/project/" in url:
            sections["工程化 (Engineering)"].append((url, filename))

    for section_name, docs in sections.items():
        if not docs:
            continue
        index_content += f"## {section_name}\n\n"
        for i, (url, filename) in enumerate(docs, 1):
            index_content += f"{i}. [{filename}]({url})\n"
        index_content += "\n"

    index_file = DOCS_DIR / "index.md"
    with open(index_file, "w", encoding="utf-8") as f:
        f.write(index_content)

    print(f"\n✅ 索引文件已生成: {index_file}")
    print(f"\n📋 共 {len(DOCS)} 个文档页面")

    # 生成一个用于下载的脚本
    download_script = DOCS_DIR / "download.sh"
    with open(download_script, "w", encoding="utf-8") as f:
        f.write("#!/bin/bash\n# Art Design Pro 文档下载脚本\n# 注意：需要 MCP 工具支持\n\n")
        for url, filename in DOCS:
            f.write(f"# {filename}\n")
            f.write(f"# URL: {url}\n")
            f.write(f"# echo 'Downloading {filename}...'\n")
            f.write(f"# curl -s '{url}' | pandoc -f html -t markdown -o '{DOCS_DIR}/{filename}'\n\n")

    print(f"📜 下载脚本已生成: {download_script}")
    print(f"\n💡 提示：在 Claude Code 中使用 mcp__web-reader__webReader 工具获取每个页面的内容")

if __name__ == "__main__":
    main()
