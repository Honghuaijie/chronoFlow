export function requiredMessage(label: string): string {
  return `请输入${label}`
}

export function selectMessage(label: string): string {
  return `请选择${label}`
}

export function isPositiveInteger(value: number): boolean {
  return Number.isInteger(value) && value > 0
}
