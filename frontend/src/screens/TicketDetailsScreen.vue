<script setup>
// Карточка: таймлайн, смена статуса, ответственный, комментарии.
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
        <button class="primary-button" @click="openScreen('search')">Вернуться к поиску</button>
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
          <span>Срок</span>
          <strong>{{ selectedTicket.deadline }}</strong>
        </div>
        <div>
          <span>Можно сменить ответственного</span>
          <strong>{{ selectedTicket.canChangeResponsible ? 'Да' : 'Нет' }}</strong>
        </div>
      </div>

      <div class="content-card compact">
        <span>Описание</span>
        <p>{{ selectedTicket.description }}</p>
      </div>

      <div class="action-grid">
        <button class="secondary-button" :disabled="isSubmittingComment">Добавить комментарий</button>
        <button class="secondary-button" :disabled="isChangingStatus">Поменять статус</button>
        <button class="secondary-button" :disabled="isChangingResponsible">Сменить ответственного</button>
        <button class="secondary-button" disabled>Оценить решение</button>
      </div>
    </article>

    <article v-else class="content-card">
      <h3>Карточка пока не выбрана</h3>
      <p>Открой заявку из списка или найдите ее по номеру, чтобы загрузить детали из Vuex store.</p>
    </article>

    <article v-if="selectedTicket" class="content-card">
      <h3>Лента событий</h3>
      <div class="timeline">
        <div v-for="item in selectedTicketTimeline" :key="`${item.actor}-${item.time}-${item.text}`" class="timeline-item">
          <strong>{{ item.actor }}</strong>
          <p>{{ item.text }}</p>
          <span>{{ item.time }}</span>
        </div>
      </div>
    </article>

    <article v-if="selectedTicket" class="content-card">
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
      </div>
    </article>

    <article v-if="selectedTicket" class="content-card">
      <h3>Смена статуса</h3>
      <div class="form-card compact">
        <label>
          Новый статус
          <select v-model="statusForm.state">
            <option disabled value="">Выберите статус</option>
            <option v-for="status in availableStatusOptions" :key="status" :value="status">
              {{ status }}
            </option>
          </select>
        </label>
        <label>
          Комментарий к переходу
          <textarea v-model="statusForm.comment" rows="3"></textarea>
        </label>
        <label>
          Дата исполнения
          <input v-model="statusForm.date" type="text" />
        </label>
        <div class="hero-actions">
          <button
            class="primary-button"
            :disabled="isChangingStatus || !statusForm.state"
            @click="submitStatusChange"
          >
            {{ isChangingStatus ? 'Сохраняем...' : 'Сменить статус' }}
          </button>
        </div>
      </div>
    </article>

    <article v-if="selectedTicket" class="content-card">
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
    </article>
  </section>
</template>
