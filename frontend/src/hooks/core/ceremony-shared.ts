import { festivalConfigList } from '@/config/modules/festival'
import type { FestivalConfig } from '@/types/config'

export function getCurrentDateString(date: Date = new Date()): string {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function isDateInFestivalRange(
  currentDate: string,
  festivalDate: string,
  festivalEndDate?: string
): boolean {
  if (!festivalEndDate) {
    return currentDate === festivalDate
  }

  const current = new Date(currentDate)
  const start = new Date(festivalDate)
  const end = new Date(festivalEndDate)
  return current >= start && current <= end
}

export function resolveCurrentFestivalData(currentDate: string): FestivalConfig | undefined {
  return festivalConfigList.find((item) =>
    isDateInFestivalRange(currentDate, item.date, item.endDate)
  )
}
