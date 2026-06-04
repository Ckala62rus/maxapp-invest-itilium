<script setup>
// Карточка: таймлайн, смена статуса, ответственный, комментарии.
import { computed, ref, watch } from 'vue'
import { confirmAction } from '@/helpers/confirmDialog'
import { validateStatusTransition } from '@/helpers/ticketWorkflow'

const props = defineProps({
  selectedTicket: {
    type: Object,
    default: null
  },
  searchQuery: {
    type: String,
    required: true
  },
  detailStatusTone: {
    type: String,
    required: true
  },
  isLoadingTicketDetails: {
    type: Boolean,
    required: true
  },
  ticketErrors: {
    type: Array,
    required: true
  },
  detailsOrigin: {
    type: String,
    required: true
  },
  selectedTicketTimeline: {
    type: Array,
    required: true
  },
  commentDraft: {
    type: String,
    required: true
  },
  commentAttachmentFiles: {
    type: Array,
    default: () => []
  },
  isSubmittingComment: {
    type: Boolean,
    required: true
  },
  commentSuccessTick: {
    type: Number,
    default: 0
  },
  ratingSuccessTick: {
    type: Number,
    default: 0
  },
  responsibleSuccessTick: {
    type: Number,
    default: 0
  },
  statusForm: {
    type: Object,
    required: true
  },
  availableStatusOptions: {
    type: Array,
    required: true
  },
  isChangingStatus: {
    type: Boolean,
    required: true
  },
  isLoadingResponsibleOptions: {
    type: Boolean,
    required: true
  },
  availableResponsibleOptions: {
    type: Array,
    required: true
  },
  isChangingResponsible: {
    type: Boolean,
    required: true
  },
  selectedResponsibleId: {
    type: String,
    required: true
  },
  isSubmittingTicketRating: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'open-screen',
  'update:comment-draft',
  'add-comment-files',
  'remove-comment-file',
  'submit-comment',
  'submit-status-change',
  'assign-responsible',
  'request-responsible-options',
  'submit-ticket-rating'
])

const activePanel = ref('')
const statusChangeError = ref('')
/** Краткое уведомление после успешной отправки комментария (панель формы при этом скрывается). */
const showCommentSuccess = ref(false)
const showResponsibleSuccess = ref(false)
const commentEmptyError = ref(false)
const isCommentPanelVisible = computed(() => activePanel.value === 'comment')
const isRatingPanelVisible = computed(() => activePanel.value === 'rating')
const ratingMark = ref(null)
const ratingComment = ref('')
const ratingCommentError = ref('')

watch(
  () => props.commentSuccessTick,
  (next, prev) => {
    if (next > 0 && next !== prev) {
      activePanel.value = ''
      showCommentSuccess.value = true
      window.setTimeout(() => {
        showCommentSuccess.value = false
      }, 6500)
    }
  }
)

watch(
  () => props.ratingSuccessTick,
  (next, prev) => {
    if (next > 0 && next !== prev) {
      activePanel.value = ''
    }
  }
)

watch(
  () => props.responsibleSuccessTick,
  (next, prev) => {
    if (next > 0 && next !== prev) {
      activePanel.value = ''
      showResponsibleSuccess.value = true
      window.setTimeout(() => {
        showResponsibleSuccess.value = false
      }, 6500)
    }
  }
)

watch(
  () => props.selectedTicket?.number,
  () => {
    showCommentSuccess.value = false
    showResponsibleSuccess.value = false
    commentEmptyError.value = false
    activePanel.value = ''
    ratingMark.value = null
    ratingComment.value = ''
    ratingCommentError.value = ''
  }
)

watch(
  () => [props.commentDraft, props.commentAttachmentFiles],
  () => {
    if (commentEmptyError.value) {
      const text = String(props.commentDraft || '').trim()
      const files = props.commentAttachmentFiles || []
      if (text || files.length) {
        commentEmptyError.value = false
      }
    }
  }
)

const canSubmitComment = computed(() => {
  const text = String(props.commentDraft || '').trim()
  const files = props.commentAttachmentFiles || []
  return Boolean(text || files.length)
})
const isStatusSelectionVisible = computed(() => activePanel.value === 'status')
const isResponsibleSelectionVisible = computed(() => activePanel.value === 'responsible')
const isActionGridVisible = computed(() => activePanel.value === '')

function ticketStateSuggestsRating(state) {
  const s = String(state || '').toLowerCase()
  return s.includes('закрыт') || s.includes('выполнен')
}

const canShowRatingAction = computed(() => {
  const t = props.selectedTicket
  if (!t?.number) {
    return false
  }
  if (t.canConfirmRating === true) {
    return true
  }
  return ticketStateSuggestsRating(t.state)
})

const OTHER_RESPONSIBLE_GROUP_KEY = '__other__'

function responsibleBracketMeta(team) {
  const m = String(team || '').match(/^\[\s*([^\]]+?)\s*\]/)
  if (m) {
    const label = m[1].trim()
    return { key: label.toLowerCase(), label }
  }
  return { key: OTHER_RESPONSIBLE_GROUP_KEY, label: 'Прочие подразделения' }
}

/** Группы по префиксу в квадратных скобках: [Барс], [Инвест] и т.д. */
const groupedResponsibleSections = computed(() => {
  const list = props.availableResponsibleOptions || []
  const buckets = new Map()
  for (const person of list) {
    const { key, label } = responsibleBracketMeta(person.team)
    if (!buckets.has(key)) {
      buckets.set(key, { key, label, people: [] })
    }
    buckets.get(key).people.push(person)
  }
  for (const section of buckets.values()) {
    section.people.sort((a, b) =>
      String(a.person || '').localeCompare(String(b.person || ''), 'ru', { sensitivity: 'base' })
    )
  }
  const arr = Array.from(buckets.values())
  arr.sort((a, b) => {
    if (a.key === OTHER_RESPONSIBLE_GROUP_KEY) return 1
    if (b.key === OTHER_RESPONSIBLE_GROUP_KEY) return -1
    return a.label.localeCompare(b.label, 'ru', { sensitivity: 'base' })
  })
  return arr
})

/** Какая группа раскрыта (аккордеон, одна за раз). */
const expandedResponsibleGroupKey = ref('')

watch(
  () => [
    isResponsibleSelectionVisible.value,
    props.isLoadingResponsibleOptions,
    groupedResponsibleSections.value.length,
    groupedResponsibleSections.value.map((g) => `${g.key}:${g.people.length}`).join('|')
  ],
  () => {
    if (!isResponsibleSelectionVisible.value || props.isLoadingResponsibleOptions) {
      return
    }
    const sections = groupedResponsibleSections.value
    if (!sections.length) {
      return
    }
    const current = expandedResponsibleGroupKey.value
    if (!current || !sections.some((s) => s.key === current)) {
      expandedResponsibleGroupKey.value = sections[0].key
    }
  },
  { flush: 'post' }
)

watch(
  () => props.selectedTicket?.number,
  () => {
    expandedResponsibleGroupKey.value = ''
  }
)

function toggleResponsibleGroup(key) {
  expandedResponsibleGroupKey.value = expandedResponsibleGroupKey.value === key ? '' : key
}

function responsiblePersonRowKey(person, index) {
  return `${person.externalId || 'noid'}-${String(person.team || '')}-${index}`
}

// The details screen stays focused on rendering the already prepared ticket data
// while the parent component continues to own store actions and navigation flow.
function openScreen(screenId) {
  emit('open-screen', screenId)
}

function updateCommentDraft(event) {
  emit('update:comment-draft', event.target.value)
}

function onCommentFilesChange(event) {
  const files = event.target?.files
  if (files && files.length) {
    emit('add-comment-files', files)
  }
  event.target.value = ''
}

function removeCommentFile(index) {
  emit('remove-comment-file', index)
}

function submitComment() {
  if (!canSubmitComment.value) {
    commentEmptyError.value = true
    return
  }
  commentEmptyError.value = false
  emit('submit-comment')
}

function submitStatusChange() {
  const validationError = validateStatusTransition(props.statusForm.state, props.statusForm)
  if (validationError) {
    statusChangeError.value = validationError
    return
  }
  statusChangeError.value = ''
  emit('submit-status-change')
}

function assignResponsible(responsibleId) {
  emit('assign-responsible', responsibleId)
}

function openStatusSelection() {
  activePanel.value = 'status'
}

function closeStatusSelection() {
  activePanel.value = ''
}

async function chooseStatus(status) {
  const statusForm = props.statusForm
  statusForm.state = status
  const validationError = validateStatusTransition(status, statusForm)
  if (validationError) {
    statusChangeError.value = validationError
    return
  }
  statusChangeError.value = ''
  const result = await confirmAction({
    title: 'Смена статуса',
    text: `Сменить статус заявки на «${status}»?`,
    confirmButtonText: 'Да, сменить',
    cancelButtonText: 'Отмена'
  })
  if (!result.isConfirmed) {
    return
  }
  submitStatusChange()
}

function openResponsibleSelection() {
  activePanel.value = 'responsible'
  emit('request-responsible-options', props.selectedTicket?.number || '')
}

function closeResponsibleSelection() {
  activePanel.value = ''
}

function openCommentPanel() {
  activePanel.value = 'comment'
}

function closeCommentPanel() {
  activePanel.value = ''
  commentEmptyError.value = false
}

function openRatingPanel() {
  activePanel.value = 'rating'
  ratingMark.value = 5
  ratingComment.value = ''
  ratingCommentError.value = ''
}

function closeRatingPanel() {
  activePanel.value = ''
  ratingCommentError.value = ''
}

function setRatingMark(mark) {
  ratingMark.value = mark
  ratingCommentError.value = ''
}

function updateRatingComment(event) {
  ratingComment.value = event.target.value
  if (ratingCommentError.value && String(ratingComment.value || '').trim()) {
    ratingCommentError.value = ''
  }
}

function lowRatingNeedsComment(mark) {
  return mark !== null && mark >= 0 && mark <= 2
}

function submitTicketRating() {
  if (ratingMark.value === null || ratingMark.value === undefined) {
    return
  }
  if (lowRatingNeedsComment(ratingMark.value) && !String(ratingComment.value || '').trim()) {
    ratingCommentError.value = 'Для оценки 0–2 комментарий обязателен.'
    return
  }
  ratingCommentError.value = ''
  emit('submit-ticket-rating', {
    mark: ratingMark.value,
    comment: String(ratingComment.value || '').trim()
  })
}

function updateStatusComment(event, statusForm) {
  statusForm.comment = event.target.value
  if (statusChangeError.value) {
    statusChangeError.value = validateStatusTransition(statusForm.state, statusForm)
  }
}

// el-date-picker отдаёт YYYY-MM-DD; клик по полю открывает календарь (в т.ч. в Qt WebEngine).
function setStatusPostponeDate(value, statusForm) {
  statusForm.date = value || ''
  if (statusChangeError.value) {
    statusChangeError.value = validateStatusTransition(statusForm.state, statusForm)
  }
}

const statusDateHint = computed(() => {
  const state = String(props.statusForm?.state || '')
  if (String(state).toLowerCase().includes('отлож')) {
    return 'Обязательно для статуса «Отложено».'
  }
  return 'Заполните при переводе в «Отложено».'
})
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Карточка заявки</p>
        <h2>{{ selectedTicket?.number || searchQuery }}</h2>
      </div>
      <span class="status-pill" :class="detailStatusTone">{{ selectedTicket?.state || 'info' }}</span>
    </div>

    <div
      v-if="selectedTicket && showCommentSuccess"
      class="comment-success-banner"
      role="status"
    >
      Комментарий успешно отправлен. Карточка заявки обновлена.
    </div>

    <div
      v-if="selectedTicket && showResponsibleSuccess"
      class="comment-success-banner"
      role="status"
    >
      Ответственный по заявке обновлён. В карточке показаны актуальные данные.
    </div>

    <article v-if="isLoadingTicketDetails" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем карточку</h3>
        <p>Получаем полную информацию о заявке и ее ленту событий из backend API.</p>
      </div>
    </article>

    <article v-else-if="selectedTicket" class="content-card">
      <p v-if="ticketErrors.length" class="status-pill rose">{{ ticketErrors[0] }}</p>
      <div class="details-grid">
        <div>
          <span>Тема</span>
          <strong>{{ selectedTicket.title }}</strong>
        </div>
        <div>
          <span>Дата создания</span>
          <strong>{{ selectedTicket.creationDate || '—' }}</strong>
        </div>
        <div>
          <span>Срок исполнения</span>
          <strong>{{ selectedTicket.deadline || '(неуказан)' }}</strong>
        </div>
        <div>
          <span>Ответственный</span>
          <strong>{{ selectedTicket.responsibleEmployee || selectedTicket.responsibleTeam || 'Не указан' }}</strong>
        </div>
        <div>
          <span>Можно сменить статус</span>
          <strong>{{ selectedTicket.canChangeStatus ? 'Да' : 'Нет' }}</strong>
        </div>
        <div>
          <span>Можно сменить ответственного</span>
          <strong>{{ selectedTicket.canChangeResponsible ? 'Да' : 'Нет' }}</strong>
        </div>
      </div>

      <div class="content-card compact ticket-description-card">
        <span>Описание</span>
        <p>{{ selectedTicket.description }}</p>
      </div>

      <div v-if="isActionGridVisible" class="action-grid">
        <button type="button" class="secondary-button" :disabled="isSubmittingComment" @click="openCommentPanel">Добавить комментарий</button>
        <button
          v-if="selectedTicket.canChangeStatus"
          class="secondary-button"
          :disabled="isChangingStatus"
          @click="openStatusSelection"
        >
          Поменять статус
        </button>
        <button
          v-if="selectedTicket.canChangeResponsible"
          class="secondary-button"
          :disabled="isChangingResponsible"
          @click="openResponsibleSelection"
        >
          Сменить ответственного
        </button>
        <button
          v-if="canShowRatingAction"
          class="secondary-button"
          :disabled="isSubmittingTicketRating"
          @click="openRatingPanel"
        >
          Оценить решение
        </button>
      </div>
    </article>

    <article v-else-if="ticketErrors.length" class="content-card">
      <h3>Заявку не удалось открыть</h3>
      <p>{{ ticketErrors[0] }}</p>
      <div class="hero-actions">
        <button type="button" class="primary-button" @click="openScreen(detailsOrigin === 'myTickets' ? 'myTickets' : 'search')">
          {{ detailsOrigin === 'myTickets' ? 'Вернуться к моим заявкам' : 'Вернуться к поиску' }}
        </button>
      </div>
    </article>

    <article v-else class="content-card">
      <h3>Карточка пока не выбрана</h3>
      <p>Открой заявку из списка или найдите ее по номеру, чтобы загрузить детали из Vuex store.</p>
    </article>

    <article v-if="selectedTicket && isCommentPanelVisible" class="content-card comment-panel-block">
      <h3>Новый комментарий</h3>
      <div class="form-card compact">
        <label>
          Текст комментария
          <textarea
            :value="commentDraft"
            rows="4"
            placeholder="Текст или вложение"
            :class="{ 'field-invalid': commentEmptyError }"
            @input="updateCommentDraft"
          ></textarea>
          <small v-if="commentEmptyError" class="field-error">Введите текст комментария или прикрепите хотя бы один файл.</small>
        </label>
        <div class="upload-box">
          <div>
            <strong>Вложения</strong>
            <p>По желанию: скриншот, документ.</p>
          </div>
          <label class="secondary-button upload-trigger">
            Добавить файл
            <input
              class="upload-input"
              type="file"
              multiple
              accept="image/*,audio/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.zip"
              @change="onCommentFilesChange"
            />
          </label>
        </div>

        <div v-if="commentAttachmentFiles.length" class="chip-list">
          <div
            v-for="(file, index) in commentAttachmentFiles"
            :key="file.name + '-' + index"
            class="file-chip"
          >
            <span>{{ file.name }}</span>
            <button
              class="file-chip-remove"
              type="button"
              :disabled="isSubmittingComment"
              @click="removeCommentFile(index)"
            >
              Удалить
            </button>
          </div>
        </div>
        <div class="hero-actions">
          <button
            class="primary-button"
            :disabled="isSubmittingComment"
            @click="submitComment"
          >
            {{ isSubmittingComment ? 'Отправляем...' : 'Отправить комментарий' }}
          </button>
        </div>
        <div class="hero-actions">
          <button type="button" class="ghost-button" :disabled="isSubmittingComment" @click="closeCommentPanel">Назад</button>
        </div>
      </div>
    </article>

    <article
      v-if="selectedTicket && selectedTicket.canChangeStatus && isStatusSelectionVisible"
      class="content-card status-change-panel"
    >
      <h3>Смена статуса</h3>
      <div class="form-card compact">
        <label>
          Комментарий к смене статуса
          <textarea
            :value="statusForm.comment"
            rows="3"
            placeholder="Обязателен для «В ожидании ответа» и «Отложено»"
            @input="updateStatusComment($event, statusForm)"
          ></textarea>
        </label>
        <label>
          отложить до
          <el-date-picker
            class="date-field status-postpone-date"
            :model-value="statusForm.date"
            type="date"
            value-format="YYYY-MM-DD"
            format="DD.MM.YYYY"
            placeholder="Выберите дату"
            :disabled="isChangingStatus"
            :teleported="false"
            @update:model-value="setStatusPostponeDate($event, statusForm)"
          />
          <small class="supporting-text">{{ statusDateHint }}</small>
        </label>
      </div>
      <p v-if="statusChangeError" class="status-pill rose">{{ statusChangeError }}</p>
      <div class="hero-actions">
        <button
          v-for="status in availableStatusOptions"
          :key="status"
          type="button"
          class="secondary-button status-change-button"
          :disabled="isChangingStatus"
          @click="chooseStatus(status)"
        >
          {{ status }}
        </button>
      </div>
      <div class="hero-actions">
        <button type="button" class="ghost-button" :disabled="isChangingStatus" @click="closeStatusSelection">Назад</button>
      </div>
    </article>

    <article v-if="selectedTicket && canShowRatingAction && isRatingPanelVisible" class="content-card rating-panel-block">
      <h3>Оценка решения</h3>
      <p class="supporting-text">Оценка по шкале 0–5. Для оценок 0, 1 и 2 комментарий обязателен.</p>
      <div class="rating-stars-row" role="group" aria-label="Оценка от 1 до 5">
        <button
          v-for="star in 5"
          :key="star"
          type="button"
          class="rating-star"
          :class="{ active: ratingMark !== null && star <= ratingMark }"
          :disabled="isSubmittingTicketRating"
          :aria-pressed="ratingMark === star ? 'true' : 'false'"
          @click="setRatingMark(star)"
        >
          <span class="rating-star-icon" aria-hidden="true">★</span>
        </button>
      </div>
      <div class="hero-actions">
        <button
          type="button"
          class="secondary-button"
          :disabled="isSubmittingTicketRating"
          @click="setRatingMark(0)"
        >
          Обращение не решено (0)
        </button>
      </div>
      <label>
        Комментарий к оценке
        <textarea
          :value="ratingComment"
          rows="3"
          placeholder="Обязателен при оценке 0–2"
          :class="{ 'field-invalid': ratingCommentError }"
          @input="updateRatingComment"
        ></textarea>
        <small v-if="ratingCommentError" class="field-error">{{ ratingCommentError }}</small>
      </label>
      <div class="hero-actions">
        <button
          class="primary-button"
          :disabled="isSubmittingTicketRating || ratingMark === null"
          @click="submitTicketRating"
        >
          {{ isSubmittingTicketRating ? 'Отправляем...' : 'Отправить оценку' }}
        </button>
        <button class="ghost-button" type="button" :disabled="isSubmittingTicketRating" @click="closeRatingPanel">
          Назад
        </button>
      </div>
    </article>

    <article
      v-if="selectedTicket && selectedTicket.canChangeResponsible && isResponsibleSelectionVisible"
      class="content-card responsible-change-panel"
    >
      <h3>Выбор нового ответственного</h3>
      <article v-if="isLoadingResponsibleOptions" class="state-card compact">
        <div class="spinner"></div>
        <div>
          <h3>Загружаем список</h3>
          <p>Получаем доступных ответственных для этой заявки.</p>
        </div>
      </article>
      <p v-else-if="!groupedResponsibleSections.length" class="supporting-text">
        Нет доступных кандидатов для этой заявки.
      </p>
      <div v-else class="responsible-groups">
        <p class="supporting-text responsible-groups-hint">
          Выберите подразделение — список сотрудников откроется ниже.
        </p>
        <div
          v-for="group in groupedResponsibleSections"
          :key="group.key"
          class="responsible-group"
        >
          <button
            type="button"
            class="responsible-group-header"
            :class="{ 'is-open': expandedResponsibleGroupKey === group.key }"
            :disabled="isChangingResponsible"
            :aria-expanded="expandedResponsibleGroupKey === group.key ? 'true' : 'false'"
            @click="toggleResponsibleGroup(group.key)"
          >
            <span class="responsible-group-title">{{ group.label }}</span>
            <span class="responsible-group-meta">{{ group.people.length }}</span>
            <span class="responsible-group-chevron" aria-hidden="true">
              {{ expandedResponsibleGroupKey === group.key ? '▼' : '▶' }}
            </span>
          </button>
          <div
            v-show="expandedResponsibleGroupKey === group.key"
            class="responsible-group-panel selector-list"
          >
            <div
              v-for="(person, idx) in group.people"
              :key="responsiblePersonRowKey(person, idx)"
              class="selector-item"
            >
              <div>
                <strong>{{ person.person }}</strong>
                <p class="responsible-person-team">
                  {{ person.team }}<template v-if="person.role"> · {{ person.role }}</template>
                </p>
              </div>
              <button
                type="button"
                class="ghost-button responsible-assign-button"
                :disabled="isChangingResponsible"
                @click="assignResponsible(person.externalId)"
              >
                {{ isChangingResponsible && selectedResponsibleId === person.externalId ? 'Сохраняем...' : 'Выбрать' }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="hero-actions">
        <button type="button" class="ghost-button" :disabled="isChangingResponsible" @click="closeResponsibleSelection">Назад</button>
      </div>
    </article>
  </section>
</template>
