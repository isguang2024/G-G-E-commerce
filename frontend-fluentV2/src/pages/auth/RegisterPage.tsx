import { useMemo, useState } from 'react'
import { Button, Field, Input, MessageBar, MessageBarBody, makeStyles } from '@fluentui/react-components'
import { AuthFooterLinks, AuthScaffold } from '@/pages/auth/AuthScaffold'

const useStyles = makeStyles({
  form: {
    display: 'grid',
    gap: '16px',
  },
  submit: {
    width: '100%',
  },
})

export function RegisterPage() {
  const styles = useStyles()
  const [account, setAccount] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const passwordMismatch = useMemo(
    () => Boolean(confirmPassword) && password !== confirmPassword,
    [confirmPassword, password],
  )

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!account.trim() || !email.trim() || !password || passwordMismatch) {
      return
    }

    setSubmitted(true)
  }

  return (
    <AuthScaffold
      title="创建账号"
      footer={<AuthFooterLinks items={[{ label: '返回登录', to: '/login' }, { label: '忘记密码', to: '/forgot-password' }]} />}
    >
      <form className={styles.form} onSubmit={handleSubmit}>
        <Field label="账号">
          <Input placeholder="请输入账号" value={account} onChange={(_, data) => setAccount(data.value)} />
        </Field>

        <Field label="邮箱">
          <Input placeholder="name@example.com" value={email} onChange={(_, data) => setEmail(data.value)} />
        </Field>

        <Field label="密码">
          <Input
            placeholder="请输入密码"
            type="password"
            value={password}
            onChange={(_, data) => setPassword(data.value)}
          />
        </Field>

        <Field label="确认密码" validationState={passwordMismatch ? 'error' : undefined} validationMessage={passwordMismatch ? '两次输入的密码不一致' : undefined}>
          <Input
            placeholder="请再次输入密码"
            type="password"
            value={confirmPassword}
            onChange={(_, data) => setConfirmPassword(data.value)}
          />
        </Field>

        {submitted ? (
          <MessageBar>
            <MessageBarBody>账号已创建。当前为本地示例页，请直接返回登录。</MessageBarBody>
          </MessageBar>
        ) : null}

        <Button
          appearance="primary"
          className={styles.submit}
          disabled={!account.trim() || !email.trim() || !password || passwordMismatch}
          type="submit"
        >
          注册
        </Button>
      </form>
    </AuthScaffold>
  )
}
