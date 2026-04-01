import { useQuery } from '@tanstack/react-query'
import { currentUserMock } from '@/shared/mocks/session.mock'
import { withDelay } from '@/shared/lib/delay'

async function fetchCurrentUser() {
  return withDelay(currentUserMock)
}

export function useCurrentUserQuery() {
  return useQuery({
    queryKey: ['session', 'current-user'],
    queryFn: fetchCurrentUser,
  })
}
