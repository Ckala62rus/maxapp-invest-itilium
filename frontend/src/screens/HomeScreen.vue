<script setup>
defineProps({
  summaryCards: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['open-screen'])

// The home screen stays presentation-focused and forwards navigation intent
// upward so App.vue remains the single owner of active screen state.
function openScreen(screenId) {
  emit('open-screen', screenId)
}
</script>

<template>
  <section class="screen">
    <div class="hero-card">
      <p class="status-pill success">Mini App готов к демонстрации</p>
      <h2>Управляйте заявками ITILIUM прямо внутри MAX</h2>
      <p>
        Пользователь сможет пройти регистрацию, создать обращение, отследить статус,
        оставить комментарий и работать с заявками в своей ответственности.
      </p>
      <div class="hero-actions">
        <button class="primary-button" @click="openScreen('create')">Создать заявку</button>
        <button class="secondary-button" @click="openScreen('myTickets')">Мои заявки</button>
      </div>
    </div>

    <div class="summary-grid">
      <article
        v-for="card in summaryCards"
        :key="card.title"
        class="summary-card"
        :class="card.tone"
      >
        <span>{{ card.title }}</span>
        <strong>{{ card.value }}</strong>
      </article>
    </div>

    <div class="state-grid">
      <article class="state-card">
        <div class="spinner"></div>
        <div>
          <h3>Loading state</h3>
          <p>Используем на экранах поиска, списка и отправки заявки.</p>
        </div>
      </article>
      <article class="state-card">
        <div class="state-icon empty">0</div>
        <div>
          <h3>Empty state</h3>
          <p>Нет заявок в выборке. Предлагаем создать новое обращение.</p>
        </div>
      </article>
      <article class="state-card">
        <div class="state-icon error">!</div>
        <div>
          <h3>Error state</h3>
          <p>Итилиум недоступен или вернул ошибку. Показываем дружелюбный текст.</p>
        </div>
      </article>
    </div>
  </section>
</template>
