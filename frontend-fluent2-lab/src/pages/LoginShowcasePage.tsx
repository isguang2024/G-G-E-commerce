import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Checkbox,
  Field,
  Input,
  Link,
  Subtitle2,
  Title2,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: {
    minHeight: '720px',
    display: 'grid',
    gridTemplateColumns: '1.1fr 0.9fr',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow16,
    backgroundColor: tokens.colorNeutralBackground1,
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  intro: {
    display: 'grid',
    alignContent: 'space-between',
    gap: '20px',
    padding: '36px',
    background:
      `radial-gradient(circle at top left, rgba(15,108,189,0.18), transparent 38%), linear-gradient(180deg, ${tokens.colorBrandBackground2} 0%, ${tokens.colorNeutralBackground3} 100%)`,
  },
  introHeader: {
    display: 'grid',
    gap: '12px',
    maxWidth: '560px',
  },
  introGrid: {
    display: 'grid',
    gap: '14px',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    '@media (max-width: 700px)': {
      gridTemplateColumns: '1fr',
    },
  },
  introCard: {
    display: 'grid',
    gap: '8px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  introBanner: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: 'rgba(255,255,255,0.10)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    backdropFilter: 'blur(8px)',
  },
  introRow: {
    display: 'grid',
    gridTemplateColumns: '1.1fr 0.9fr',
    gap: '14px',
    '@media (max-width: 700px)': {
      gridTemplateColumns: '1fr',
    },
  },
  formWrap: {
    display: 'grid',
    placeItems: 'center',
    padding: '36px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.02) 0%, rgba(15,108,189,0.06) 100%)',
  },
  formCard: {
    width: 'min(420px, 100%)',
    display: 'grid',
    gap: '18px',
    padding: '28px',
    borderRadius: tokens.borderRadiusXLarge,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow4,
  },
  formHeader: {
    display: 'grid',
    gap: '8px',
  },
  formFields: {
    display: 'grid',
    gap: '14px',
  },
  formMeta: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
    flexWrap: 'wrap',
  },
  footerText: {
    color: tokens.colorNeutralForeground3,
  },
});

export function LoginShowcasePage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <section className={styles.intro}>
        <div className={styles.introBanner}>
          <Badge appearance="filled" color="brand">
            Sign-in test
          </Badge>
          <div className={styles.introRow}>
            <div className={styles.introHeader}>
              <Title2>Fluent 2 登录页测试样式</Title2>
              <Body1>
                这一页验证后台产品常见的双栏登录结构。左侧负责解释产品上下文和可信感，右侧只承接登录表单，不把低频内容塞进表单区域。
              </Body1>
            </div>
            <div className={styles.introHeader}>
              <Subtitle2>在登录前就建立方向感。</Subtitle2>
              <Caption1>清晰、克制、可靠，先让用户知道自己将进入哪里，再让表单接管输入。</Caption1>
            </div>
          </div>
        </div>

        <div className={styles.introGrid}>
          <article className={styles.introCard}>
            <Body1Strong>聚焦主动作</Body1Strong>
            <Caption1>首屏只保留登录和找回密码，不把注册、公告、营销卡片同时堆进来。</Caption1>
          </article>
          <article className={styles.introCard}>
            <Body1Strong>建立可信感</Body1Strong>
            <Caption1>用简洁的功能摘要、状态说明和环境信息告诉用户自己将进入哪里。</Caption1>
          </article>
          <article className={styles.introCard}>
            <Body1Strong>适合企业后台</Body1Strong>
            <Caption1>克制背景、清晰层级和轻量品牌表达，更适合平台、工作台和协作产品入口。</Caption1>
          </article>
          <article className={styles.introCard}>
            <Body1Strong>支持后续扩展</Body1Strong>
            <Caption1>后续可平滑加上租户选择、二次验证、SSO 入口和安全提示。</Caption1>
          </article>
        </div>
      </section>

      <section className={styles.formWrap}>
        <form className={styles.formCard}>
          <div className={styles.formHeader}>
            <Body1Strong>登录工作台</Body1Strong>
            <Caption1>输入组织账号后进入 Fluent 2 React 实验场。</Caption1>
          </div>

          <div className={styles.formFields}>
            <Field label="账号">
              <Input placeholder="name@contoso.com" />
            </Field>

            <Field label="密码">
              <Input type="password" placeholder="请输入密码" />
            </Field>
          </div>

          <div className={styles.formMeta}>
            <Checkbox label="保持登录状态" defaultChecked />
            <Link href="#">忘记密码</Link>
          </div>

          <Button appearance="primary">登录</Button>
          <Button appearance="secondary">使用 Microsoft 账号继续</Button>

          <Caption1 className={styles.footerText}>
            当前页面是视觉测试样式，不接真实认证接口。
          </Caption1>
        </form>
      </section>
    </div>
  );
}
