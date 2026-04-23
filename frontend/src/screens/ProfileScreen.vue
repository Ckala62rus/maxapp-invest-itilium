<script setup>
// Профиль MAX / ITILIUM, статусы регистрации, реквизиты из GET /users/me.
defineProps({
  currentUser: {
    type: Object,
    default: null
  },
  currentIdentity: {
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
        </div>
      </div>
      <div class="details-grid">
        <div>
          <span>MAX ID</span>
          <strong>{{ currentIdentity?.userId || currentUser?.userId || '...' }}</strong>
        </div>
        <div>
          <span>Статус в ITILIUM</span>
          <strong>{{ profileStatusText }}</strong>
        </div>
        <div>
          <span>Организация</span>
          <strong>{{ currentUser?.organization || '—' }}</strong>
        </div>
        <div>
          <span>Подразделение</span>
          <strong>{{ currentUser?.department || '—' }}</strong>
        </div>
        <div>
          <span>Должность</span>
          <strong>{{ currentUser?.position || '—' }}</strong>
        </div>
        <div>
          <span>Номеров заявок в профиле</span>
          <strong>{{ currentUser?.servicecalls?.length ?? '—' }}</strong>
        </div>
        <div>
          <span>Роль в MAX</span>
          <strong>Пользователь mini app</strong>
        </div>
        <div>
          <span>Регион (из подразделения)</span>
          <strong>{{ profileRegion }}</strong>
        </div>
      </div>
      <p v-if="currentUser?.statusMessage" class="status-pill warning">{{ currentUser.statusMessage }}</p>
      <button
        class="primary-button wide"
        @click="openScreen(currentUser?.employeeFound && !currentUser?.registrationRequired ? 'home' : 'registration')"
      >
        {{ currentUser?.employeeFound && !currentUser?.registrationRequired ? 'Перейти на главную' : 'Перейти к регистрации' }}
      </button>
    </article>
  </section>
</template>
