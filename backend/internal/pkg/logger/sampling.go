package logger

import "time"

// samplingTick 是采样窗口。1 秒的窗口能覆盖常见高并发 burst，
// 又不会把罕见错误日志压到下游看不到。写在独立文件里便于测试替换。
const samplingTick = time.Second
