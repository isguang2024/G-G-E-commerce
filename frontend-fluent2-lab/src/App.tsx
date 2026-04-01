import * as React from 'react';
import {
  Badge,
  Body1,
  Button,
  Caption1,
  Divider,
  FluentProvider,
  Spinner,
  Switch,
  Tab,
  TabList,
  Title2,
  makeStyles,
  tokens,
  webDarkTheme,
  webLightTheme,
} from '@fluentui/react-components';
import { PageKey, PageSource, pages } from './lab/catalog';

const pageComponents: Record<PageKey, React.LazyExoticComponent<() => JSX.Element>> = {
  'analytics-overview': React.lazy(() =>
    import('./pages/AnalyticsOverviewPage').then(module => ({ default: module.AnalyticsOverviewPage })),
  ),
  'app-shell': React.lazy(() =>
    import('./pages/AppShellShowcasePage').then(module => ({ default: module.AppShellShowcasePage })),
  ),
  'project-board': React.lazy(() =>
    import('./pages/ProjectBoardPage').then(module => ({ default: module.ProjectBoardPage })),
  ),
  'settings-center': React.lazy(() =>
    import('./pages/SettingsCenterPage').then(module => ({ default: module.SettingsCenterPage })),
  ),
  'people-directory': React.lazy(() =>
    import('./pages/PeopleDirectoryPage').then(module => ({ default: module.PeopleDirectoryPage })),
  ),
  'calendar-planner': React.lazy(() =>
    import('./pages/CalendarPlannerPage').then(module => ({ default: module.CalendarPlannerPage })),
  ),
  'resource-library': React.lazy(() =>
    import('./pages/ResourceLibraryPage').then(module => ({ default: module.ResourceLibraryPage })),
  ),
  'notification-preferences': React.lazy(() =>
    import('./pages/NotificationPreferencesPage').then(module => ({ default: module.NotificationPreferencesPage })),
  ),
  'audit-timeline': React.lazy(() =>
    import('./pages/AuditTimelinePage').then(module => ({ default: module.AuditTimelinePage })),
  ),
  'fluent-gallery': React.lazy(() =>
    import('./pages/FluentComponentGalleryPage').then(module => ({ default: module.FluentComponentGalleryPage })),
  ),
  'react-actions-navigation': React.lazy(() =>
    import('./pages/FluentReactComponentPages').then(module => ({ default: module.FluentActionsNavigationComponentsPage })),
  ),
  'react-forms-selection': React.lazy(() =>
    import('./pages/FluentReactComponentPages').then(module => ({ default: module.FluentFormsSelectionComponentsPage })),
  ),
  'react-feedback-overlays': React.lazy(() =>
    import('./pages/FluentReactComponentPages').then(module => ({ default: module.FluentFeedbackOverlaysComponentsPage })),
  ),
  'react-identity-content': React.lazy(() =>
    import('./pages/FluentReactComponentPages').then(module => ({ default: module.FluentIdentityContentComponentsPage })),
  ),
  'react-icon-library': React.lazy(() =>
    import('./pages/FluentReactIconLibraryPage').then(module => ({ default: module.FluentReactIconLibraryPage })),
  ),
  'react-composition-navigation-command': React.lazy(() =>
    import('./pages/FluentReactCompositionPages').then(module => ({
      default: module.FluentNavigationCommandPatternsPage,
    })),
  ),
  'react-composition-form-feedback': React.lazy(() =>
    import('./pages/FluentReactCompositionPages').then(module => ({
      default: module.FluentFormFeedbackPatternsPage,
    })),
  ),
  'react-composition-content-collaboration': React.lazy(() =>
    import('./pages/FluentReactCompositionPages').then(module => ({
      default: module.FluentContentCollaborationPatternsPage,
    })),
  ),
  'fluent-spec': React.lazy(() =>
    import('./pages/FluentSpecWorkspacePage').then(module => ({ default: module.FluentSpecWorkspacePage })),
  ),
  'approval-workbench': React.lazy(() =>
    import('./pages/ApprovalWorkbenchPage').then(module => ({ default: module.ApprovalWorkbenchPage })),
  ),
  'teams-responsive': React.lazy(() =>
    import('./pages/TeamsResponsivenessPage').then(module => ({ default: module.TeamsResponsivenessPage })),
  ),
  'teams-channel': React.lazy(() =>
    import('./pages/TeamsChannelWorkspacePage').then(module => ({ default: module.TeamsChannelWorkspacePage })),
  ),
  'teams-conversation': React.lazy(() =>
    import('./pages/TeamsConversationPage').then(module => ({ default: module.TeamsConversationPage })),
  ),
  'message-center': React.lazy(() =>
    import('./pages/MessageCenterPage').then(module => ({ default: module.MessageCenterPage })),
  ),
  'login-showcase': React.lazy(() =>
    import('./pages/LoginShowcasePage').then(module => ({ default: module.LoginShowcasePage })),
  ),
  'grid-drawer': React.lazy(() =>
    import('./pages/WorkspaceGridDrawerPage').then(module => ({ default: module.WorkspaceGridDrawerPage })),
  ),
  'tree-detail': React.lazy(() =>
    import('./pages/TreeDetailPanePage').then(module => ({ default: module.TreeDetailPanePage })),
  ),
  'form-sections': React.lazy(() =>
    import('./pages/FormSectionsPage').then(module => ({ default: module.FormSectionsPage })),
  ),
  'operations-command': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.OperationsCommandCenterPage })),
  ),
  'tenant-overview': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.TenantOverviewPage })),
  ),
  'billing-operations': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.BillingOperationsPage })),
  ),
  'incident-response': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.IncidentResponsePage })),
  ),
  'policy-studio': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.PolicyStudioPage })),
  ),
  'asset-inventory': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.AssetInventoryPage })),
  ),
  'support-desk': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.SupportDeskPage })),
  ),
  'search-workspace': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.SearchWorkspacePage })),
  ),
  'release-control': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.ReleaseControlPage })),
  ),
  'knowledge-hub': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.KnowledgeHubPage })),
  ),
  'shell-guidelines': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.ShellGuidelinesPage })),
  ),
  'data-grid-spec': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.DataGridSpecPage })),
  ),
  'form-patterns': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.FormPatternsPage })),
  ),
  'navigation-spec': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.NavigationSpecPage })),
  ),
  'token-governance': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.TokenGovernancePage })),
  ),
  'accessibility-review': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.AccessibilityReviewPage })),
  ),
  'handoff-patterns': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.HandoffPatternsPage })),
  ),
  'template-gallery': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.TemplateGalleryPage })),
  ),
  'motion-principles': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.MotionPrinciplesPage })),
  ),
  'sidepanel-reference': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.SidepanelReferencePage })),
  ),
  'meeting-command': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.MeetingCommandPage })),
  ),
  'frontline-briefing': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.FrontlineBriefingPage })),
  ),
  'file-collaboration': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.FileCollaborationPage })),
  ),
  'approval-conversation': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.ApprovalConversationPage })),
  ),
  'onboarding-hub': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.OnboardingHubPage })),
  ),
  'community-announcements': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.CommunityAnnouncementsPage })),
  ),
  'shift-handoff': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.ShiftHandoffPage })),
  ),
  'webinar-operations': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.WebinarOperationsPage })),
  ),
  'incident-swarm': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.IncidentSwarmPage })),
  ),
  'partner-standup': React.lazy(() =>
    import('./pages/ScenarioExpansionPages').then(module => ({ default: module.PartnerStandupPage })),
  ),
};

const useStyles = makeStyles({
  app: {
    minHeight: '100vh',
    background: `linear-gradient(180deg, ${tokens.colorNeutralBackground2} 0%, ${tokens.colorNeutralBackground3} 100%)`,
    color: tokens.colorNeutralForeground1,
  },
  shell: {
    width: 'min(1480px, 100%)',
    marginLeft: 'auto',
    marginRight: 'auto',
    paddingTop: '24px',
    paddingRight: '24px',
    paddingBottom: '32px',
    paddingLeft: '24px',
    display: 'grid',
    gap: '18px',
  },
  hero: {
    display: 'grid',
    gap: '14px',
    padding: '20px 22px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow8,
  },
  heroTop: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '16px',
    flexWrap: 'wrap',
  },
  heroMeta: {
    display: 'grid',
    gap: '10px',
    maxWidth: '880px',
  },
  controls: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    flexWrap: 'wrap',
  },
  sourceTabs: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  switcher: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  pageButton: {
    justifyContent: 'space-between',
    minWidth: '128px',
  },
  summaryRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  pageHost: {
    backgroundColor: tokens.colorNeutralBackground1,
    borderRadius: tokens.borderRadiusXLarge,
    boxShadow: tokens.shadow8,
    paddingTop: '24px',
    paddingRight: '24px',
    paddingBottom: '24px',
    paddingLeft: '24px',
  },
  loading: {
    minHeight: '320px',
    display: 'grid',
    placeItems: 'center',
  },
});

function renderPage(page: PageKey) {
  const PageComponent = pageComponents[page];
  return <PageComponent />;
}

const sourceMeta: Record<'all' | PageSource, { label: string; count: (items: typeof pages) => number }> = {
  all: {
    label: '全部',
    count: items => items.length,
  },
  foundation: {
    label: '共享基座',
    count: items => items.filter(item => item.source === 'foundation').length,
  },
  primitives: {
    label: '基础控件集',
    count: items => items.filter(item => item.source === 'primitives').length,
  },
  'fluent-web': {
    label: 'Fluent 2 Web',
    count: items => items.filter(item => item.source === 'fluent-web').length,
  },
  teams: {
    label: 'Teams',
    count: items => items.filter(item => item.source === 'teams').length,
  },
};

export default function App() {
  const styles = useStyles();
  const [currentPage, setCurrentPage] = React.useState<PageKey>('teams-channel');
  const [sourceFilter, setSourceFilter] = React.useState<'all' | PageSource>('all');
  const [isDark, setIsDark] = React.useState(false);
  const filteredPages = React.useMemo(
    () => pages.filter(page => (sourceFilter === 'all' ? true : page.source === sourceFilter)),
    [sourceFilter],
  );
  const currentMeta = pages.find(page => page.key === currentPage) ?? pages[0];

  React.useEffect(() => {
    if (!filteredPages.some(page => page.key === currentPage)) {
      setCurrentPage(filteredPages[0]?.key ?? 'teams-channel');
    }
  }, [currentPage, filteredPages]);

  return (
    <FluentProvider theme={isDark ? webDarkTheme : webLightTheme}>
      <div className={styles.app}>
        <div className={styles.shell}>
          <header className={styles.hero}>
            <div className={styles.heroTop}>
              <div className={styles.heroMeta}>
                <Title2>Fluent 2 React 设计实验场</Title2>
                <Body1>
                  这里不复刻现有 Vue 架构，只验证共享 Fluent 2 基座、Fluent 2 Web 文档式布局，以及 Teams
                  协作式布局能否长期共存。
                </Body1>
                <div className={styles.summaryRow}>
                  <Badge appearance="filled" color="brand">
                    {pages.length} 个页面
                  </Badge>
                  <Badge appearance="tint">基础控件集</Badge>
                  <Badge appearance="tint" color="success">
                    双设计源
                  </Badge>
                  <Badge appearance="outline">当前页：{currentMeta.label}</Badge>
                </div>
              </div>

              <div className={styles.controls}>
                <Caption1>{isDark ? '暗色预览' : '亮色预览'}</Caption1>
                <Switch checked={isDark} onChange={(_, data) => setIsDark(data.checked)} labelPosition="before" />
              </div>
            </div>

            <div className={styles.sourceTabs}>
              <TabList
                selectedValue={sourceFilter}
                onTabSelect={(_, data) => setSourceFilter(data.value as 'all' | PageSource)}
              >
                {Object.entries(sourceMeta).map(([key, value]) => (
                  <Tab key={key} value={key}>
                    {value.label} {value.count(pages)}
                  </Tab>
                ))}
              </TabList>
            </div>

            <div className={styles.switcher}>
              {filteredPages.map(page => (
                <Button
                  key={page.key}
                  className={styles.pageButton}
                  appearance={page.key === currentPage ? 'primary' : 'secondary'}
                  onClick={() => setCurrentPage(page.key)}
                >
                  {page.label}
                </Button>
              ))}
            </div>

            <Body1>{currentMeta.description}</Body1>
            <Caption1>设计侧重点：{currentMeta.emphasis}</Caption1>
          </header>

          <Divider />

          <main className={styles.pageHost}>
            <React.Suspense
              fallback={
                <div className={styles.loading}>
                  <Spinner label="正在加载测试页面" />
                </div>
              }
            >
              {renderPage(currentPage)}
            </React.Suspense>
          </main>
        </div>
      </div>
    </FluentProvider>
  );
}
