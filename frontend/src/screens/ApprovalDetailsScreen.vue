<script setup>
defineProps({
  approval: {
    type: Object,
    default: null
  },
  isLoading: {
    type: Boolean,
    required: true
  },
  isSubmittingVote: {
    type: Boolean,
    required: true
  },
  errors: {
    type: Array,
    required: true
  },
  voteComment: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['open-screen', 'update:vote-comment', 'submit-vote'])

function updateVoteComment(event) {
  emit('update:vote-comment', event.target.value)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Задача на согласование</p>
        <h2>{{ approval?.number || 'Согласование' }}</h2>
      </div>
      <button class="ghost-button details-back-button" type="button" @click="emit('open-screen', 'myApprovals')">
        ← К списку
      </button>
    </div>

    <article v-if="isLoading" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем задачу</h3>
        <p>Получаем сведения о согласовании из ITILIUM.</p>
      </div>
    </article>

    <article v-else-if="approval" class="content-card approval-details-card">
      <p v-if="errors.length" class="status-pill rose">{{ errors[0] }}</p>
      <div class="details-grid">
        <div>
          <span>Документ</span>
          <strong>{{ approval.document || '—' }}</strong>
        </div>
        <div>
          <span>Срок исполнения</span>
          <strong>{{ approval.deadlineDate || '—' }}</strong>
        </div>
        <div>
          <span>Дата исполнения</span>
          <strong>{{ approval.executionDate || '—' }}</strong>
        </div>
        <div>
          <span>Результат согласования</span>
          <strong>{{ approval.resultsNegotiation || 'Ожидает решения' }}</strong>
        </div>
        <div>
          <span>Согласовал</span>
          <strong>{{ approval.author || '—' }}</strong>
        </div>
      </div>

      <div class="content-card compact ticket-description-card">
        <span>Описание</span>
        <p>{{ approval.description || 'Описание не предоставлено.' }}</p>
      </div>

      <label class="approval-comment-field">
        Комментарий к решению
        <textarea
          :value="voteComment"
          rows="3"
          placeholder="Необязательно"
          :disabled="isSubmittingVote"
          @input="updateVoteComment"
        ></textarea>
      </label>

      <div class="approval-actions">
        <button
          class="primary-button"
          type="button"
          :disabled="isSubmittingVote"
          @click="emit('submit-vote', 'accept')"
        >
          {{ isSubmittingVote ? 'Отправляем…' : 'Согласовать' }}
        </button>
        <button
          class="secondary-button approval-reject-button"
          type="button"
          :disabled="isSubmittingVote"
          @click="emit('submit-vote', 'reject')"
        >
          Отклонить
        </button>
      </div>
    </article>

    <article v-else class="content-card">
      <h3>Согласование не удалось открыть</h3>
      <p>{{ errors[0] || 'Повторите попытку из списка согласований.' }}</p>
      <div class="hero-actions">
        <button class="primary-button" type="button" @click="emit('open-screen', 'myApprovals')">К списку</button>
      </div>
    </article>
  </section>
</template>
