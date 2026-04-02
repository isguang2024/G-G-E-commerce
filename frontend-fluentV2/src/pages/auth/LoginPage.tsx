import { useMemo, useState } from 'react'
import { Button, Checkbox, Field, Input, MessageBar, MessageBarBody, Spinner, makeStyles, tokens } from '@fluentui/react-components'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/auth.store'
import { useLoginMutation } from '@/features/auth/auth.service'
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
  const [searchParams] = useSearchParams()
  const setAuthenticated = useAuthStore((state) => state.setAuthenticated)
  const loginMutation = useLoginMutation()
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')
  const [keepSignedIn, setKeepSignedIn] = useState(true)

  const redirectTo = useMemo(() => {
    const locationState = location.state as { from?: string } | null
    const redirect = searchParams.get('redirect') || locationState?.from
    return redirect || appConfig.defaultRoute
  }, [location.state, searchParams])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!account.trim() || !password) {
      return
    }

    const result = await loginMutation.mutateAsync({
      username: account.trim(),
      password,
    })
    setAuthenticated({
      session: result.session,
      currentUser: result.loginUser,
      rememberMe: keepSignedIn,
    })
    navigate(redirectTo, { replace: true })
  }

  return (
    <AuthScaffold
      title="登录工作台"
      footer={<AuthFooterLinks items={[{ label: '创建账号', to: '/register' }]} />}
    >
      <form className={styles.form} onSubmit={(event) => void handleSubmit(event)}>
        {loginMutation.isError ? (
          <MessageBar intent="error">
            <MessageBarBody>{loginMutation.error.message || '登录失败，请检查账号或密码。'}</MessageBarBody>
          </MessageBar>
        ) : null}

        <Field label="账号">
          <Input placeholder="请输入用户名" value={account} onChange={(_, data) => setAccount(data.value)} />
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

        <Button appearance="primary" disabled={!account.trim() || !password || loginMutation.isPending} style={{ width: '100%' }} type="submit">
          {loginMutation.isPending ? <Spinner size="tiny" labelPosition="after">登录中</Spinner> : '登录'}
        </Button>
      </form>
    </AuthScaffold>
  )
}
