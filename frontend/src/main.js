import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'sweetalert2/dist/sweetalert2.min.css'
import App from './App.vue'
import store from './store'
import { ensureMobileInteractions } from '@/helpers/ensureMobileInteractions'
import './styles.css'

// Точка входа: Vue 3 + Element Plus + Vuex (модули auth и tickets).
createApp(App).use(store).use(ElementPlus).mount('#app')
ensureMobileInteractions()
