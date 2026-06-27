import dayjs from 'dayjs'

type CronSets = {
  seconds: Set<number>
  minutes: Set<number>
  hours: Set<number>
  days: Set<number>
  months: Set<number>
  weekdays: Set<number>
  dayRestricted: boolean
  weekdayRestricted: boolean
}

const WEEKDAY_LABELS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

function range(min: number, max: number): number[] {
  return Array.from({ length: max - min + 1 }, (_, index) => min + index)
}

function parseField(field: string, min: number, max: number, normalize?: (value: number) => number): Set<number> | null {
  const values = new Set<number>()
  const normalizedField = field.trim()
  if (!normalizedField || normalizedField === '*' || normalizedField === '?') {
    return new Set(range(min, max).map((value) => (normalize ? normalize(value) : value)))
  }

  for (const part of normalizedField.split(',')) {
    const token = part.trim()
    if (!token) {
      return null
    }
    const [base, stepText] = token.split('/')
    const step = stepText ? Number(stepText) : 1
    if (!Number.isInteger(step) || step <= 0) {
      return null
    }

    let start = min
    let end = max
    if (base && base !== '*' && base !== '?') {
      if (base.includes('-')) {
        const [left, right] = base.split('-').map(Number)
        if (!Number.isInteger(left) || !Number.isInteger(right)) {
          return null
        }
        start = left
        end = right
      } else {
        const exact = Number(base)
        if (!Number.isInteger(exact)) {
          return null
        }
        if (stepText) {
          start = exact
          end = max
        } else {
          start = exact
          end = exact
        }
      }
    }

    if (start > end) {
      return null
    }
    for (let value = start; value <= end; value += step) {
      const next = normalize ? normalize(value) : value
      if (next < min || next > max) {
        return null
      }
      values.add(next)
    }
  }

  return values
}

function parseCron(expr: string): CronSets | null {
  const fields = expr.trim().split(/\s+/)
  if (fields.length !== 6) {
    return null
  }
  const [secondField, minuteField, hourField, dayField, monthField, weekdayField] = fields
  const seconds = parseField(secondField, 0, 59)
  const minutes = parseField(minuteField, 0, 59)
  const hours = parseField(hourField, 0, 23)
  const days = parseField(dayField, 1, 31)
  const months = parseField(monthField, 1, 12)
  const weekdays = parseField(weekdayField, 0, 7, (value) => (value === 7 ? 0 : value))
  if (!seconds || !minutes || !hours || !days || !months || !weekdays) {
    return null
  }
  return {
    seconds,
    minutes,
    hours,
    days,
    months,
    weekdays,
    dayRestricted: dayField !== '*' && dayField !== '?',
    weekdayRestricted: weekdayField !== '*' && weekdayField !== '?',
  }
}

function matchesDate(date: Date, cron: CronSets): boolean {
  const month = date.getMonth() + 1
  const day = date.getDate()
  const weekday = date.getDay()
  if (!cron.months.has(month) || !cron.hours.has(date.getHours()) || !cron.minutes.has(date.getMinutes())) {
    return false
  }

  const dayMatches = cron.days.has(day)
  const weekdayMatches = cron.weekdays.has(weekday)
  if (cron.dayRestricted && cron.weekdayRestricted) {
    return dayMatches || weekdayMatches
  }
  return dayMatches && weekdayMatches
}

export function getNextRunTime(expr: string, from = new Date()): Date | null {
  const cron = parseCron(expr)
  if (!cron) {
    return null
  }

  const sortedSeconds = Array.from(cron.seconds).sort((a, b) => a - b)
  let cursor = new Date(from.getTime() + 1000)
  cursor.setMilliseconds(0)

  const maxMinutes = 370 * 24 * 60
  for (let index = 0; index < maxMinutes; index += 1) {
    const minuteCursor = new Date(cursor)
    minuteCursor.setSeconds(0, 0)
    if (matchesDate(minuteCursor, cron)) {
      const minSecond = index === 0 ? cursor.getSeconds() : 0
      const second = sortedSeconds.find((item) => item >= minSecond)
      if (second !== undefined) {
        const candidate = new Date(minuteCursor)
        candidate.setSeconds(second, 0)
        if (candidate.getTime() > from.getTime()) {
          return candidate
        }
      }
    }
    cursor = new Date(minuteCursor.getTime() + 60 * 1000)
  }
  return null
}

export function getNextRunTimes(expr: string, count = 5, from = new Date()): Date[] {
  const times: Date[] = []
  let cursor = from
  for (let index = 0; index < count; index += 1) {
    const next = getNextRunTime(expr, cursor)
    if (!next) {
      break
    }
    times.push(next)
    cursor = next
  }
  return times
}

export function formatNextRunTime(expr: string): string {
  const next = getNextRunTime(expr)
  return next ? dayjs(next).format('YYYY-MM-DD HH:mm:ss') : '无法计算'
}

export function formatNextRunTimes(expr: string, count = 5): string[] {
  return getNextRunTimes(expr, count).map((item) => dayjs(item).format('YYYY-MM-DD HH:mm:ss'))
}

export function describeCron(expr: string): string {
  const fields = expr.trim().split(/\s+/)
  if (fields.length !== 6) {
    return 'Cron 表达式需为 6 段'
  }
  const [second, minute, hour, day, month, weekday] = fields
  if (second === '0' && minute.startsWith('*/') && hour === '*' && day === '*' && month === '*' && weekday === '*') {
    return `每 ${minute.slice(2)} 分钟`
  }
  if (hour.startsWith('*/') && day === '*' && month === '*' && weekday === '*') {
    return `每 ${hour.slice(2)} 小时的第 ${minute.padStart(2, '0')} 分钟`
  }
  if (day === '*' && month === '*' && weekday === '*') {
    return `每天 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}:${second.padStart(2, '0')}`
  }
  if (day === '*' && month === '*' && /^[0-7]$/.test(weekday)) {
    return `${WEEKDAY_LABELS[Number(weekday) % 7]} ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}:${second.padStart(2, '0')}`
  }
  if (month === '*' && weekday === '*') {
    return `每月 ${day} 日 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}:${second.padStart(2, '0')}`
  }
  return expr
}
