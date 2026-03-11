#!/usr/bin/env python3
"""
Art Design Pro 组件搜索工具

用法：
  python3 search.py "表格"           # 搜索中文关键词
  python3 search.py "table"          # 搜索英文关键词
  python3 search.py "form" --category forms  # 按分类搜索
"""

import csv
import sys
import argparse
from pathlib import Path

# 组件数据库路径
COMPONENTS_DB = Path(__file__).parent.parent / "data" / "components.csv"


def load_components():
    """加载所有组件"""
    components = []
    with open(COMPONENTS_DB, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            components.append(row)
    return components


def search_components(keyword: str, category: str = None):
    """搜索组件"""
    components = load_components()
    keyword = keyword.lower()

    results = []

    for comp in components:
        # 如果指定了分类，先过滤
        if category and comp["category"] != category:
            continue

        # 搜索范围：中文名称、英文名称、描述、常见用法
        search_fields = [
            comp.get("name_cn") or "",
            comp.get("name_en") or "",
            comp.get("description") or "",
            comp.get("component") or "",
            comp.get("common_usage") or "",
        ]

        # 任何字段匹配即命中
        if any(keyword in field.lower() for field in search_fields):
            results.append(comp)

    return results


def format_component(comp: dict):
    """格式化单个组件信息"""
    return f"""
┌─ {comp['component']} ({comp['name_cn']})
│
│  📝 {comp['description']}
│
│  📦 导入: {comp['import_path']}
│
│  🔧 Props: {comp['props']}
│  🎯 Slots: {comp['slots']}
│
│  💡 常见场景: {comp['common_usage']}
└──────────────────────────────────────────────"""


def print_results(results: list):
    """打印搜索结果"""
    if not results:
        print("❌ 未找到匹配的组件")
        print("\n💡 提示：尝试使用更通用的关键词")
        print("   例如：'表格'、'表单'、'图表'、'卡片'、'布局'")
        return

    print(f"\n✅ 找到 {len(results)} 个组件:\n")

    for i, comp in enumerate(results, 1):
        print(f"{i}. {comp['component']} - {comp['name_cn']}")
        print(f"   {comp['description']}")
        print(f"   用途: {comp['common_usage']}")
        print()

    # 如果结果超过5个，询问是否查看详情
    if len(results) > 5:
        print("💡 使用 --detail 参数查看组件详细信息\n")
    else:
        print("\n详细信息:")
        print("=" * 60)
        for comp in results:
            print(format_component(comp))
            print()


def print_detailed_results(results: list):
    """打印详细搜索结果"""
    if not results:
        print("❌ 未找到匹配的组件")
        return

    print(f"\n✅ 找到 {len(results)} 个组件:\n")
    print("=" * 60)

    for i, comp in enumerate(results, 1):
        print(f"\n【{i}】{comp['component']} - {comp['name_cn']}")
        print(format_component(comp))


def list_all_categories():
    """列出所有分类"""
    components = load_components()
    categories = sorted(set(c["category"] for c in components))

    print("\n📁 组件分类:\n")
    for cat in categories:
        count = sum(1 for c in components if c["category"] == cat)
        print(f"  • {cat}: {count} 个组件")

    print("\n💡 使用 --category <分类名> 搜索特定分类的组件")


def main():
    parser = argparse.ArgumentParser(
        description="Art Design Pro 组件搜索工具",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例：
  python3 search.py "表格"           # 搜索中文关键词
  python3 search.py "table"          # 搜索英文关键词
  python3 search.py "form" --category forms  # 按分类搜索
  python3 search.py --list           # 列出所有分类
  python3 search.py "chart" --detail # 查看详细信息
        """
    )

    parser.add_argument("keyword", nargs="?", help="搜索关键词")
    parser.add_argument("--category", "-c", help="按分类过滤 (tables, forms, cards, charts, layouts, etc.)")
    parser.add_argument("--detail", "-d", action="store_true", help="显示详细信息")
    parser.add_argument("--list", "-l", action="store_true", help="列出所有分类")

    args = parser.parse_args()

    # 列出所有分类
    if args.list:
        list_all_categories()
        return

    # 必须提供关键词
    if not args.keyword:
        parser.print_help()
        return

    # 搜索组件
    results = search_components(args.keyword, args.category)

    # 打印结果
    if args.detail:
        print_detailed_results(results)
    else:
        print_results(results)


if __name__ == "__main__":
    main()
