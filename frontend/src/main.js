import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import store from './store'
import './styles.css'
import 'sweetalert2/dist/sweetalert2.min.css'

// Точка входа: Vue 3 + Element Plus + Vuex (модули auth и tickets).
createApp(App).use(store).use(ElementPlus).mount('#app')
