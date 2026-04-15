/**
 * 观察性日志页面（audit-log / telemetry-log）共用的 ElDatePicker shortcuts。
 *
 * 设计取舍：
 *  - 6 个常用窗口覆盖绝大多数排障场景（小时级 / 天级），避免超长下拉列表；
 *  - 返回 [Date, Date]，ElDatePicker 根据父组件 `value-format` 自动序列化；
 *  - "今天" / "昨天" 按本地 0 点切片，便于跟运维聊"今天上午那一次" 时对齐；
 *  - 最近 N 对齐 `Date.now() - N`，不是"整点对齐"，保证窗口右端始终包含当前请求。
 *
 * 不在这里暴露"最近 90 天"：日志表按索引 + LIMIT 分页返回，90 天的跨度通常应
 * 由后台聚合接口来做，而不是把前端查询页打满。
 */

export type ObservabilityShortcut = {
  text: string
  value: () => [Date, Date]
}

const MS_PER_HOUR = 3600 * 1000
const MS_PER_DAY = 24 * MS_PER_HOUR

/** 从今天本地 0:00 构造一个 Date。保留原对象不被外部改写。 */
function startOfToday(): Date {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  return d
}

/** 以 start 为基础，+offsetDays 天 +(-1)ms 取得"前一天 23:59:59.999"或"当天 23:59:59.999"。 */
function endOfDay(start: Date): Date {
  const d = new Date(start.getTime())
  d.setHours(23, 59, 59, 999)
  return d
}

export const observabilityTimeShortcuts: ObservabilityShortcut[] = [
  {
    text: '最近 1 小时',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - MS_PER_HOUR)
      return [start, end]
    }
  },
  {
    text: '最近 24 小时',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - MS_PER_DAY)
      return [start, end]
    }
  },
  {
    text: '最近 7 天',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 7 * MS_PER_DAY)
      return [start, end]
    }
  },
  {
    text: '最近 30 天',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 30 * MS_PER_DAY)
      return [start, end]
    }
  },
  {
    text: '今天',
    value: () => {
      const start = startOfToday()
      return [start, endOfDay(start)]
    }
  },
  {
    text: '昨天',
    value: () => {
      const today0 = startOfToday()
      const start = new Date(today0.getTime() - MS_PER_DAY)
      return [start, endOfDay(start)]
    }
  }
]
