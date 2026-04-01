import type { PropsWithChildren, ReactNode } from 'react'
import { Body1, Caption1, Divider, makeStyles, tokens, Title2 } from '@fluentui/react-components'
import { Link as RouterLink } from 'react-router-dom'
import { AppLogo } from '@/shared/ui/AppLogo'

const useStyles = makeStyles({
  page: {
    position: 'relative',
    isolation: 'isolate',
    overflow: 'hidden',
    minHeight: '100vh',
    display: 'grid',
    placeItems: 'center',
    padding: '24px',
    backgroundColor: tokens.colorNeutralBackground1,
    backgroundImage: [
      `radial-gradient(circle at 18% 18%, ${tokens.colorBrandBackground2} 0%, transparent 24%)`,
      `radial-gradient(circle at 80% 12%, ${tokens.colorBrandBackground2} 0%, transparent 14%)`,
      `radial-gradient(circle at 92% 28%, ${tokens.colorBrandBackground2} 0%, transparent 12%)`,
      `radial-gradient(circle at 86% 72%, ${tokens.colorBrandBackground2} 0%, transparent 10%)`,
      `linear-gradient(180deg, ${tokens.colorNeutralBackground3} 0%, ${tokens.colorNeutralBackground1} 48%, ${tokens.colorNeutralBackground2} 100%)`,
    ].join(', '),
    backgroundRepeat: 'no-repeat',
    backgroundSize: '100% 100%',
    '::before': {
      content: '""',
      position: 'absolute',
      inset: '0',
      zIndex: -1,
      backgroundImage: `linear-gradient(135deg, ${tokens.colorBrandBackground} 0%, transparent 38%)`,
      opacity: 0.08,
      pointerEvents: 'none',
    },
    '::after': {
      content: '""',
      position: 'absolute',
      inset: '8% -8% auto auto',
      width: '36vw',
      minWidth: '280px',
      aspectRatio: '1 / 1',
      borderRadius: '50%',
      backgroundImage: [
        `radial-gradient(circle at 26% 28%, ${tokens.colorBrandBackground2} 0%, transparent 24%)`,
        `radial-gradient(circle at 72% 22%, ${tokens.colorBrandBackground} 0%, transparent 20%)`,
        `radial-gradient(circle at 64% 70%, ${tokens.colorBrandBackground2} 0%, transparent 18%)`,
      ].join(', '),
      opacity: 0.16,
      filter: 'blur(22px)',
      zIndex: -1,
      pointerEvents: 'none',
    },
  },
  card: {
    width: 'min(440px, 100%)',
    display: 'grid',
    gap: '20px',
    padding: '28px 32px 24px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    boxShadow: tokens.shadow16,
    '@media (max-width: 480px)': {
      padding: '24px 20px 20px',
    },
  },
  header: {
    display: 'grid',
    gap: '14px',
  },
  brand: {
    display: 'inline-flex',
    alignSelf: 'start',
  },
  heading: {
    display: 'grid',
    gap: '6px',
  },
  form: {
    display: 'grid',
    gap: '16px',
  },
  footer: {
    display: 'grid',
    gap: '12px',
  },
  footerLinks: {
    display: 'flex',
    alignItems: 'center',
    gap: '14px',
    flexWrap: 'wrap',
  },
  link: {
    color: tokens.colorBrandForegroundLink,
    textDecorationLine: 'none',
    ':hover': {
      color: tokens.colorBrandForegroundLinkHover,
      textDecorationLine: 'underline',
    },
  },
})

export function AuthScaffold({
  title,
  description,
  children,
  footer,
}: PropsWithChildren<{
  title: string
  description?: string
  footer?: ReactNode
}>) {
  const styles = useStyles()

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles.brand}>
            <AppLogo />
          </div>
          <div className={styles.heading}>
            <Title2>{title}</Title2>
            {description ? <Body1>{description}</Body1> : null}
          </div>
        </div>

        <div className={styles.form}>{children}</div>

        {footer ? (
          <div className={styles.footer}>
            <Divider />
            {footer}
          </div>
        ) : null}
      </div>
    </div>
  )
}

export function AuthFooterLinks({
  items,
}: {
  items: Array<{ label: string; to: string }>
}) {
  const styles = useStyles()

  return (
    <div className={styles.footerLinks}>
      {items.map((item) => (
        <Caption1 as="span" key={item.to}>
          <RouterLink className={styles.link} to={item.to}>
            {item.label}
          </RouterLink>
        </Caption1>
      ))}
    </div>
  )
}
