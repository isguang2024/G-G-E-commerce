import { useState } from 'react'
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

export function ForgotPasswordPage() {
  const styles = useStyles()
  const [identifier, setIdentifier] = useState('')
  const [submitted, setSubmitted] = useState(false)

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!identifier.trim()) {
      return
    }

    setSubmitted(true)
  }

  return (
    <AuthScaffold
      title="找回密码"
      description="输入账号或邮箱，我们会发送重置方式。"
      footer={<AuthFooterLinks items={[{ label: '返回登录', to: '/login' }, { label: '创建账号', to: '/register' }]} />}
    >
      <form className={styles.form} onSubmit={handleSubmit}>
        <Field label="账号或邮箱">
          <Input placeholder="name@example.com" value={identifier} onChange={(_, data) => setIdentifier(data.value)} />
        </Field>

        {submitted ? (
          <MessageBar>
            <MessageBarBody>重置说明已发送。当前为本地示例页，不会触发真实邮件。</MessageBarBody>
          </MessageBar>
        ) : null}

        <Button appearance="primary" className={styles.submit} disabled={!identifier.trim()} type="submit">
          发送重置链接
        </Button>
      </form>
    </AuthScaffold>
  )
}
