package utils

import "math"

// Vec3 三维向量结构体，用于表示位置、速度等。
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Rotation 欧拉角旋转结构体。
type Rotation struct {
	Pitch float64 `json:"pitch"` // 俯仰角
	Yaw   float64 `json:"yaw"`   // 偏航角
	Roll  float64 `json:"roll"`  // 翻滚角
}

// Distance 计算两个三维点之间的欧几里得距离。
//
// 参数：
//   - a, b: 两个三维坐标点
//
// 返回值：距离值（float64）
func Distance(a, b Vec3) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dz := b.Z - a.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// DistanceSquared 计算两点之间距离的平方（避免开方运算，用于比较）。
func DistanceSquared(a, b Vec3) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dz := b.Z - a.Z
	return dx*dx + dy*dy + dz*dz
}

// Lerp 在两个向量之间进行线性插值。
//
// 参数：
//   - a: 起点向量
//   - b: 终点向量
//   - t: 插值因子 (0.0 ~ 1.0)
//
// 返回值：插值后的向量
func Lerp(a, b Vec3, t float64) Vec3 {
	return Vec3{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
	}
}

// Clamp 将值限制在最小值和最大值之间。
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampInt 将整数值限制在最小值和最大值之间。
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
