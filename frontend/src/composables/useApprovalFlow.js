import { computed, ref, watch } from 'vue'
import { confirmAction } from '@/helpers/confirmDialog'
import { withBusyModal } from '@/helpers/busyModal'
import {
  actionTypes as approvalActionTypes,
  getterTypes as approvalGetterTypes
} from '@/store/modules/approvals'

const approvalsPageSize = 5

export function useApprovalFlow({ store, activeScreen, submitBanner }) {
  const currentApprovalsPage = ref(1)
  const voteComment = ref('')

  // The list contains the vote number, description, and deadline hours returned by list_negotations.
  const approvalSummaries = computed(() => store.getters[approvalGetterTypes.items] || [])
  const selectedApproval = computed(() => store.getters[approvalGetterTypes.selectedApproval] || null)
  const isLoadingApprovals = computed(() => store.getters[approvalGetterTypes.isLoadingList])
  const isLoadingApprovalDetails = computed(() => store.getters[approvalGetterTypes.isLoadingDetails])
  const isSubmittingApprovalVote = computed(() => store.getters[approvalGetterTypes.isSubmittingVote])
  const approvalListErrors = computed(() => store.getters[approvalGetterTypes.listError] || [])
  const approvalDetailsErrors = computed(() => store.getters[approvalGetterTypes.detailsError] || [])

  const paginatedApprovals = computed(() => {
    // Paginate already-complete summaries locally; no extra detail request is made for list cards.
    const start = (currentApprovalsPage.value - 1) * approvalsPageSize
    return approvalSummaries.value.slice(start, start + approvalsPageSize)
  })
  const approvalsPageCount = computed(() => Math.ceil(approvalSummaries.value.length / approvalsPageSize))

  watch(approvalSummaries, () => {
    if (approvalsPageCount.value > 0 && currentApprovalsPage.value > approvalsPageCount.value) {
      currentApprovalsPage.value = approvalsPageCount.value
    }
  })

  watch(() => activeScreen.value, (screen) => {
    if (screen === 'myApprovals') {
      loadApprovals()
    }
  })

  async function loadApprovals() {
    currentApprovalsPage.value = 1
    return store.dispatch(approvalActionTypes.loadList)
  }

  async function openApproval(number) {
    const normalizedNumber = String(number || '').trim()
    if (!normalizedNumber) {
      return
    }

    voteComment.value = ''
    activeScreen.value = 'approvalDetails'
    await store.dispatch(approvalActionTypes.loadDetails, normalizedNumber)
  }

  function setApprovalsPage(page) {
    const next = Number(page) || 1
    if (next < 1 || next > approvalsPageCount.value) {
      return
    }
    currentApprovalsPage.value = next
  }

  function setVoteComment(value) {
    voteComment.value = value
  }

  async function submitApprovalVote(state) {
    const approvalNumber = String(selectedApproval.value?.number || '').trim()
    if (!approvalNumber || !['accept', 'reject'].includes(state)) {
      return
    }

    const isAccept = state === 'accept'
    const confirmation = await confirmAction({
      title: isAccept ? 'Согласовать задачу' : 'Отклонить согласование',
      text: isAccept
        ? 'Подтвердить согласование этой задачи?'
        : 'Подтвердить отклонение этого согласования?',
      confirmButtonText: isAccept ? 'Согласовать' : 'Отклонить',
      cancelButtonText: 'Отмена'
    })
    if (!confirmation.isConfirmed) {
      return
    }

    const response = await withBusyModal(
      isAccept ? 'Согласовываем задачу…' : 'Отклоняем согласование…',
      () => store.dispatch(approvalActionTypes.vote, {
        number: approvalNumber,
        data: {
          state,
          commentText: String(voteComment.value || '').trim()
        }
      })
    )

    if (response?.data?.success) {
      submitBanner.value = response.data.message || (isAccept ? 'Согласование подтверждено.' : 'Согласование отклонено.')
      voteComment.value = ''
      await loadApprovals()
      activeScreen.value = 'myApprovals'
    }

    return response
  }

  return {
    currentApprovalsPage,
    voteComment,
    paginatedApprovals,
    approvalsPageCount,
    selectedApproval,
    isLoadingApprovals,
    isLoadingApprovalDetails,
    isSubmittingApprovalVote,
    approvalListErrors,
    approvalDetailsErrors,
    loadApprovals,
    openApproval,
    setApprovalsPage,
    setVoteComment,
    submitApprovalVote
  }
}
