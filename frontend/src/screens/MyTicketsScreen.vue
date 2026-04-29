<script setup>
// Список «мои заявки»: из API или из servicecalls профиля + пагинация.
defineProps({
  isLoadingMyTickets: {
    type: Boolean,
    required: true
  },
  listErrors: {
    type: Array,
    required: true
  },
  paginatedTickets: {
    type: Array,
    required: true
  },
  pageCount: {
    type: Number,
    required: true
  },
  currentTicketsPage: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['open-ticket-details', 'set-tickets-page'])

// The list screen keeps rendering concerns local and emits user intent upward
// so App.vue still owns the shared pagination and ticket-loading flow.
function openTicketDetails(ticketNumber) {
  emit('open-ticket-details', ticketNumber)
}

function setTicketsPage(page) {
  emit('set-tickets-page', page)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Мои заявки</p>
        <h2>История обращений</h2>
      </div>
      <span class="status-pill info">Номера и карточки из ITILIUM</span>
    </div>

    <article v-if="isLoadingMyTickets" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем список</h3>
        <p>Подгружаем ваши заявки из поля servicecalls профиля.</p>
      </div>
    </article>

    <p v-else-if="listErrors.length && !paginatedTickets.length" class="status-pill rose">{{ listErrors[0] }}</p>

    <article v-else-if="!paginatedTickets.length" class="content-card">
      <h3>Заявок пока нет</h3>
      <p>В профиле ITILIUM пока нет номеров в поле servicecalls.</p>
    </article>

    <div v-else class="list-stack">
      <article
        v-for="ticket in paginatedTickets"
        :key="ticket.number"
        class="ticket-card"
        @click="openTicketDetails(ticket.number)"
      >
        <div class="ticket-topline">
          <strong>Заявка {{ ticket.number }}</strong>
          <span class="status-pill" :class="ticket.tone">{{ ticket.state }}</span>
        </div>
        <h3>{{ ticket.title || 'Тема не указана' }}</h3>
        <p>Дата создания: {{ ticket.creationDate || '—' }}</p>
      </article>
    </div>

    <div v-if="pageCount > 1" class="pagination">
      <button
        v-for="page in pageCount"
        :key="page"
        class="page-button"
        :class="{ active: page === currentTicketsPage }"
        @click="setTicketsPage(page)"
      >
        {{ page }}
      </button>
    </div>
  </section>
</template>
