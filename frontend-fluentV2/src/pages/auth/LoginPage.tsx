import { useState } from 'react'
import { Button, Checkbox, Field, Input, makeStyles, tokens } from '@fluentui/react-components'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/features/session/auth.store'
import { AuthFooterLinks, AuthScaffold } from '@/pages/auth/AuthScaffold'
import { appConfig } from '@/shared/config/app-config'

const useStyles = makeStyles({
  form: {
    display: 'grid',
    gap: '16px',
  },
  row: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: '12px',
    flexWrap: 'wrap',
  },
  linkButton: {
    color: tokens.colorBrandForegroundLink,
  },
})

export function LoginPage() {
  const styles = useStyles()
  const navigate = useNavigate()
  const location = useLocation()
  const signIn = useAuthStore((state) => state.signIn)
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')
  const [keepSignedIn, setKeepSignedIn] = useState(true)

  const redirectTo =
    (location.state as { from?: string } | null)?.from || appConfig.defaultRoute

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!account.trim() || !password) {
      return
    }

    signIn({ account, rememberMe: keepSignedIn })
    navigate(redirectTo, { replace: true })
  }

  return (
    <AuthScaffold
      title="登录工作台"
      footer={<AuthFooterLinks items={[{ label: '创建账号', to: '/register' }]} />}
    >
      <form className={styles.form} onSubmit={handleSubmit}>
        <Field label="账号">
          <Input placeholder="name@example.com" value={account} onChange={(_, data) => setAccount(data.value)} />
        </Field>

        <Field label="密码">
          <Input
            placeholder="请输入密码"
            type="password"
            value={password}
            onChange={(_, data) => setPassword(data.value)}
          />
        </Field>

        <div className={styles.row}>
          <Checkbox
            checked={keepSignedIn}
            label="保持登录状态"
            onChange={(_, data) => setKeepSignedIn(Boolean(data.checked))}
          />
          <Button appearance="transparent" className={styles.linkButton} type="button" onClick={() => navigate('/forgot-password')}>
            忘记密码
          </Button>
        </div>

        <Button appearance="primary" disabled={!account.trim() || !password} style={{ width: '100%' }} type="submit">
          登录
        </Button>
      </form>
    </AuthScaffold>
  )
}
