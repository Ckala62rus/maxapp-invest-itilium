<script setup>
defineProps({
  searchQuery: {
    type: String,
    required: true
  },
  ticketErrors: {
    type: Array,
    required: true
  },
  isLoadingTicketDetails: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:search-query', 'search-ticket', 'open-ticket-details'])

// The search screen only edits the current query and bubbles actions upward,
// keeping the actual ticket API flow centralized in App.vue and the store.
function updateSearchQuery(event) {
  emit('update:search-query', event.target.value)
}

function searchTicket() {
  emit('search-ticket')
}

function openTicketDetails() {
  emit('open-ticket-details')
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Поиск</p>
        <h2>Поиск заявки по номеру</h2>
      </div>
      <span class="status-pill info">Быстрый доступ к карточке</span>
    </div>

    <article class="content-card form-card">
      <label>
        Номер заявки
        <input :value="searchQuery" type="text" @input="updateSearchQuery" />
      </label>
      <p v-if="ticketErrors.length" class="status-pill rose">{{ ticketErrors[0] }}</p>
      <div class="hero-actions">
        <button class="primary-button" :disabled="isLoadingTicketDetails" @click="searchTicket">
          {{ isLoadingTicketDetails ? 'Ищем заявку...' : 'Найти заявку' }}
        </button>
        <button class="secondary-button" :disabled="isLoadingTicketDetails" @click="openTicketDetails">
          Открыть карточку по номеру
        </button>
      </div>
    </article>
  </section>
</template>
