import * as React from 'react';
import {
  Accordion,
  AccordionHeader,
  AccordionItem,
  AccordionPanel,
  Avatar,
  AvatarGroup,
  Badge,
  Body1Strong,
  Breadcrumb,
  BreadcrumbButton,
  BreadcrumbDivider,
  BreadcrumbItem,
  Button,
  Caption1,
  Card,
  CardHeader,
  Carousel,
  CarouselCard,
  CarouselSlider,
  CarouselViewport,
  Checkbox,
  Combobox,
  Dialog,
  DialogActions,
  DialogBody,
  DialogContent,
  DialogSurface,
  DialogTitle,
  DialogTrigger,
  Divider,
  Drawer,
  DrawerBody,
  DrawerFooter,
  DrawerHeader,
  DrawerHeaderTitle,
  Dropdown,
  Field,
  FluentProvider,
  Image,
  InfoLabel,
  Input,
  Label,
  Link,
  List,
  ListItem,
  Menu,
  MenuItem,
  MenuList,
  MenuPopover,
  MenuTrigger,
  MessageBar,
  Nav,
  NavCategory,
  NavCategoryItem,
  NavItem,
  NavSectionHeader,
  NavSubItem,
  NavSubItemGroup,
  Option,
  Persona,
  Popover,
  PopoverSurface,
  PopoverTrigger,
  ProgressBar,
  Radio,
  RadioGroup,
  Rating,
  SearchBox,
  Select,
  Skeleton,
  Slider,
  SpinButton,
  Spinner,
  Switch,
  Tab,
  TabList,
  Tag,
  TagPicker,
  TagPickerButton,
  TagPickerControl,
  TagPickerGroup,
  TagPickerInput,
  TagPickerList,
  TagPickerOption,
  Text,
  Textarea,
  Toast,
  ToastTitle,
  Toaster,
  Toolbar,
  ToolbarButton,
  Tooltip,
  Tree,
  TreeItem,
  TreeItemLayout,
  makeStyles,
  tokens,
  teamsLightTheme,
  useToastController,
  useTagPickerFilter,
  Title3,
} from '@fluentui/react-components';
import { LabBadgeRow, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

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
  layout: {
    display: 'grid',
    gap: '18px',
  },
  coverage: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 980px)': {
      gridTemplateColumns: '1fr',
    },
  },
  triple: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 980px)': {
      gridTemplateColumns: '1fr',
    },
  },
  stack: {
    display: 'grid',
    gap: '12px',
  },
  row: {
    display: 'flex',
    gap: '10px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  toolbarRow: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '12px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  preview: {
    display: 'grid',
    gap: '12px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  navRail: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  navItem: {
    display: 'grid',
    gap: '4px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  formGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 760px)': {
      gridTemplateColumns: '1fr',
    },
  },
  contentBlock: {
    display: 'grid',
    gap: '10px',
  },
  overlayStage: {
    display: 'grid',
    gap: '12px',
    minHeight: '180px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(180deg, rgba(98,100,167,0.12) 0%, rgba(98,100,167,0.03) 48%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  imageFrame: {
    width: '100%',
    height: '160px',
    objectFit: 'cover',
    borderRadius: tokens.borderRadiusLarge,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  personaStack: {
    display: 'grid',
    gap: '10px',
  },
  treeFrame: {
    display: 'grid',
    gap: '8px',
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  textScale: {
    display: 'grid',
    gap: '4px',
  },
  cardGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 760px)': {
      gridTemplateColumns: '1fr',
    },
  },
  helperNote: {
    color: tokens.colorNeutralForeground3,
  },
  providerPreview: {
    display: 'grid',
    gap: '10px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  providerInner: {
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
});

type GalleryProps = {
  title: string;
  description: string;
  emphasis: string;
  coverage: string[];
  stats: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
  children: React.ReactNode;
};

function PageScaffold({ title, description, emphasis, coverage, stats, children }: GalleryProps) {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.hero}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Fluent 2 React
          </Badge>
          <Badge appearance="tint">组件页</Badge>
          <Badge appearance="outline">{emphasis}</Badge>
        </LabBadgeRow>
        <div className={styles.heroRow}>
          <div className={styles.stack}>
            <Title3>{title}</Title3>
            <Text>{description}</Text>
            <div className={styles.coverage}>
              {coverage.map(item => (
                <Tag key={item}>{item}</Tag>
              ))}
            </div>
          </div>
          <LabStatGrid items={stats} />
        </div>
      </header>
      <div className={styles.layout}>{children}</div>
    </div>
  );
}

function ComponentSection({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <LabSurfaceCard>
      <LabSectionTitle title={title} description={description} />
      {children}
    </LabSurfaceCard>
  );
}

function ComponentsNoted({ names, note }: { names: string[]; note: string }) {
  const styles = useStyles();

  return (
    <LabSurfaceCard subtle>
      <LabSectionTitle title="目录补充" description={note} />
      <div className={styles.coverage}>
        {names.map(name => (
          <Badge key={name} appearance="outline">
            {name}
          </Badge>
        ))}
      </div>
    </LabSurfaceCard>
  );
}

function TagPickerDemo() {
  const [query, setQuery] = React.useState('');
  const [selectedOptions, setSelectedOptions] = React.useState<string[]>(['Fluent 2']);
  const options = React.useMemo(() => ['Fluent 2', 'React', 'Teams', 'Workbench', 'Accessibility', 'Motion'], []);

  const children = useTagPickerFilter({
    query,
    options,
    noOptionsElement: <TagPickerOption value="no-results">没有匹配项</TagPickerOption>,
    renderOption: option => (
      <TagPickerOption key={option} value={option}>
        {option}
      </TagPickerOption>
    ),
    filter: option => !selectedOptions.includes(option),
  });

  return (
    <TagPicker
      onOptionSelect={(_, data) => {
        setSelectedOptions(data.selectedOptions as string[]);
        setQuery('');
      }}
      selectedOptions={selectedOptions}
    >
      <TagPickerControl secondaryAction={<TagPickerButton aria-label="打开候选标签" />}>
        <TagPickerGroup aria-label="Selected tags">
          {selectedOptions.map(option => (
            <Tag key={option}>{option}</Tag>
          ))}
        </TagPickerGroup>
        <TagPickerInput
          aria-label="选择标签"
          value={query}
          onChange={event => setQuery(event.target.value)}
        />
      </TagPickerControl>
      <TagPickerList>{children}</TagPickerList>
    </TagPicker>
  );
}

function ProviderPreview() {
  return (
    <FluentProvider theme={teamsLightTheme}>
      <div>
        <Badge appearance="filled" color="brand">
          Teams Light Theme
        </Badge>
        <Caption1>通过嵌套 Provider 快速验证另一套主题 token 的落点。</Caption1>
      </div>
    </FluentProvider>
  );
}

export function FluentActionsNavigationComponentsPage() {
  const styles = useStyles();

  return (
    <PageScaffold
      title="React 组件页：命令与导航"
      description="对照 Fluent 2 官方 React 组件目录，把命令、导航、分层切换和信息检索类组件集中到一页，用来验证工具栏和页面导航语法。"
      emphasis="命令与导航"
      coverage={['Button', 'Breadcrumb', 'Link', 'Menu', 'Toolbar', 'TabList', 'SearchBox', 'Tree', 'Accordion', 'Nav']}
      stats={[
        { label: '覆盖组件', value: '10', tone: 'brand' },
        { label: '真实示例', value: '10', tone: 'success' },
        { label: '目录补全', value: '完成', tone: 'warning' },
      ]}
    >
      <div className={styles.grid}>
        <ComponentSection title="命令条与按钮" description="主操作、次操作和工具栏命令区。">
          <div className={styles.toolbarRow}>
            <Toolbar aria-label="Command toolbar">
              <ToolbarButton appearance="primary">创建</ToolbarButton>
              <ToolbarButton>刷新</ToolbarButton>
              <ToolbarButton>导出</ToolbarButton>
            </Toolbar>
            <div className={styles.row}>
              <Button appearance="secondary">次级动作</Button>
              <Button appearance="primary">主动作</Button>
            </div>
          </div>
        </ComponentSection>

        <ComponentSection title="面包屑与链接" description="用于层级感知和跨页面导航。">
          <div className={styles.stack}>
            <Breadcrumb aria-label="Breadcrumb">
              <BreadcrumbItem>
                <BreadcrumbButton>Workspace</BreadcrumbButton>
              </BreadcrumbItem>
              <BreadcrumbDivider />
              <BreadcrumbItem>
                <BreadcrumbButton>Design system</BreadcrumbButton>
              </BreadcrumbItem>
              <BreadcrumbDivider />
              <BreadcrumbItem>
                <BreadcrumbButton current>Components</BreadcrumbButton>
              </BreadcrumbItem>
            </Breadcrumb>
            <div className={styles.row}>
              <Link href="#">查看 Storybook</Link>
              <Link inline href="#">
                打开规范说明
              </Link>
            </div>
          </div>
        </ComponentSection>

        <ComponentSection title="Menu 与上下文命令" description="适合头像、按钮或行级更多操作。">
          <div className={styles.preview}>
            <Menu>
              <MenuTrigger disableButtonEnhancement>
                <Button>更多操作</Button>
              </MenuTrigger>
              <MenuPopover>
                <MenuList>
                  <MenuItem>重命名</MenuItem>
                  <MenuItem>复制链接</MenuItem>
                  <MenuItem>移动到归档</MenuItem>
                </MenuList>
              </MenuPopover>
            </Menu>
            <Caption1>把隐藏操作收进 menu，避免把工具栏做成一排低频按钮。</Caption1>
          </div>
        </ComponentSection>

        <ComponentSection title="TabList 与分区切换" description="适合同一页内的轻量模式切换。">
          <TabList defaultSelectedValue="overview">
            <Tab value="overview">概览</Tab>
            <Tab value="active">处理中</Tab>
            <Tab value="archive">归档</Tab>
          </TabList>
        </ComponentSection>

        <ComponentSection title="SearchBox 与检索入口" description="顶部检索或页面内过滤。">
          <SearchBox placeholder="搜索组件、模式或示例" />
        </ComponentSection>

        <ComponentSection title="Tree 与层级浏览" description="分类、结构树和目录浏览。">
          <div className={styles.treeFrame}>
            <Tree aria-label="component tree" defaultOpenItems={['inputs']}>
              <TreeItem itemType="branch" value="inputs">
                <TreeItemLayout>Inputs</TreeItemLayout>
                <Tree>
                  <TreeItem itemType="leaf" value="input">
                    <TreeItemLayout>Input</TreeItemLayout>
                  </TreeItem>
                  <TreeItem itemType="leaf" value="textarea">
                    <TreeItemLayout>Textarea</TreeItemLayout>
                  </TreeItem>
                </Tree>
              </TreeItem>
              <TreeItem itemType="branch" value="navigation">
                <TreeItemLayout>Navigation</TreeItemLayout>
                <Tree>
                  <TreeItem itemType="leaf" value="breadcrumb">
                    <TreeItemLayout>Breadcrumb</TreeItemLayout>
                  </TreeItem>
                </Tree>
              </TreeItem>
            </Tree>
          </div>
        </ComponentSection>

        <ComponentSection title="Accordion 与收纳结构" description="折叠信息块和规则列表。">
          <Accordion collapsible>
            <AccordionItem value="usage">
              <AccordionHeader>使用时机</AccordionHeader>
              <AccordionPanel>适合长说明或次级配置项，避免把整页说明直接铺满。</AccordionPanel>
            </AccordionItem>
            <AccordionItem value="dos">
              <AccordionHeader>注意事项</AccordionHeader>
              <AccordionPanel>让每个折叠块只承担一个解释主题，不要把所有信息堆进同一面板。</AccordionPanel>
            </AccordionItem>
          </Accordion>
        </ComponentSection>

        <ComponentSection title="Nav 与应用壳导航" description="主导航、分组标题和子项层级。">
          <div className={styles.navRail}>
            <Nav selectedValue="overview">
              <NavItem value="overview">概览</NavItem>
              <NavCategory value="governance">
                <NavCategoryItem>治理中心</NavCategoryItem>
                <NavSubItemGroup>
                  <NavSubItem value="menus">菜单治理</NavSubItem>
                  <NavSubItem value="roles">角色策略</NavSubItem>
                </NavSubItemGroup>
              </NavCategory>
              <NavSectionHeader>支持区域</NavSectionHeader>
              <NavItem value="help">帮助中心</NavItem>
            </Nav>
          </div>
        </ComponentSection>
      </div>
    </PageScaffold>
  );
}

export function FluentFormsSelectionComponentsPage() {
  const styles = useStyles();

  return (
    <PageScaffold
      title="React 组件页：表单与选择"
      description="把字段、选择、输入和轻量评分组件组织成一页，用来验证后台设置页和编辑页的字段语法。"
      emphasis="表单与选择"
      coverage={['Field', 'Label', 'InfoLabel', 'Input', 'Textarea', 'Combobox', 'Dropdown', 'Select', 'Checkbox', 'RadioGroup', 'Switch', 'Slider', 'SpinButton', 'Rating', 'TagPicker']}
      stats={[
        { label: '覆盖组件', value: '15', tone: 'brand' },
        { label: '真实示例', value: '15', tone: 'success' },
        { label: '目录补全', value: '完成', tone: 'warning' },
      ]}
    >
      <div className={styles.grid}>
        <ComponentSection title="Field 与文本输入" description="字段容器、标签和短文本输入。">
          <div className={styles.formGrid}>
            <Field label="租户名称">
              <Input placeholder="输入租户名称" />
            </Field>
            <Field label="环境代号" hint="通常使用简短的环境标识">
              <Input placeholder="prod-cn" />
            </Field>
          </div>
        </ComponentSection>

        <ComponentSection title="Label 与基础命名" description="独立标签适合自定义组合布局。">
          <div className={styles.stack}>
            <Label htmlFor="policy-name">策略名称</Label>
            <Input id="policy-name" placeholder="输入策略名称" />
          </div>
        </ComponentSection>

        <ComponentSection title="InfoLabel 与说明型标签" description="适合解释复杂字段。">
          <div className={styles.stack}>
            <InfoLabel info="切换后会影响默认显示范围。">默认菜单空间</InfoLabel>
            <Field label="详细说明">
              <Textarea placeholder="输入较长说明" resize="vertical" />
            </Field>
          </div>
        </ComponentSection>

        <ComponentSection title="Combobox / Dropdown / Select" description="列表选择与带检索的输入。">
          <div className={styles.formGrid}>
            <Field label="Combobox">
              <Combobox placeholder="选择团队">
                <Option>核心平台</Option>
                <Option>数据治理</Option>
                <Option>伙伴运营</Option>
              </Combobox>
            </Field>
            <Field label="Dropdown">
              <Dropdown placeholder="选择环境">
                <Option>Production</Option>
                <Option>Preview</Option>
                <Option>Sandbox</Option>
              </Dropdown>
            </Field>
            <Field label="Select">
              <Select defaultValue="viewer">
                <option value="viewer">Viewer</option>
                <option value="editor">Editor</option>
                <option value="owner">Owner</option>
              </Select>
            </Field>
          </div>
        </ComponentSection>

        <ComponentSection title="复选、单选与开关" description="布尔控制和枚举选择。">
          <div className={styles.stack}>
            <Checkbox label="启用风险提醒" defaultChecked />
            <Switch label="允许跨空间显示" defaultChecked />
            <RadioGroup defaultValue="managed" layout="horizontal">
              <Radio value="managed" label="托管模式" />
              <Radio value="manual" label="手动模式" />
            </RadioGroup>
          </div>
        </ComponentSection>

        <ComponentSection title="Slider / SpinButton / Rating" description="数值范围、步进输入和体验评分。">
          <div className={styles.stack}>
            <Field label="密度">
              <Slider defaultValue={40} />
            </Field>
            <Field label="轮询频率">
              <SpinButton defaultValue={15} />
            </Field>
            <div className={styles.stack}>
              <InfoLabel info="用于轻量主观评分。">样式满意度</InfoLabel>
              <Rating value={3} max={5} />
            </div>
          </div>
        </ComponentSection>

        <ComponentSection title="Tag 与已选值" description="轻量标签与已选择项。">
          <div className={styles.row}>
            <Tag>Foundation</Tag>
            <Tag appearance="brand">Fluent Web</Tag>
            <Tag>Teams</Tag>
          </div>
        </ComponentSection>

        <ComponentSection title="TagPicker 与多值选择" description="输入、筛选和多值选择一体化。">
          <TagPickerDemo />
        </ComponentSection>
      </div>
    </PageScaffold>
  );
}

function OverlayDemo() {
  const styles = useStyles();
  const [isDrawerOpen, setIsDrawerOpen] = React.useState(false);
  const { dispatchToast } = useToastController();

  return (
    <div className={styles.stack}>
      <div className={styles.row}>
        <Dialog>
          <DialogTrigger disableButtonEnhancement>
            <Button>Dialog</Button>
          </DialogTrigger>
          <DialogSurface>
            <DialogBody>
              <DialogTitle>删除预览</DialogTitle>
              <DialogContent>这个 Dialog 用来展示危险操作的二次确认。</DialogContent>
              <DialogActions>
                <DialogTrigger disableButtonEnhancement>
                  <Button appearance="secondary">取消</Button>
                </DialogTrigger>
                <Button appearance="primary">确认</Button>
              </DialogActions>
            </DialogBody>
          </DialogSurface>
        </Dialog>

        <Popover>
          <PopoverTrigger disableButtonEnhancement>
            <Button>Popover</Button>
          </PopoverTrigger>
          <PopoverSurface>
            <Text>Popover 适合非阻断型上下文说明。</Text>
          </PopoverSurface>
        </Popover>

        <Tooltip content="Tooltip 提供就近补充信息" relationship="label">
          <Button>Tooltip</Button>
        </Tooltip>

        <Button onClick={() => setIsDrawerOpen(true)}>Drawer</Button>
        <Button
          onClick={() =>
            dispatchToast(
              <Toast>
                <ToastTitle>保存成功</ToastTitle>
              </Toast>,
              { intent: 'success' },
            )
          }
        >
          Toast
        </Button>
      </div>

      <Drawer open={isDrawerOpen} onOpenChange={(_, data) => setIsDrawerOpen(data.open)} position="end">
        <DrawerHeader>
          <DrawerHeaderTitle action={<Button appearance="subtle" onClick={() => setIsDrawerOpen(false)}>关闭</Button>}>
            侧边抽屉
          </DrawerHeaderTitle>
        </DrawerHeader>
        <DrawerBody>
          <Text>Drawer 适合承接二级详情、属性区和轻编辑表单。</Text>
        </DrawerBody>
        <DrawerFooter>
          <Button appearance="secondary" onClick={() => setIsDrawerOpen(false)}>
            取消
          </Button>
          <Button appearance="primary">保存</Button>
        </DrawerFooter>
      </Drawer>
    </div>
  );
}

export function FluentFeedbackOverlaysComponentsPage() {
  const styles = useStyles();

  return (
    <PageScaffold
      title="React 组件页：反馈与浮层"
      description="把提示、进度、加载、浮层和反馈组件集中成一页，用来验证工作台里的状态表达和补充层级。"
      emphasis="反馈与浮层"
      coverage={['Dialog', 'Drawer', 'Popover', 'Tooltip', 'MessageBar', 'ProgressBar', 'Spinner', 'Skeleton', 'Toast', 'Badge', 'Divider']}
      stats={[
        { label: '覆盖组件', value: '11', tone: 'brand' },
        { label: '真实示例', value: '10', tone: 'success' },
        { label: '反馈层级', value: '4', tone: 'warning' },
      ]}
    >
      <Toaster />
      <div className={styles.grid}>
        <ComponentSection title="浮层组件" description="对话框、抽屉、气泡和提示信息。">
          <div className={styles.overlayStage}>
            <OverlayDemo />
          </div>
        </ComponentSection>

        <ComponentSection title="全局与局部反馈" description="MessageBar、Badge 和 Divider。">
          <div className={styles.stack}>
            <MessageBar>这是用于页面级提醒的 MessageBar。</MessageBar>
            <div className={styles.row}>
              <Badge appearance="filled" color="brand">
                处理中
              </Badge>
              <Badge appearance="tint" color="success">
                已完成
              </Badge>
              <Badge appearance="tint" color="warning">
                需关注
              </Badge>
            </div>
            <Divider />
            <Caption1>Divider 用来分隔不同反馈块，避免页面节奏失控。</Caption1>
          </div>
        </ComponentSection>

        <ComponentSection title="进度与加载" description="ProgressBar、Spinner 与 Skeleton。">
          <div className={styles.stack}>
            <ProgressBar value={0.62} />
            <Spinner label="同步中" />
            <div className={styles.stack}>
              <Skeleton>
                <div className={styles.stack}>
                  <div className={styles.preview} />
                  <div className={styles.navItem} />
                  <div className={styles.navItem} />
                </div>
              </Skeleton>
            </div>
          </div>
        </ComponentSection>

        <ComponentSection title="状态说明" description="反馈组件应该强调时机，而不是一页里同时堆很多提示。">
          <div className={styles.stack}>
            <Caption1>Dialog：阻断型确认</Caption1>
            <Caption1>Drawer：二级详情与轻编辑</Caption1>
            <Caption1>Popover / Tooltip：就近解释</Caption1>
            <Caption1>Toast：瞬时结果反馈</Caption1>
          </div>
        </ComponentSection>
      </div>
    </PageScaffold>
  );
}

export function FluentIdentityContentComponentsPage() {
  const styles = useStyles();

  return (
    <PageScaffold
      title="React 组件页：身份与内容"
      description="对照 Fluent 2 React 组件目录，把身份信息、内容容器、图像和文字系统集中展示，方便验证内容页与目录页的基础积木。"
      emphasis="身份与内容"
      coverage={['Avatar', 'AvatarGroup', 'Persona', 'Card', 'Image', 'List', 'Text', 'Tag', 'Accordion', 'Carousel', 'Icon', 'FluentProvider']}
      stats={[
        { label: '覆盖组件', value: '12', tone: 'brand' },
        { label: '真实示例', value: '12', tone: 'success' },
        { label: '图标承接', value: '独立页', tone: 'warning' },
      ]}
    >
      <div className={styles.grid}>
        <ComponentSection title="Avatar / AvatarGroup / Persona" description="人员、群组和上下文身份信息。">
          <div className={styles.personaStack}>
            <div className={styles.row}>
              <Avatar name="Ada Lovelace" color="brand" />
              <Avatar name="Grace Hopper" color="colorful" />
              <AvatarGroup layout="stack">
                <Avatar name="A" />
                <Avatar name="B" />
                <Avatar name="C" />
              </AvatarGroup>
            </div>
            <Persona
              name="北区运营负责人"
              secondaryText="负责事件响应与升级处理"
              presence={{ status: 'available' }}
            />
          </div>
        </ComponentSection>

        <ComponentSection title="Card 与内容容器" description="目录、详情和轻量摘要的内容块。">
          <div className={styles.cardGrid}>
            <Card>
              <CardHeader header={<Body1Strong>组件目录</Body1Strong>} description={<Caption1>说明型卡片</Caption1>} />
            </Card>
            <Card>
              <CardHeader header={<Body1Strong>审阅队列</Body1Strong>} description={<Caption1>带状态的容器</Caption1>} />
            </Card>
          </div>
        </ComponentSection>

        <ComponentSection title="Image 与视觉容器" description="图像类内容和预览模块。">
          <div className={styles.contentBlock}>
            <Image
              className={styles.imageFrame}
              src="https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&w=1200&q=80"
              alt="Office workspace"
            />
            <Caption1>Image 用于承载真实内容，不建议用作工作台里的装饰背景。</Caption1>
          </div>
        </ComponentSection>

        <ComponentSection title="Carousel 与轮播内容" description="适合精选内容、模板轮播或宣传位。">
          <Carousel activeIndex={0}>
            <CarouselViewport>
              <CarouselSlider>
                <CarouselCard>
                  <LabSurfaceCard subtle>
                    <Body1Strong>模板工作区</Body1Strong>
                    <Caption1>用于展示高频模板或精选页面样式。</Caption1>
                  </LabSurfaceCard>
                </CarouselCard>
                <CarouselCard>
                  <LabSurfaceCard subtle>
                    <Body1Strong>协作频道</Body1Strong>
                    <Caption1>用于展示会议、文件、公告等协作工作面。</Caption1>
                  </LabSurfaceCard>
                </CarouselCard>
                <CarouselCard>
                  <LabSurfaceCard subtle>
                    <Body1Strong>规范页</Body1Strong>
                    <Caption1>用于展示设计 Token、数据网格和导航结构规则。</Caption1>
                  </LabSurfaceCard>
                </CarouselCard>
              </CarouselSlider>
            </CarouselViewport>
          </Carousel>
        </ComponentSection>

        <ComponentSection title="List / Tag / Text" description="列表内容、标签和文字层级。">
          <div className={styles.stack}>
            <List>
              <ListItem>List 适合简单的竖向信息重复。</ListItem>
              <ListItem>Tag 适合表示已选值或分类。</ListItem>
              <ListItem>Text 组件用于统一排版语法。</ListItem>
            </List>
            <div className={styles.row}>
              <Tag>Marketplace</Tag>
              <Tag>Governance</Tag>
              <Tag>Teams</Tag>
            </div>
            <div className={styles.textScale}>
              <Text weight="semibold">Semibold text</Text>
              <Text>Body text for content sections.</Text>
            </div>
          </div>
        </ComponentSection>

        <ComponentSection title="FluentProvider 与主题作用域" description="局部主题试验和主题切换演示。">
          <div className={styles.providerPreview}>
            <Caption1>当前实验页在默认 Provider 下运行，下面演示嵌套 Teams 主题。</Caption1>
            <div className={styles.providerInner}>
              <ProviderPreview />
            </div>
          </div>
        </ComponentSection>
      </div>

      <LabSurfaceCard subtle>
        <LabSectionTitle
          title="图标承接说明"
          description="官方目录里的 Icon 能力已由独立的“React 图标总览”页承接，用于浏览并复制图标 ID。"
        />
        <Caption1 className={styles.helperNote}>这样做比在身份页里塞少量示例更实用，也更适合后续直接查找图标导出名。</Caption1>
      </LabSurfaceCard>
    </PageScaffold>
  );
}
