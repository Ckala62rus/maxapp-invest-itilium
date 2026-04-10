import { createStore } from 'vuex'
import auth from '@/store/modules/auth'

// The root store aggregates domain modules and shared state.
export default createStore({
  state: {},
  getters: {},
  mutations: {},
  actions: {},
  modules: {
    auth
  }
})
