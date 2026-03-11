#!/usr/bin/env python3
"""
Art Design Pro 组件列表工具

列出所有可用的 Art Design Pro 组件
"""

import csv
from pathlib import Path

COMPONENTS_DB = Path(__file__).parent.parent / "data" / "components.csv"


def main():
    """列出所有组件"""

    with open(COMPONENTS_DB, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)

        # 按分类组织
        categories = {}
        for row in reader:
            cat = row["category"]
            if cat not in categories:
                categories[cat] = []
            categories[cat].append(row)

    print("\n" + "=" * 70)
    print("📦 Art Design Pro 组件库完整列表".center(70))
    print("=" * 70)

    # 定义分类显示顺序和中文标题
    category_titles = {
        "tables": ("📊 表格与数据展示", "tables"),
        "forms": ("📝 表单与输入", "forms"),
        "cards": ("🎴 卡片组件", "cards"),
        "charts": ("📈 图表组件", "charts"),
        "layouts": ("🎨 布局与导航", "layouts"),
        "media": ("🎬 媒体组件", "media"),
        "banners": ("🎪 横幅组件", "banners"),
        "text-effect": ("✨ 文本特效", "text-effect"),
        "base": ("🔧 基础组件", "base"),
        "widget": ("🎯 小部件", "widget"),
        "others": ("🔌 其他组件", "others"),
    }

    # 按定义顺序输出
    for cat_key, (title, _) in category_titles.items():
        if cat_key not in categories:
            continue

        print(f"\n{title}")
        print("-" * 70)

        for comp in categories[cat_key]:
            print(
                f"  • {comp['component']:<30} {comp['name_cn']:<20} {comp['name_en']}"
            )

    print("\n" + "=" * 70)
    print(f"📊 共计 {sum(len(comps) for comps in categories.values())} 个组件")
    print("=" * 70 + "\n")

    print("💡 使用方法:")
    print("  python3 scripts/search.py \"关键词\"    # 搜索组件")
    print("  python3 scripts/search.py --list        # 查看所有分类\n")


if __name__ == "__main__":
    main()
