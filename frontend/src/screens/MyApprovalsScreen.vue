<script setup>
defineProps({
  isLoading: {
    type: Boolean,
    required: true
  },
  errors: {
    type: Array,
    required: true
  },
  approvals: {
    type: Array,
    required: true
  },
  pageCount: {
    type: Number,
    required: true
  },
  currentPage: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['open-approval', 'set-page'])
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Мои согласования</p>
        <h2>Ожидают решения</h2>
      </div>
      <span class="status-pill purple">ITILIUM</span>
    </div>

    <article v-if="isLoading" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем согласования</h3>
        <p>Получаем задачи, доступные для вашего MAX ID.</p>
      </div>
    </article>

    <p v-else-if="errors.length && !approvals.length" class="status-pill rose">{{ errors[0] }}</p>

    <article v-else-if="!approvals.length" class="content-card">
      <h3>Нет активных согласований</h3>
      <p>Новые задачи появятся здесь, когда ITILIUM направит их вам.</p>
    </article>

    <div v-else class="list-stack">
      <button
        v-for="approval in approvals"
        :key="approval.number"
        class="ticket-card approval-card"
        type="button"
        @click="emit('open-approval', approval.number)"
      >
        <div class="ticket-topline">
          <strong>Согласование {{ approval.number }}</strong>
          <span class="status-pill amber">Требуется решение</span>
        </div>
        <h3>{{ approval.description || 'Описание задачи не указано' }}</h3>
        <p v-if="approval.deadlineHours > 0">Срок исполнения: {{ approval.deadlineHours }} ч.</p>
        <p v-else>Срок исполнения: —</p>
      </button>
    </div>

    <div v-if="pageCount > 1" class="pagination" aria-label="Страницы согласований">
      <button
        v-for="page in pageCount"
        :key="page"
        type="button"
        class="page-button"
        :class="{ active: page === currentPage }"
        @click="emit('set-page', page)"
      >
        {{ page }}
      </button>
    </div>
  </section>
</template>
