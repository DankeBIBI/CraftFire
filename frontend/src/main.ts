import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import "./assets/styles/main.css";

/**
 * CraftFire 前端应用入口。
 * 初始化 Vue 3 应用、Pinia 状态管理和全局样式。
 */
const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.mount("#app");
