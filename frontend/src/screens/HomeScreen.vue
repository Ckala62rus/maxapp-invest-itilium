<script setup>
// Главная: сводка и быстрые переходы после успешного онбординга.
defineProps({
  summaryCards: {
    type: Array,
    required: true
  },
  maxBridgeState: {
    type: Object,
    required: true
  },
  rawInitData: {
    type: String,
    required: true
  },
  rawInitDataUnsafeUserId: {
    type: String,
    required: true
  },
  showDebugInfo: {
    type: Boolean,
    default: false
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
      <div class="hero-actions home-hero-actions">
        <button class="primary-button wide" @click="openScreen('create')">Создать заявку</button>
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

    <article v-if="showDebugInfo" class="content-card">
      <h3>MAX bridge debug</h3>
      <p><strong>window.WebApp:</strong> {{ maxBridgeState.isAvailable ? 'доступен' : 'недоступен' }}</p>
      <p><strong>initDataUnsafe.user.id:</strong> {{ rawInitDataUnsafeUserId || 'пусто' }}</p>
      <p><strong>initData length:</strong> {{ rawInitData.length }}</p>
      <p><strong>initData raw:</strong></p>
      <pre class="debug-pre">{{ rawInitData || 'ПУСТО' }}</pre>
    </article>
  </section>
</template>
