package player

import (
	"CraftFire/backend/internal/utils"
)

// InterpolationState 用于服务端对玩家位置进行插值计算的状态。
// 在低 Tick Rate (8 TPS) 下，通过插值平滑玩家运动。
type InterpolationState struct {
	PreviousPosition utils.Vec3
	TargetPosition   utils.Vec3
	PreviousRotation utils.Rotation
	TargetRotation   utils.Rotation
	StartTime        int64 // 插值开始时间（毫秒）
	Duration         int64 // 插值持续时间（毫秒），默认 125ms
}

// NewInterpolationState 创建一个新的插值状态实例。
//
// 参数：
//   - currentPos: 当前位置
//   - targetPos: 目标位置
//   - durationMs: 插值持续时间（毫秒）
func NewInterpolationState(currentPos, targetPos utils.Vec3, durationMs int64) *InterpolationState {
	return &InterpolationState{
		PreviousPosition: currentPos,
		TargetPosition:   targetPos,
		Duration:         durationMs,
	}
}

// Interpolate 根据当前时间计算插值后的位置。
//
// 参数：
//   - currentTimeMs: 当前时间戳（毫秒）
//
// 返回值：插值后的位置向量
func (is *InterpolationState) Interpolate(currentTimeMs int64) utils.Vec3 {
	elapsed := currentTimeMs - is.StartTime
	if elapsed >= is.Duration {
		return is.TargetPosition
	}

	t := float64(elapsed) / float64(is.Duration)
	// 使用三次平滑插值 (smoothstep) 提高视觉效果
	t = t * t * (3 - 2*t)

	return utils.Lerp(is.PreviousPosition, is.TargetPosition, t)
}
