import { createApp } from 'vue'
import App from './App.vue'
import store from './store'
import './styles.css'

// The app uses a shared Vuex store so screen state can move out
// of the prototype component into domain modules step by step.
createApp(App).use(store).mount('#app')
