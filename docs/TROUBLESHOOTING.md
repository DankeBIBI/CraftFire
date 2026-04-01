# CraftFire 常见问题排查

## 构建问题

### `wails doctor` 报错缺少 WebView2

**Windows**: 下载安装 [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)。

### Go 版本不兼容

```
go: go.mod requires go >= 1.19
```

升级 Go 到 1.19 或更高版本：[https://go.dev/dl/](https://go.dev/dl/)

### npm install 失败

```bash
# 清理缓存后重试
cd frontend
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### Wails 构建找不到前端

确保 `wails.json` 中 `frontend:dir` 指向正确路径：

```json
{
	"frontend:dir": "./frontend"
}
```

## 运行时问题

### 应用启动后白屏

1. 检查 `frontend/dist/` 是否已构建（`wails dev` 会自动处理）
2. 检查浏览器控制台（F12）是否有 JS 错误
3. 确认 Vue 组件无编译错误：`cd frontend && npm run type-check`

### WebSocket 连接失败

**症状**: 加入房间后无法看到其他玩家

**排查**:

1. 确认双方在同一局域网
2. 检查防火墙是否放行了对应端口（6 位房间号端口）
3. 确认房间创建者的应用仍在运行
4. 查看 Go 日志输出是否有连接错误

```bash
# Windows 放行端口示例
netsh advfirewall firewall add rule name="CraftFire" dir=in action=allow protocol=TCP localport=100000-999999
```

### 局域网发现找不到房间

**排查**:

1. 确认双方在同一子网
2. 检查 UDP 9999 端口是否被防火墙阻止
3. 路由器是否允许 UDP 广播
4. 尝试手动输入房间号 + IP 加入

### 帧率低 / 画面卡顿

1. **降低渲染距离**: 设置面板 → 视频 → 渲染距离调低
2. **关闭阴影**: 设置面板 → 视频 → 阴影质量设为关闭
3. **检查 GPU 驱动**: 更新显卡驱动到最新版本
4. **检查分块数**: 过多加载的分块会增加内存压力

### 鼠标无法锁定 (Pointer Lock)

1. 点击游戏画面区域以激活 Pointer Lock
2. 按 `ESC` 可解锁鼠标
3. 部分安全软件可能阻止 Pointer Lock，检查是否有拦截提示

### 管理员面板无法登录

1. 确认你是房间创建者（仅房主可访问管理员面板）
2. 默认密码在 `config.yml` 的 `security.adminPassword` 字段
3. 密码区分大小写
4. Session 过期后需重新登录

## 开发调试

### 启用调试模式

在游戏中按 `F3` 可切换调试信息显示（FPS、坐标、玩家数）。

也可在浏览器 `localStorage` 中设置：

```javascript
localStorage.setItem("DEBUG_MODE", "true");
```

### 查看 WebSocket 消息

浏览器开发工具 → Network → WS → 选中 WebSocket 连接可查看实时消息。

### Go 后端日志

调整 `config.yml` 中的日志级别：

```yaml
logging:
  level: "DEBUG" # DEBUG | INFO | WARN | ERROR
  format: "text"
  output: "console"
```

### 前端类型检查

```bash
cd frontend
npm run type-check
```

### 后端测试

```bash
cd backend
go test ./... -v
```

### 性能基准

```bash
cd backend
go test -bench ./... -benchmem
```

## 网络问题

### 延迟过高

- 确保使用有线网络（WiFi 可能增加 10-50ms 延迟）
- 检查局域网是否有大量其他流量
- 8 TPS 设计下正常延迟应 < 100ms

### 玩家位置抖动

这通常是插值不充分导致的。确认：

1. `InterpolationEngine` 正确运行
2. 客户端收到的 `player_state_sync` 频率稳定
3. 检查网络丢包率

## 其他

### 模型导入失败

1. 确认文件格式为 GLTF/GLB/OBJ/FBX
2. 单个模型文件建议 < 10MB
3. 检查模型文件是否损坏（尝试用 Blender 打开）
4. GLB 格式兼容性最好，推荐优先使用

### 世界数据丢失

- 世界数据每 300 秒自动保存（可在 `config.yml` 调整 `world.persistenceInterval`）
- 突然关闭应用可能丢失最近的修改
- 数据存储在 `~/.CraftFire/worlds/` 目录

### 还是无法解决？

1. 查看项目 Issues 是否有类似问题
2. 提交新 Issue 并附上：
   - 操作系统版本
   - Go / Node.js / Wails 版本
   - 错误日志（Go 控制台 + 浏览器控制台）
   - 复现步骤
