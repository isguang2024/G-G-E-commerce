import * as React from 'react';
import {
  Avatar,
  AvatarGroup,
  Badge,
  Body1,
  Body1Strong,
  Breadcrumb,
  BreadcrumbButton,
  BreadcrumbDivider,
  BreadcrumbItem,
  Button,
  Caption1,
  Card,
  CardHeader,
  Checkbox,
  Dialog,
  DialogActions,
  DialogBody,
  DialogContent,
  DialogSurface,
  DialogTitle,
  DialogTrigger,
  Dropdown,
  Field,
  Input,
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
  NavSubItem,
  NavSubItemGroup,
  Option,
  Persona,
  ProgressBar,
  SearchBox,
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
  Toolbar,
  ToolbarButton,
  makeStyles,
  tokens,
  useTagPickerFilter,
  Title3,
} from '@fluentui/react-components';
import { LabBadgeRow, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '20px',
  },
  hero: {
    display: 'grid',
    gap: '14px',
    padding: '22px 24px',
    borderRadius: tokens.borderRadiusXLarge,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.12) 0%, rgba(91,95,199,0.05) 48%, rgba(255,255,255,0.02) 100%)',
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1.15fr 0.85fr',
    gap: '16px',
    '@media (max-width: 980px)': {
      gridTemplateColumns: '1fr',
    },
  },
  coverage: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  stack: {
    display: 'grid',
    gap: '12px',
  },
  row: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
    alignItems: 'center',
  },
  commandLayout: {
    display: 'grid',
    gridTemplateColumns: '260px minmax(0, 1fr)',
    gap: '16px',
    '@media (max-width: 1080px)': {
      gridTemplateColumns: '1fr',
    },
  },
  navPanel: {
    display: 'grid',
    gap: '14px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  commandMain: {
    display: 'grid',
    gap: '16px',
  },
  masthead: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  toolbarRow: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '12px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  splitColumns: {
    display: 'grid',
    gridTemplateColumns: '1.3fr 0.7fr',
    gap: '16px',
    '@media (max-width: 1080px)': {
      gridTemplateColumns: '1fr',
    },
  },
  board: {
    display: 'grid',
    gap: '12px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  boardGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 880px)': {
      gridTemplateColumns: '1fr',
    },
  },
  issueCard: {
    display: 'grid',
    gap: '10px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  sideRail: {
    display: 'grid',
    gap: '12px',
  },
  sideSection: {
    display: 'grid',
    gap: '10px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  formLayout: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1.1fr) 340px',
    gap: '16px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  formWorkbench: {
    display: 'grid',
    gap: '16px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  formGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 820px)': {
      gridTemplateColumns: '1fr',
    },
  },
  feedbackRail: {
    display: 'grid',
    gap: '12px',
  },
  footerActions: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '12px',
    flexWrap: 'wrap',
    alignItems: 'center',
    paddingTop: '12px',
    borderTop: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  contentLayout: {
    display: 'grid',
    gridTemplateColumns: '1.05fr 0.95fr 280px',
    gap: '16px',
    '@media (max-width: 1240px)': {
      gridTemplateColumns: '1fr',
    },
  },
  editorialRail: {
    display: 'grid',
    gap: '12px',
  },
  showcaseBlock: {
    display: 'grid',
    gap: '12px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  cardStack: {
    display: 'grid',
    gap: '12px',
  },
  timelineCard: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  helperText: {
    color: tokens.colorNeutralForeground3,
  },
});

type PatternPageProps = {
  title: string;
  description: string;
  emphasis: string;
  coverage: string[];
  stats: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
  children: React.ReactNode;
};

function PatternPage({ title, description, emphasis, coverage, stats, children }: PatternPageProps) {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.hero}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Fluent 2 React
          </Badge>
          <Badge appearance="tint">组合模式页</Badge>
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
      {children}
    </div>
  );
}

function PatternSection({
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

function TopicTagPicker() {
  const [query, setQuery] = React.useState('');
  const [selectedOptions, setSelectedOptions] = React.useState<string[]>(['性能']);
  const options = React.useMemo(() => ['性能', '可访问性', '审批', '数据质量', '导航', '主题'], []);

  const children = useTagPickerFilter({
    query,
    options,
    noOptionsElement: <TagPickerOption value="no-match">没有更多标签</TagPickerOption>,
    renderOption: option => (
      <TagPickerOption key={option} value={option}>
        {option}
      </TagPickerOption>
    ),
    filter: option => !selectedOptions.includes(option),
  });

  return (
    <TagPicker
      selectedOptions={selectedOptions}
      onOptionSelect={(_, data) => {
        setSelectedOptions(data.selectedOptions as string[]);
        setQuery('');
      }}
    >
      <TagPickerControl secondaryAction={<TagPickerButton aria-label="打开主题标签" />}>
        <TagPickerGroup aria-label="selected topics">
          {selectedOptions.map(option => (
            <Tag key={option}>{option}</Tag>
          ))}
        </TagPickerGroup>
        <TagPickerInput aria-label="选择主题" value={query} onChange={event => setQuery(event.target.value)} />
      </TagPickerControl>
      <TagPickerList>{children}</TagPickerList>
    </TagPicker>
  );
}

export function FluentNavigationCommandPatternsPage() {
  const styles = useStyles();

  return (
    <PatternPage
      title="React 组合页：导航与命令工作台"
      description="把 Nav、Toolbar、SearchBox、TabList、Menu 和 Breadcrumb 组合成一张真实的导航指挥面，用来验证应用壳内部的主导航、检索和命令密度。"
      emphasis="导航与命令"
      coverage={['Nav', 'Toolbar', 'SearchBox', 'TabList', 'Menu', 'Breadcrumb', 'Persona', 'Badge']}
      stats={[
        { label: '工作面', value: '1', tone: 'brand' },
        { label: '组合组件', value: '8', tone: 'success' },
        { label: '主任务', value: '导航调度', tone: 'warning' },
      ]}
    >
      <div className={styles.commandLayout}>
        <aside className={styles.navPanel}>
          <SearchBox placeholder="搜索导航、命令或频道" />
          <Nav selectedValue="command-center">
            <NavItem value="command-center">指挥中心</NavItem>
            <NavCategory value="operations">
              <NavCategoryItem>运营区域</NavCategoryItem>
              <NavSubItemGroup>
                <NavSubItem value="signals">信号列表</NavSubItem>
                <NavSubItem value="rollout">发布节奏</NavSubItem>
                <NavSubItem value="handoff">交接窗口</NavSubItem>
              </NavSubItemGroup>
            </NavCategory>
            <NavCategory value="governance">
              <NavCategoryItem>治理区域</NavCategoryItem>
              <NavSubItemGroup>
                <NavSubItem value="menus">菜单治理</NavSubItem>
                <NavSubItem value="spaces">空间管理</NavSubItem>
              </NavSubItemGroup>
            </NavCategory>
          </Nav>
          <LabSurfaceCard subtle>
            <LabSectionTitle title="当前值守" description="让侧栏承担方向感，不承担全部操作。" />
            <Persona name="值班负责人" secondaryText="北区核心平台" />
          </LabSurfaceCard>
        </aside>

        <div className={styles.commandMain}>
          <section className={styles.masthead}>
            <Breadcrumb aria-label="Breadcrumb">
              <BreadcrumbItem>
                <BreadcrumbButton>Workspace</BreadcrumbButton>
              </BreadcrumbItem>
              <BreadcrumbDivider />
              <BreadcrumbItem>
                <BreadcrumbButton>Operations</BreadcrumbButton>
              </BreadcrumbItem>
              <BreadcrumbDivider />
              <BreadcrumbItem>
                <BreadcrumbButton current>Command deck</BreadcrumbButton>
              </BreadcrumbItem>
            </Breadcrumb>
            <div className={styles.toolbarRow}>
              <div className={styles.stack}>
                <Body1Strong>北区运行指挥台</Body1Strong>
                <Caption1>把主导航、命令条和检索入口收口到同一张工作台首屏。</Caption1>
              </div>
              <Toolbar aria-label="command toolbar">
                <ToolbarButton appearance="primary">创建任务</ToolbarButton>
                <ToolbarButton>刷新信号</ToolbarButton>
                <Menu>
                  <MenuTrigger disableButtonEnhancement>
                    <ToolbarButton>更多</ToolbarButton>
                  </MenuTrigger>
                  <MenuPopover>
                    <MenuList>
                      <MenuItem>导出值守清单</MenuItem>
                      <MenuItem>共享当前视图</MenuItem>
                      <MenuItem>固定到首页</MenuItem>
                    </MenuList>
                  </MenuPopover>
                </Menu>
              </Toolbar>
            </div>
            <TabList defaultSelectedValue="signals">
              <Tab value="signals">实时信号</Tab>
              <Tab value="actions">待执行动作</Tab>
              <Tab value="history">历史回放</Tab>
            </TabList>
          </section>

          <div className={styles.splitColumns}>
            <section className={styles.board}>
              <LabSectionTitle title="待处理信号" description="主工作区专注列表理解与命令触发。" />
              <div className={styles.boardGrid}>
                {[
                  { title: '身份服务抖动', status: '需关注', owner: '平台稳定性' },
                  { title: '菜单缓存失配', status: '处理中', owner: '导航治理' },
                  { title: '发布窗口确认', status: '已准备', owner: '发布经理' },
                ].map(item => (
                  <div key={item.title} className={styles.issueCard}>
                    <div className={styles.row}>
                      <Badge appearance="tint" color={item.status === '需关注' ? 'warning' : item.status === '处理中' ? 'brand' : 'success'}>
                        {item.status}
                      </Badge>
                      <Caption1>{item.owner}</Caption1>
                    </div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Body1>
                      点击信号后进入详情区，低频动作通过 menu 收纳，避免工作区出现按钮噪音。
                    </Body1>
                  </div>
                ))}
              </div>
            </section>

            <aside className={styles.sideRail}>
              <div className={styles.sideSection}>
                <LabSectionTitle title="当前责任人" description="侧栏承接上下文而不是主操作。" />
                <AvatarGroup layout="stack">
                  <Avatar name="陈" />
                  <Avatar name="林" />
                  <Avatar name="刘" />
                </AvatarGroup>
                <Persona name="平台值班经理" secondaryText="会在 14:30 接管发布节奏" />
              </div>
              <div className={styles.sideSection}>
                <LabSectionTitle title="命令准则" description="用短列表解释组合模式为什么成立。" />
                <List>
                  <ListItem>SearchBox 靠近导航，承担全局定位。</ListItem>
                  <ListItem>Toolbar 只保留高频动作。</ListItem>
                  <ListItem>TabList 管视图切换，不接管导航层级。</ListItem>
                </List>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </PatternPage>
  );
}

export function FluentFormFeedbackPatternsPage() {
  const styles = useStyles();

  return (
    <PatternPage
      title="React 组合页：表单与反馈工作台"
      description="把 Field、Input、Dropdown、TagPicker、Checkbox、Switch、MessageBar、ProgressBar 和 Dialog 组合成一张真实的提交工作台，用来验证表单与反馈的主次层级。"
      emphasis="表单与反馈"
      coverage={['Field', 'Input', 'Dropdown', 'TagPicker', 'Checkbox', 'Switch', 'MessageBar', 'ProgressBar', 'Dialog']}
      stats={[
        { label: '工作面', value: '1', tone: 'brand' },
        { label: '组合组件', value: '9', tone: 'success' },
        { label: '主任务', value: '提交与确认', tone: 'warning' },
      ]}
    >
      <div className={styles.formLayout}>
        <section className={styles.formWorkbench}>
          <LabSectionTitle
            title="策略发布表单"
            description="左侧承担完整输入，右侧承担状态、校验和确认，不把反馈散落在字段之间。"
          />
          <MessageBar>
            当前表单用于验证“字段输入 + 多值选择 + 提交反馈”是否能在一张 Fluent 2 页面里稳定共存。
          </MessageBar>
          <div className={styles.formGrid}>
            <Field label="策略名称" required>
              <Input placeholder="输入策略名称" />
            </Field>
            <Field label="发布环境">
              <Dropdown placeholder="选择环境">
                <Option>Production</Option>
                <Option>Preview</Option>
                <Option>Sandbox</Option>
              </Dropdown>
            </Field>
            <Field label="责任人">
              <Input placeholder="输入负责人" />
            </Field>
            <Field label="影响范围">
              <Dropdown defaultValue="核心工作区">
                <Option>核心工作区</Option>
                <Option>合作方空间</Option>
                <Option>仅试点租户</Option>
              </Dropdown>
            </Field>
          </div>
          <Field label="主题标签" hint="选择本次发布涉及的模块主题。">
            <TopicTagPicker />
          </Field>
          <Field label="变更说明">
            <Textarea resize="vertical" placeholder="说明本次策略变更的目标、影响面和回滚要点。" />
          </Field>
          <div className={styles.row}>
            <Checkbox label="需要值班确认" defaultChecked />
            <Checkbox label="更新知识库摘要" />
            <Switch label="同步通知相关团队" defaultChecked />
          </div>
          <div className={styles.footerActions}>
            <div className={styles.stack}>
              <Caption1>保存前仍需完成风险检查与审批确认。</Caption1>
              <ProgressBar value={0.68} />
            </div>
            <div className={styles.row}>
              <Button appearance="secondary">保存草稿</Button>
              <Dialog>
                <DialogTrigger disableButtonEnhancement>
                  <Button appearance="primary">提交发布</Button>
                </DialogTrigger>
                <DialogSurface>
                  <DialogBody>
                    <DialogTitle>确认提交发布</DialogTitle>
                    <DialogContent>这次提交会触发审批流和通知同步，是否继续？</DialogContent>
                    <DialogActions>
                      <DialogTrigger disableButtonEnhancement>
                        <Button appearance="secondary">取消</Button>
                      </DialogTrigger>
                      <Button appearance="primary">确认提交</Button>
                    </DialogActions>
                  </DialogBody>
                </DialogSurface>
              </Dialog>
            </div>
          </div>
        </section>

        <aside className={styles.feedbackRail}>
          <PatternSection title="校验摘要" description="把反馈收口在一侧，页面节奏会更稳。">
            <div className={styles.stack}>
              <div className={styles.row}>
                <Badge appearance="tint" color="success">
                  已通过
                </Badge>
                <Caption1>字段完整性</Caption1>
              </div>
              <div className={styles.row}>
                <Badge appearance="tint" color="warning">
                  待确认
                </Badge>
                <Caption1>风险审批链</Caption1>
              </div>
              <div className={styles.row}>
                <Badge appearance="tint" color="brand">
                  已准备
                </Badge>
                <Caption1>通知范围草案</Caption1>
              </div>
            </div>
          </PatternSection>

          <PatternSection title="反馈原则" description="表单页不应该把所有反馈塞成一排 MessageBar。">
            <List>
              <ListItem>字段级问题留在 Field 内部。</ListItem>
              <ListItem>页面级状态收口成一块 MessageBar。</ListItem>
              <ListItem>提交确认通过 Dialog 阻断，而不是二次跳页。</ListItem>
            </List>
          </PatternSection>
        </aside>
      </div>
    </PatternPage>
  );
}

export function FluentContentCollaborationPatternsPage() {
  const styles = useStyles();

  return (
    <PatternPage
      title="React 组合页：内容与协作工作台"
      description="把 Card、Persona、AvatarGroup、List、Tag 和内容块组合成一张内容协作面，用来验证知识流、责任人和精选内容如何在同一页里共存。"
      emphasis="内容与协作"
      coverage={['Card', 'Persona', 'AvatarGroup', 'List', 'Tag', 'Menu', 'Badge']}
      stats={[
        { label: '工作面', value: '1', tone: 'brand' },
        { label: '组合组件', value: '7', tone: 'success' },
        { label: '主任务', value: '内容协作', tone: 'warning' },
      ]}
    >
      <div className={styles.contentLayout}>
        <section className={styles.editorialRail}>
          <div className={styles.showcaseBlock}>
            <LabSectionTitle title="本周精选内容" description="左列强调内容主线，而不是平铺很多卡片。" />
            <Card>
              <CardHeader
                header={<Body1Strong>设计系统迁移节奏</Body1Strong>}
                description={<Caption1>作者：平台体验团队</Caption1>}
                action={
                  <Menu>
                    <MenuTrigger disableButtonEnhancement>
                      <Button appearance="subtle">操作</Button>
                    </MenuTrigger>
                    <MenuPopover>
                      <MenuList>
                        <MenuItem>固定到顶部</MenuItem>
                        <MenuItem>共享到频道</MenuItem>
                      </MenuList>
                    </MenuPopover>
                  </Menu>
                }
              />
              <Body1>
                这块用来测试 CardHeader、正文和轻操作是否能形成稳定的内容主视觉，而不是退化成普通列表卡。
              </Body1>
            </Card>
            <div className={styles.row}>
              <Tag appearance="brand">精选</Tag>
              <Tag>规范沉淀</Tag>
              <Tag>待扩散</Tag>
            </div>
          </div>

          <div className={styles.showcaseBlock}>
            <LabSectionTitle title="更新流" description="中频内容适合用时间线样式，而不是同尺寸卡片矩阵。" />
            <div className={styles.cardStack}>
              {[
                '导航结构规范完成第一轮评审',
                '表单反馈组合页进入设计实验线',
                '图标总览页开始支持复制基础 ID',
              ].map(item => (
                <div key={item} className={styles.timelineCard}>
                  <div className={styles.row}>
                    <Badge appearance="tint" color="brand">
                      更新
                    </Badge>
                    <Caption1>今日 09:20</Caption1>
                  </div>
                  <Body1Strong>{item}</Body1Strong>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className={styles.showcaseBlock}>
          <LabSectionTitle title="协作上下文" description="中列负责把责任人、讨论和任务串成工作流。" />
          <Persona
            name="设计系统负责人"
            secondaryText="当前在收口基础控件集的组合工作面"
            textAlignment="start"
          />
          <AvatarGroup layout="stack">
            <Avatar name="梁" />
            <Avatar name="赵" />
            <Avatar name="周" />
            <Avatar name="吴" />
          </AvatarGroup>
          <MessageBar>这块测试内容页里是否能自然承载协作提醒，而不需要跳成纯 Teams 页面。</MessageBar>
          <List>
            <ListItem>将组件页与场景页继续分层。</ListItem>
            <ListItem>给图标页增加真实导入名复制。</ListItem>
            <ListItem>补浏览器回归，检查暗色主题层级。</ListItem>
          </List>
        </section>

        <aside className={styles.sideRail}>
          <div className={styles.sideSection}>
            <LabSectionTitle title="协作原则" description="右侧只放上下文，不抢内容主线。" />
            <Caption1 className={styles.helperText}>内容页主线应该是内容本身，协作上下文作为辅助层出现。</Caption1>
          </div>
          <div className={styles.sideSection}>
            <LabSectionTitle title="状态标签" />
            <div className={styles.stack}>
              <Badge appearance="tint" color="success">
                可发布
              </Badge>
              <Badge appearance="tint" color="warning">
                待评审
              </Badge>
              <Badge appearance="tint" color="brand">
                进行中
              </Badge>
            </div>
          </div>
        </aside>
      </div>
    </PatternPage>
  );
}
