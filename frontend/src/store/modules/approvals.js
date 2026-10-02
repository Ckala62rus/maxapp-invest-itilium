import approvalsApi from '@/api/approvals'

function normalizeApprovalError(error) {
  const backendMessage = String(error?.response?.data?.message || '').trim()
  if (backendMessage) {
    return backendMessage
  }

  const status = error?.response?.status
  if (status === 404) {
    return 'Согласование не найдено в ITILIUM.'
  }
  if (status >= 500) {
    return 'ITILIUM временно недоступен. Повторите попытку позже.'
  }

  return String(error?.message || 'Не удалось выполнить запрос к ITILIUM.').trim()
}

const state = {
  items: [],
  selectedApproval: null,
  isLoadingList: false,
  isLoadingDetails: false,
  isSubmittingVote: false,
  listError: [],
  detailsError: []
}

export const mutationTypes = {
  loadListStart: '[approvals] loadListStart',
  loadListSuccess: '[approvals] loadListSuccess',
  loadListFail: '[approvals] loadListFail',
  loadDetailsStart: '[approvals] loadDetailsStart',
  loadDetailsSuccess: '[approvals] loadDetailsSuccess',
  loadDetailsFail: '[approvals] loadDetailsFail',
  voteStart: '[approvals] voteStart',
  voteSuccess: '[approvals] voteSuccess',
  voteFail: '[approvals] voteFail'
}

export const actionTypes = {
  loadList: '[approvals] loadList',
  loadDetails: '[approvals] loadDetails',
  vote: '[approvals] vote'
}

export const getterTypes = {
  items: '[approvals] items',
  selectedApproval: '[approvals] selectedApproval',
  isLoadingList: '[approvals] isLoadingList',
  isLoadingDetails: '[approvals] isLoadingDetails',
  isSubmittingVote: '[approvals] isSubmittingVote',
  listError: '[approvals] listError',
  detailsError: '[approvals] detailsError'
}

const getters = {
  [getterTypes.items]: (localState) => localState.items,
  [getterTypes.selectedApproval]: (localState) => localState.selectedApproval,
  [getterTypes.isLoadingList]: (localState) => localState.isLoadingList,
  [getterTypes.isLoadingDetails]: (localState) => localState.isLoadingDetails,
  [getterTypes.isSubmittingVote]: (localState) => localState.isSubmittingVote,
  [getterTypes.listError]: (localState) => localState.listError,
  [getterTypes.detailsError]: (localState) => localState.detailsError
}

const mutations = {
  [mutationTypes.loadListStart](localState) {
    localState.isLoadingList = true
    localState.listError = []
  },
  [mutationTypes.loadListSuccess](localState, items) {
    localState.isLoadingList = false
    localState.items = Array.isArray(items) ? items : []
  },
  [mutationTypes.loadListFail](localState, errors) {
    localState.isLoadingList = false
    localState.listError = errors
  },
  [mutationTypes.loadDetailsStart](localState) {
    localState.isLoadingDetails = true
    localState.selectedApproval = null
    localState.detailsError = []
  },
  [mutationTypes.loadDetailsSuccess](localState, approval) {
    localState.isLoadingDetails = false
    localState.selectedApproval = approval
  },
  [mutationTypes.loadDetailsFail](localState, errors) {
    localState.isLoadingDetails = false
    localState.selectedApproval = null
    localState.detailsError = errors
  },
  [mutationTypes.voteStart](localState) {
    localState.isSubmittingVote = true
    localState.detailsError = []
  },
  [mutationTypes.voteSuccess](localState) {
    localState.isSubmittingVote = false
  },
  [mutationTypes.voteFail](localState, errors) {
    localState.isSubmittingVote = false
    localState.detailsError = errors
  }
}

const actions = {
  [actionTypes.loadList](context) {
    context.commit(mutationTypes.loadListStart)
    return approvalsApi.listApprovals()
      .then((response) => {
        context.commit(mutationTypes.loadListSuccess, response?.data?.data || [])
        return response
      })
      .catch((error) => {
        context.commit(mutationTypes.loadListFail, [normalizeApprovalError(error)])
        return error
      })
  },
  [actionTypes.loadDetails](context, number) {
    context.commit(mutationTypes.loadDetailsStart)
    return approvalsApi.getApproval(number)
      .then((response) => {
        context.commit(mutationTypes.loadDetailsSuccess, response?.data?.data || null)
        return response
      })
      .catch((error) => {
        context.commit(mutationTypes.loadDetailsFail, [normalizeApprovalError(error)])
        return error
      })
  },
  [actionTypes.vote](context, payload) {
    context.commit(mutationTypes.voteStart)
    return approvalsApi.voteApproval(payload.number, payload.data)
      .then((response) => {
        context.commit(mutationTypes.voteSuccess)
        return response
      })
      .catch((error) => {
        context.commit(mutationTypes.voteFail, [normalizeApprovalError(error)])
        return error
      })
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}
