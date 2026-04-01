package profile

// CharacterRenderer 角色模型渲染数据组装器。
// 负责为前端个人页面的 3D 角色预览准备数据。
type CharacterRenderer struct{}

// NewCharacterRenderer 创建一个新的角色渲染器。
func NewCharacterRenderer() *CharacterRenderer {
	return &CharacterRenderer{}
}

// CharacterRenderData 角色渲染所需的数据。
type CharacterRenderData struct {
	ModelURL     string            `json:"modelUrl"`
	Animations   []string          `json:"animations"`
	Equipment    map[string]string `json:"equipment"`
	SkinSettings map[string]string `json:"skinSettings"`
}

// GetRenderData 获取角色渲染数据。
func (r *CharacterRenderer) GetRenderData(playerId string) (*CharacterRenderData, error) {
	return &CharacterRenderData{
		ModelURL:   "/models/player.glb",
		Animations: []string{"idle", "walk", "run", "attack"},
		Equipment: map[string]string{
			"weapon": "pistol",
			"armor":  "default",
		},
		SkinSettings: map[string]string{
			"color": "#FFFFFF",
		},
	}, nil
}
