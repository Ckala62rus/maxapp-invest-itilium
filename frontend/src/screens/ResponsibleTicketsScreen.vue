<script setup>
defineProps({
  isLoadingResponsibleTickets: {
    type: Boolean,
    required: true
  },
  listErrors: {
    type: Array,
    required: true
  },
  normalizedResponsibleTickets: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['open-ticket-details'])

// The responsibility screen stays focused on rendering the already prepared
// ticket list and forwards selection to the parent ticket flow.
function openTicketDetails(ticketNumber) {
  emit('open-ticket-details', ticketNumber)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Ответственность</p>
        <h2>Заявки, закрепленные за мной</h2>
      </div>
      <span class="status-pill warning">Требуют реакции</span>
    </div>

    <article v-if="isLoadingResponsibleTickets" class="state-card">
      <div class="spinner"></div>
      <div>
        <h3>Загружаем ответственность</h3>
        <p>Получаем заявки, закрепленные за текущим пользователем.</p>
      </div>
    </article>

    <p v-else-if="listErrors.length" class="status-pill rose">{{ listErrors[0] }}</p>

    <div v-else class="list-stack">
      <article
        v-for="ticket in normalizedResponsibleTickets"
        :key="ticket.number"
        class="ticket-card"
        @click="openTicketDetails(ticket.number)"
      >
        <div class="ticket-topline">
          <strong>{{ ticket.number }}</strong>
          <span class="status-pill" :class="ticket.tone">{{ ticket.state }}</span>
        </div>
        <h3>{{ ticket.title }}</h3>
        <p>Инициатор: {{ ticket.owner }}</p>
      </article>
    </div>
  </section>
</template>
