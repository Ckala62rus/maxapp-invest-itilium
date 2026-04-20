<script setup>
// Карточка: таймлайн, смена статуса, ответственный, комментарии.
import { computed, ref } from 'vue'

defineProps({
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
  isSubmittingComment: {
    type: Boolean,
    required: true
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
  }
})

const emit = defineEmits([
  'open-screen',
  'update:comment-draft',
  'submit-comment',
  'submit-status-change',
  'assign-responsible'
])

const activePanel = ref('')
const statusChangeError = ref('')
const isCommentPanelVisible = computed(() => activePanel.value === 'comment')
const isStatusSelectionVisible = computed(() => activePanel.value === 'status')
const isResponsibleSelectionVisible = computed(() => activePanel.value === 'responsible')
const isActionGridVisible = computed(() => activePanel.value === '')

// The details screen stays focused on rendering the already prepared ticket data
// while the parent component continues to own store actions and navigation flow.
function openScreen(screenId) {
  emit('open-screen', screenId)
}

function updateCommentDraft(event) {
  emit('update:comment-draft', event.target.value)
}

function submitComment() {
  emit('submit-comment')
}

function submitStatusChange() {
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

function chooseStatus(status, statusForm) {
  statusForm.state = status
  if (isWaitingForResponseStatus(status) && !String(statusForm.comment || '').trim()) {
    statusChangeError.value = 'Для статуса «В ожидании ответа» комментарий обязателен.'
    return
  }
  statusChangeError.value = ''
  submitStatusChange()
}

function openResponsibleSelection() {
  activePanel.value = 'responsible'
}

function closeResponsibleSelection() {
  activePanel.value = ''
}

function openCommentPanel() {
  activePanel.value = 'comment'
}

function closeCommentPanel() {
  activePanel.value = ''
}

function updateStatusComment(event, statusForm) {
  statusForm.comment = event.target.value
  if (statusChangeError.value && String(statusForm.comment || '').trim()) {
    statusChangeError.value = ''
  }
}

function isWaitingForResponseStatus(status) {
  return String(status || '').toLowerCase().includes('в ожидании ответа')
}
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

    <article v-if="isLoadingTicketDetails" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем карточку</h3>
        <p>Получаем полную информацию о заявке и ее ленту событий из backend API.</p>
      </div>
    </article>

    <article v-else-if="ticketErrors.length" class="content-card">
      <h3>Заявку не удалось открыть</h3>
      <p>{{ ticketErrors[0] }}</p>
      <div class="hero-actions">
        <button class="primary-button" @click="openScreen(detailsOrigin === 'myTickets' ? 'myTickets' : 'search')">
          {{ detailsOrigin === 'myTickets' ? 'Вернуться к моим заявкам' : 'Вернуться к поиску' }}
        </button>
      </div>
    </article>

    <article v-else-if="selectedTicket" class="content-card">
      <div class="details-grid">
        <div>
          <span>Краткая тема</span>
          <strong>{{ selectedTicket.title }}</strong>
        </div>
        <div>
          <span>Ответственная команда</span>
          <strong>{{ selectedTicket.responsibleTeam }}</strong>
        </div>
        <div>
          <span>Дата создания</span>
          <strong>{{ selectedTicket.creationDate || '—' }}</strong>
        </div>
        <div>
          <span>Срок исполнения</span>
          <strong>{{ selectedTicket.deadline }}</strong>
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
        <button class="secondary-button" :disabled="isSubmittingComment" @click="openCommentPanel">Добавить комментарий</button>
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
        <button class="secondary-button" disabled>Оценить решение</button>
      </div>
    </article>

    <article v-else class="content-card">
      <h3>Карточка пока не выбрана</h3>
      <p>Открой заявку из списка или найдите ее по номеру, чтобы загрузить детали из Vuex store.</p>
    </article>

    <article v-if="selectedTicket && isCommentPanelVisible" class="content-card">
      <h3>Новый комментарий</h3>
      <div class="form-card compact">
        <label>
          Текст комментария
          <textarea :value="commentDraft" rows="4" @input="updateCommentDraft"></textarea>
        </label>
        <div class="hero-actions">
          <button
            class="primary-button"
            :disabled="isSubmittingComment || !commentDraft.trim()"
            @click="submitComment"
          >
            {{ isSubmittingComment ? 'Отправляем...' : 'Отправить комментарий' }}
          </button>
          <button class="secondary-button" disabled>Прикрепить файл</button>
        </div>
        <div class="hero-actions">
          <button class="ghost-button" :disabled="isSubmittingComment" @click="closeCommentPanel">Назад</button>
        </div>
      </div>
    </article>

    <article v-if="selectedTicket && selectedTicket.canChangeStatus && isStatusSelectionVisible" class="content-card">
      <h3>Смена статуса</h3>
      <div class="form-card compact">
        <label>
          Комментарий к смене статуса
          <textarea
            :value="statusForm.comment"
            rows="3"
            placeholder="Комментарий обязателен для статуса «В ожидании ответа»"
            @input="updateStatusComment($event, statusForm)"
          ></textarea>
        </label>
      </div>
      <p v-if="statusChangeError" class="status-pill rose">{{ statusChangeError }}</p>
      <div class="hero-actions">
        <button
          v-for="status in availableStatusOptions"
          :key="status"
          class="secondary-button"
          :disabled="isChangingStatus"
          @click="chooseStatus(status, statusForm)"
        >
          {{ status }}
        </button>
      </div>
      <div class="hero-actions">
        <button class="ghost-button" :disabled="isChangingStatus" @click="closeStatusSelection">Назад</button>
      </div>
    </article>

    <article v-if="selectedTicket && selectedTicket.canChangeResponsible && isResponsibleSelectionVisible" class="content-card">
      <h3>Выбор нового ответственного</h3>
      <article v-if="isLoadingResponsibleOptions" class="state-card compact">
        <div class="spinner"></div>
        <div>
          <h3>Загружаем список</h3>
          <p>Получаем доступных ответственных для этой заявки.</p>
        </div>
      </article>
      <div v-else class="selector-list">
        <div v-for="person in availableResponsibleOptions" :key="person.externalId || person.person" class="selector-item">
          <div>
            <strong>{{ person.person }}</strong>
            <p>{{ person.team }} · {{ person.role }}</p>
          </div>
          <button
            class="ghost-button"
            :disabled="isChangingResponsible"
            @click="assignResponsible(person.externalId)"
          >
            {{ isChangingResponsible && selectedResponsibleId === person.externalId ? 'Сохраняем...' : 'Выбрать' }}
          </button>
        </div>
      </div>
      <div class="hero-actions">
        <button class="ghost-button" :disabled="isChangingResponsible" @click="closeResponsibleSelection">Назад</button>
      </div>
    </article>
  </section>
</template>
