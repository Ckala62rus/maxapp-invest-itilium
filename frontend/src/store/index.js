import { createStore } from 'vuex'
import auth from '@/store/modules/auth'
import tickets from '@/store/modules/tickets'

/** Корневой store: только модули auth и tickets, без глобального дублирования сущностей. */
export default createStore({
  state: {},
  getters: {},
  mutations: {},
  actions: {},
  modules: {
    auth,
    tickets
  }
})
