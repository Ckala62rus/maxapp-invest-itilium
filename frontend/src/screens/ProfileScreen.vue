<script setup>
defineProps({
  currentUser: {
    type: Object,
    default: null
  },
  profileInitials: {
    type: String,
    required: true
  },
  profileStatusText: {
    type: String,
    required: true
  },
  profileRegion: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['open-screen'])

// The profile screen only renders derived auth state and emits the next target
// screen, keeping navigation and auth requests centralized in App.vue.
function openScreen(screenId) {
  emit('open-screen', screenId)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Профиль</p>
        <h2>Пользователь MAX</h2>
      </div>
      <span class="status-pill info">Авторизация пройдена</span>
    </div>

    <article class="content-card">
      <div class="profile-row">
        <div class="avatar">{{ profileInitials }}</div>
        <div>
          <h3>{{ currentUser?.fullName || 'Загрузка профиля...' }}</h3>
          <p>@{{ currentUser?.username || 'unknown' }} · user_id {{ currentUser?.userId || '...' }}</p>
        </div>
      </div>
      <div class="details-grid">
        <div>
          <span>Статус в ITILIUM</span>
          <strong>{{ profileStatusText }}</strong>
        </div>
        <div>
          <span>Роль в MAX</span>
          <strong>Пользователь mini app</strong>
        </div>
        <div>
          <span>Регион</span>
          <strong>{{ profileRegion }}</strong>
        </div>
        <div>
          <span>Последний вход</span>
          <strong>09.04.2026 22:30</strong>
        </div>
      </div>
      <button
        class="primary-button wide"
        @click="openScreen(currentUser?.registrationRequired ? 'registration' : 'home')"
      >
        {{ currentUser?.registrationRequired ? 'Перейти к регистрации' : 'Перейти на главную' }}
      </button>
    </article>
  </section>
</template>
