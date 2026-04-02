import type { CSSProperties, ReactNode } from 'react'
import type { NavIconKey } from '@/shared/types/navigation'

const baseStyle: CSSProperties = {
  width: 18,
  height: 18,
  display: 'block',
}

function strokeIcon(children: ReactNode) {
  return (
    <svg viewBox="0 0 24 24" fill="none" style={baseStyle} stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      {children}
    </svg>
  )
}

export function AppIcon({ icon }: { icon: NavIconKey }) {
  switch (icon) {
    case 'home':
      return strokeIcon(<path d="M4 11.5 12 5l8 6.5V20H4z" />)
    case 'workspace':
      return strokeIcon(
        <>
          <rect x="4" y="5" width="7" height="6" rx="1.5" />
          <rect x="13" y="5" width="7" height="6" rx="1.5" />
          <rect x="4" y="13" width="16" height="6" rx="1.5" />
        </>,
      )
    case 'team':
      return strokeIcon(
        <>
          <circle cx="9" cy="9" r="2.5" />
          <circle cx="16.5" cy="10" r="2" />
          <path d="M5.5 18c.8-2.2 2.4-3.3 4.8-3.3S14.3 15.8 15 18" />
          <path d="M15.5 17c.4-1.4 1.5-2.2 3.2-2.2" />
        </>,
      )
    case 'message':
      return strokeIcon(
        <>
          <path d="M5 6.5h14a2 2 0 0 1 2 2V16a2 2 0 0 1-2 2H9l-4 3v-3H5a2 2 0 0 1-2-2V8.5a2 2 0 0 1 2-2Z" />
          <path d="M8 11h8M8 14h5" />
        </>,
      )
    case 'system':
      return strokeIcon(
        <>
          <circle cx="12" cy="12" r="3.2" />
          <path d="M12 4v2.2M12 17.8V20M4 12h2.2M17.8 12H20M6.3 6.3l1.6 1.6M16.1 16.1l1.6 1.6M6.3 17.7l1.6-1.6M16.1 7.9l1.6-1.6" />
        </>,
      )
    case 'menu':
      return strokeIcon(
        <>
          <path d="M5 7h14M5 12h14M5 17h14" />
          <path d="M4 7h0M4 12h0M4 17h0" />
        </>,
      )
    case 'page':
      return strokeIcon(
        <>
          <path d="M7 4h7l4 4v12H7z" />
          <path d="M14 4v4h4M9 12h6M9 16h6" />
        </>,
      )
    case 'role':
      return strokeIcon(
        <>
          <path d="M7 11a5 5 0 1 1 10 0c0 4-5 8-5 8s-5-4-5-8Z" />
          <path d="M12 8.5v5M9.5 11h5" />
        </>,
      )
    case 'user':
      return strokeIcon(
        <>
          <circle cx="12" cy="8.5" r="3" />
          <path d="M6 18c1.2-3 3.2-4.5 6-4.5s4.8 1.5 6 4.5" />
        </>,
      )
    case 'api':
      return strokeIcon(
        <>
          <rect x="4" y="7" width="6" height="10" rx="1.5" />
          <rect x="14" y="7" width="6" height="10" rx="1.5" />
          <path d="M10 12h4M7 7V4M17 20v-3" />
        </>,
      )
    case 'package':
      return strokeIcon(
        <>
          <path d="M4 8.5 12 4l8 4.5-8 4.5Z" />
          <path d="M4 8.5V16l8 4 8-4V8.5M12 13v7" />
        </>,
      )
    case 'inbox':
      return strokeIcon(
        <>
          <path d="M5 7h14l2 10H3Z" />
          <path d="M8 12h8l-1.5 3h-5Z" />
        </>,
      )
    case 'space':
      return strokeIcon(
        <>
          <path d="M4 7.5 12 4l8 3.5-8 3.5Z" />
          <path d="M4 12l8 3.5 8-3.5M4 16.5 12 20l8-3.5" />
        </>,
      )
    default:
      return strokeIcon(
        <>
          <rect x="5" y="6" width="14" height="12" rx="2" />
          <path d="M9 10h6M9 14h4" />
        </>,
      )
  }
}
