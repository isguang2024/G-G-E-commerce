import * as React from 'react';
import {
  Body1,
  Button,
  Divider,
  Title2,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { WorkspaceGridDrawerPage } from './pages/WorkspaceGridDrawerPage';
import { TreeDetailPanePage } from './pages/TreeDetailPanePage';
import { FormSectionsPage } from './pages/FormSectionsPage';

type PageKey = 'grid-drawer' | 'tree-detail' | 'form-sections';

const useStyles = makeStyles({
  app: {
    minHeight: '100vh',
    backgroundColor: tokens.colorNeutralBackground2,
    color: tokens.colorNeutralForeground1,
  },
  shell: {
    width: 'min(1400px, 100%)',
    marginLeft: 'auto',
    marginRight: 'auto',
    paddingTop: '24px',
    paddingRight: '24px',
    paddingBottom: '32px',
    paddingLeft: '24px',
    display: 'grid',
    gap: '16px',
  },
  hero: {
    display: 'grid',
    gap: '8px',
  },
  switcher: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  pageHost: {
    backgroundColor: tokens.colorNeutralBackground1,
    borderRadius: tokens.borderRadiusLarge,
    boxShadow: tokens.shadow4,
    paddingTop: '20px',
    paddingRight: '20px',
    paddingBottom: '20px',
    paddingLeft: '20px',
  },
});

const pages: Array<{ key: PageKey; label: string; description: string }> = [
  {
    key: 'grid-drawer',
    label: '列表 + 抽屉',
    description: '适合治理页、资源列表和右侧详情编辑场景。',
  },
  {
    key: 'tree-detail',
    label: '树 + 详情',
    description: '适合菜单树、资源树、分类树和结构管理场景。',
  },
  {
    key: 'form-sections',
    label: '分组表单',
    description: '适合设置页、详情编辑页和抽屉配置表单。',
  },
];

function renderPage(page: PageKey) {
  switch (page) {
    case 'grid-drawer':
      return <WorkspaceGridDrawerPage />;
    case 'tree-detail':
      return <TreeDetailPanePage />;
    case 'form-sections':
      return <FormSectionsPage />;
    default:
      return null;
  }
}

export default function App() {
  const styles = useStyles();
  const [currentPage, setCurrentPage] = React.useState<PageKey>('grid-drawer');
  const currentMeta = pages.find(page => page.key === currentPage) ?? pages[0];

  return (
    <div className={styles.app}>
      <div className={styles.shell}>
        <header className={styles.hero}>
          <Title2>Fluent React 页面起手式</Title2>
          <Body1>
            这是一组可直接复制到 React + Fluent UI React v9 项目中的后台页面模板。
          </Body1>
        </header>

        <div className={styles.switcher}>
          {pages.map(page => (
            <Button
              key={page.key}
              appearance={page.key === currentPage ? 'primary' : 'secondary'}
              onClick={() => setCurrentPage(page.key)}
            >
              {page.label}
            </Button>
          ))}
        </div>

        <Body1>{currentMeta.description}</Body1>
        <Divider />

        <main className={styles.pageHost}>{renderPage(currentPage)}</main>
      </div>
    </div>
  );
}
