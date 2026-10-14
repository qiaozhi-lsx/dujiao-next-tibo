export type CountdownTone = 'safe' | 'warn' | 'late' | 'critical'

/** 倒计时颜色分档：>300s 绿 / ≤300s 橙 / ≤180s 深橙 / ≤60s 红。 */
export const countdownTone = (remainSeconds: number): CountdownTone => {
  if (remainSeconds <= 60) return 'critical'
  if (remainSeconds <= 180) return 'late'
  if (remainSeconds <= 300) return 'warn'
  return 'safe'
}
