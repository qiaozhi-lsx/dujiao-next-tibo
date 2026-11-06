// 从 /public/config 响应推算服务器当前时间(毫秒时间戳)。
//
// 该接口的 Cache-Control 为 `public, max-age=0, s-maxage=10`:浏览器不缓存,
// 但 CDN 边缘会缓存 10s 并在响应里带 Age 头(单位秒)。响应体里的 server_time
// 是源站生成响应那一刻的时间,所以要把 Age 加回去,再加上半个往返时延。

export interface EstimateServerNowInput {
    /** 响应体中的 server_time(毫秒时间戳);缺失时无法估算 */
    serverTime?: number
    /** 响应头 Age(秒);缺失或非法按 0 处理 */
    ageHeader?: string | null
    /** 发出请求时的客户端时间(毫秒) */
    requestTime: number
    /** 收到响应时的客户端时间(毫秒) */
    responseTime: number
}

/**
 * 估算收到响应那一刻的服务器时间;server_time 缺失返回 null,调用方不应更新偏移量。
 */
export const estimateServerNow = (input: EstimateServerNowInput): number | null => {
    const { serverTime, ageHeader, requestTime, responseTime } = input
    if (typeof serverTime !== 'number' || !Number.isFinite(serverTime)) return null
    const ageSeconds = Number(ageHeader || 0)
    const ageMs = Number.isFinite(ageSeconds) && ageSeconds > 0 ? ageSeconds * 1000 : 0
    return serverTime + ageMs + (responseTime - requestTime) / 2
}
