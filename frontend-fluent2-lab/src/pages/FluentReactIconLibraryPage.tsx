import * as React from 'react';
import {
  Accordion,
  AccordionHeader,
  AccordionItem,
  AccordionPanel,
  Badge,
  Body1,
  Caption1,
  Field,
  MessageBar,
  SearchBox,
  Tag,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import * as FluentIcons from '@fluentui/react-icons';
import { LabBadgeRow, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

type IconVariant = 'Filled' | 'Regular';

type IconEntry = {
  id: string;
  group: string;
  previewName: string;
  sizes: string[];
  variants: IconVariant[];
};

type IconComponent = React.ComponentType<{ className?: string }>;

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  hero: {
    display: 'grid',
    gap: '14px',
    padding: '20px 22px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.12) 0%, rgba(15,108,189,0.03) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '16px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  heroMeta: {
    display: 'grid',
    gap: '10px',
  },
  row: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  stack: {
    display: 'grid',
    gap: '12px',
  },
  toolbar: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) auto',
    gap: '12px',
    alignItems: 'end',
    '@media (max-width: 920px)': {
      gridTemplateColumns: '1fr',
    },
  },
  groupList: {
    display: 'grid',
    gap: '12px',
  },
  accordionItem: {
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    overflow: 'hidden',
  },
  groupHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    flexWrap: 'wrap',
  },
  iconGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(148px, 1fr))',
    gap: '12px',
  },
  iconButton: {
    width: '100%',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    paddingTop: '14px',
    paddingRight: '12px',
    paddingBottom: '14px',
    paddingLeft: '12px',
    display: 'grid',
    gap: '10px',
    textAlign: 'left',
    cursor: 'pointer',
    transitionDuration: '150ms',
    transitionProperty: 'transform, border-color, background-color',
    ':hover': {
      transform: 'translateY(-1px)',
      border: `1px solid ${tokens.colorBrandStroke1}`,
      backgroundColor: tokens.colorNeutralBackground2,
    },
    ':focus-visible': {
      outlineStyle: 'solid',
      outlineWidth: '2px',
      outlineColor: tokens.colorStrokeFocus2,
      outlineOffset: '2px',
    },
  },
  iconPreview: {
    width: '100%',
    height: '68px',
    display: 'grid',
    placeItems: 'center',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    color: tokens.colorBrandForeground1,
  },
  iconGlyph: {
    fontSize: '28px',
  },
  iconMeta: {
    display: 'grid',
    gap: '6px',
  },
  iconId: {
    wordBreak: 'break-word',
  },
  tags: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '6px',
  },
  resultGrid: {
    display: 'grid',
    gap: '12px',
  },
  helperText: {
    color: tokens.colorNeutralForeground3,
  },
});

function resolveGroup(id: string) {
  const first = id.charAt(0).toUpperCase();
  return /[A-Z]/.test(first) ? first : '#';
}

function choosePreviewName(id: string, availableNames: Set<string>) {
  const candidates = [
    `${id}Regular`,
    `${id}Filled`,
    `${id}20Regular`,
    `${id}20Filled`,
    `${id}24Regular`,
    `${id}24Filled`,
    `${id}16Regular`,
    `${id}16Filled`,
  ];

  return candidates.find(name => availableNames.has(name)) ?? candidates[0];
}

const iconEntries = (() => {
  const rawNames = Object.keys(FluentIcons).filter(name => /(?:Filled|Regular)$/.test(name));
  const availableNames = new Set(rawNames);
  const entryMap = new Map<string, { sizes: Set<string>; variants: Set<IconVariant> }>();

  for (const name of rawNames) {
    const matched = name.match(/^(.+?)(\d+)?(Filled|Regular)$/);
    if (!matched) {
      continue;
    }

    const [, id, size, variant] = matched;
    const current = entryMap.get(id) ?? { sizes: new Set<string>(), variants: new Set<IconVariant>() };
    if (size) {
      current.sizes.add(size);
    }
    current.variants.add(variant as IconVariant);
    entryMap.set(id, current);
  }

  return Array.from(entryMap.entries())
    .map(([id, current]) => ({
      id,
      group: resolveGroup(id),
      previewName: choosePreviewName(id, availableNames),
      sizes: Array.from(current.sizes).sort((left, right) => Number(left) - Number(right)),
      variants: Array.from(current.variants).sort(),
    }))
    .sort((left, right) => left.id.localeCompare(right.id, 'en'));
})();

const groupedEntries = iconEntries.reduce<Record<string, IconEntry[]>>((accumulator, entry) => {
  const current = accumulator[entry.group] ?? [];
  current.push(entry);
  accumulator[entry.group] = current;
  return accumulator;
}, {});

const groupKeys = Object.keys(groupedEntries).sort((left, right) => {
  if (left === '#') {
    return 1;
  }
  if (right === '#') {
    return -1;
  }
  return left.localeCompare(right, 'en');
});

function IconCard({
  entry,
  onCopy,
}: {
  entry: IconEntry;
  onCopy: (id: string) => void;
}) {
  const styles = useStyles();
  const Icon = FluentIcons[entry.previewName as keyof typeof FluentIcons] as IconComponent | undefined;

  return (
    <button className={styles.iconButton} type="button" onClick={() => onCopy(entry.id)} title={`复制 ${entry.id}`}>
      <div className={styles.iconPreview}>{Icon ? <Icon className={styles.iconGlyph} /> : null}</div>
      <div className={styles.iconMeta}>
        <Body1 className={styles.iconId}>{entry.id}</Body1>
        <Caption1 className={styles.helperText}>点击复制基础图标 ID</Caption1>
      </div>
      <div className={styles.tags}>
        {entry.variants.map(variant => (
          <Tag key={variant}>{variant}</Tag>
        ))}
        {entry.sizes.length ? <Tag appearance="outline">尺寸 {entry.sizes.join(' / ')}</Tag> : null}
      </div>
    </button>
  );
}

export function FluentReactIconLibraryPage() {
  const styles = useStyles();
  const [query, setQuery] = React.useState('');
  const [copiedId, setCopiedId] = React.useState<string | null>(null);
  const deferredQuery = React.useDeferredValue(query.trim().toLowerCase());

  const filteredEntries = React.useMemo(() => {
    if (!deferredQuery) {
      return iconEntries;
    }

    return iconEntries.filter(
      entry =>
        entry.id.toLowerCase().includes(deferredQuery) ||
        entry.variants.some(variant => variant.toLowerCase().includes(deferredQuery)) ||
        entry.sizes.some(size => size.includes(deferredQuery)),
    );
  }, [deferredQuery]);

  const filteredGroups = React.useMemo(() => {
    return groupKeys
      .map(group => ({
        group,
        items: filteredEntries.filter(entry => entry.group === group),
      }))
      .filter(section => section.items.length > 0);
  }, [filteredEntries]);

  const dualVariantCount = React.useMemo(
    () => filteredEntries.filter(entry => entry.variants.length === 2).length,
    [filteredEntries],
  );

  const handleCopy = React.useCallback(async (id: string) => {
    try {
      await navigator.clipboard.writeText(id);
    } catch {
      const temporary = document.createElement('textarea');
      temporary.value = id;
      temporary.style.position = 'fixed';
      temporary.style.opacity = '0';
      document.body.appendChild(temporary);
      temporary.select();
      document.execCommand('copy');
      document.body.removeChild(temporary);
    }

    setCopiedId(id);
  }, []);

  return (
    <div className={styles.page}>
      <header className={styles.hero}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Fluent 2 React
          </Badge>
          <Badge appearance="tint">基础控件集</Badge>
          <Badge appearance="outline">图标总览</Badge>
        </LabBadgeRow>

        <div className={styles.heroRow}>
          <div className={styles.heroMeta}>
            <Title3>React 图标总览</Title3>
            <Body1>
              展示 <code>@fluentui/react-icons</code> 的基础图标 ID。页面默认折叠尺寸变体，点击卡片即可复制基础图标
              ID，方便你在组件页或业务页里快速拼出 <code>Regular / Filled</code> 导出名。
            </Body1>
            <div className={styles.row}>
              <Tag>点击卡片复制 ID</Tag>
              <Tag>按首字母分组</Tag>
              <Tag>折叠尺寸变体</Tag>
            </div>
          </div>

          <LabStatGrid
            items={[
              { label: '基础图标 ID', value: String(iconEntries.length), tone: 'brand' },
              { label: '分组数量', value: String(groupKeys.length), tone: 'success' },
              { label: '双变体图标', value: String(dualVariantCount), tone: 'warning' },
            ]}
          />
        </div>
      </header>

      <LabSurfaceCard>
        <div className={styles.stack}>
          <LabSectionTitle
            title="检索与复制"
            description="输入图标关键字即可过滤。默认复制基础 ID，例如 AccessTime，再按需要拼出 AccessTimeRegular 或 AccessTimeFilled。"
          />
          <div className={styles.toolbar}>
            <Field label="搜索图标">
              <SearchBox
                placeholder="例如 Access、Calendar、Shield"
                value={query}
                onChange={(_, data) => setQuery(data.value)}
              />
            </Field>
            <div className={styles.row}>
              <Badge appearance="filled">{filteredEntries.length} 个结果</Badge>
              {copiedId ? (
                <Badge appearance="tint" color="success">
                  最近复制：{copiedId}
                </Badge>
              ) : null}
            </div>
          </div>
          <MessageBar>
            图标页优先服务于组件试验和业务实现，不建议把图标名字直接暴露给业务文案。
          </MessageBar>
        </div>
      </LabSurfaceCard>

      {query ? (
        <LabSurfaceCard>
          <LabSectionTitle title="搜索结果" description="搜索状态下直接平铺结果，便于连续复制多个图标 ID。" />
          <div className={styles.resultGrid}>
            <div className={styles.iconGrid}>
              {filteredEntries.map(entry => (
                <IconCard key={entry.id} entry={entry} onCopy={handleCopy} />
              ))}
            </div>
            {!filteredEntries.length ? <MessageBar>没有命中图标，请换一个关键词。</MessageBar> : null}
          </div>
        </LabSurfaceCard>
      ) : (
        <LabSurfaceCard>
          <LabSectionTitle
            title="按首字母浏览"
            description="保持分组浏览，避免一开始就渲染全部图标卡片。展开需要的字母区即可查看并复制。"
          />
          <div className={styles.groupList}>
            <Accordion collapsible multiple defaultOpenItems={['A', 'C']}>
              {filteredGroups.map(section => (
                <AccordionItem key={section.group} value={section.group} className={styles.accordionItem}>
                  <AccordionHeader>
                    <div className={styles.groupHeader}>
                      <Body1>{section.group} 组</Body1>
                      <Badge appearance="outline">{section.items.length} 个图标</Badge>
                    </div>
                  </AccordionHeader>
                  <AccordionPanel>
                    <div className={styles.iconGrid}>
                      {section.items.map(entry => (
                        <IconCard key={entry.id} entry={entry} onCopy={handleCopy} />
                      ))}
                    </div>
                  </AccordionPanel>
                </AccordionItem>
              ))}
            </Accordion>
          </div>
        </LabSurfaceCard>
      )}

      <LabSurfaceCard subtle>
        <LabSectionTitle
          title="使用建议"
          description="复制的是基础图标 ID，不包含尺寸后缀，适合在代码里统一派生变体。"
        />
        <div className={styles.stack}>
          <Caption1 className={styles.helperText}>如果某个图标只存在尺寸变体，页面仍会展示该基础 ID，并在标签里标出可用尺寸。</Caption1>
          <Caption1 className={styles.helperText}>建议业务页保持低噪图标密度，把图标作为节奏提示，而不是主要装饰。</Caption1>
        </div>
      </LabSurfaceCard>
    </div>
  );
}
