import axios from '@/api/axios'
import urls from '@/api/urls'

/** Список номеров согласований текущего пользователя. */
const listApprovals = () => axios.get(urls.approvals)

/** Карточка одного согласования. */
const getApproval = (number) => axios.get(urls.approvalDetails(number))

/** Решение по согласованию: state принимает accept или reject. */
const voteApproval = (number, payload) => axios.post(urls.approvalVote(number), payload)

export default {
  listApprovals,
  getApproval,
  voteApproval
}
