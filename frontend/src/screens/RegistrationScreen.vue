<script setup>
// Форма регистрации сотрудника при registrationRequired / дозаполнение полей.
defineProps({
  currentIdentity: {
    type: Object,
    default: null
  },
  currentUser: {
    type: Object,
    default: null
  },
  registrationForm: {
    type: Object,
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
  authErrors: {
    type: Array,
    required: true
  },
  isRegistrationSubmitting: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['submit-registration'])

// The registration form edits the shared reactive object from App.vue, so the
// auth flow keeps one source of truth while the screen stays template-focused.
function submitRegistration() {
  emit('submit-registration')
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">MAX авторизация пройдена</p>
        <h2>Вас не нашли в ITILIUM</h2>
      </div>
      <span class="status-pill warning">Требуется заполнение формы</span>
    </div>

    <article class="content-card form-card">
      <div class="compact">
        <span class="status-pill info">MAX ID: {{ currentIdentity?.userId || currentUser?.userId || registrationForm.employeeNumber || 'не получен' }}</span>
        <span class="status-pill rose">{{ currentUser?.statusMessage || 'Пользователь не найден в ITILIUM.' }}</span>
      </div>
      <div class="debug-panel">
        <p><strong>window.WebApp:</strong> {{ maxBridgeState.isAvailable ? 'доступен' : 'недоступен' }}</p>
        <p><strong>initDataUnsafe.user.id:</strong> {{ rawInitDataUnsafeUserId || 'пусто' }}</p>
        <p><strong>initData length:</strong> {{ rawInitData.length }}</p>
        <p><strong>initData raw:</strong></p>
        <pre class="debug-pre">{{ rawInitData || 'ПУСТО' }}</pre>
      </div>
      <p class="supporting-text">
        Заполните данные ниже, чтобы отправить заявку на привязку вашего аккаунта MAX к карточке сотрудника.
      </p>
      <label>
        Max Id
        <input v-model="registrationForm.employeeNumber" type="text" readonly disabled />
      </label>
      <label>
        ФИО
        <input
          v-model="registrationForm.fullName"
          type="text"
          required
          placeholder="Введите ФИО"
        />
      </label>
      <label>
        Организация
        <input
          v-model="registrationForm.organization"
          type="text"
          required
          placeholder="Введите организацию"
        />
      </label>
      <label>
        Подразделение
        <input
          v-model="registrationForm.department"
          type="text"
          required
          placeholder="Введите подразделение"
        />
      </label>
      <label>
        Должность
        <input
          v-model="registrationForm.position"
          type="text"
          required
          placeholder="Введите должность"
        />
      </label>
      <label>
        Телефон
        <input
          v-model="registrationForm.phone"
          type="text"
          placeholder="+7 (900) 123-45-67"
        />
      </label>
      <label>
        Комментарий
        <textarea v-model="registrationForm.comment" rows="4"></textarea>
      </label>
      <p v-if="authErrors.length" class="status-pill rose">{{ authErrors[0] }}</p>
      <button class="primary-button wide" :disabled="isRegistrationSubmitting" @click="submitRegistration">
        {{ isRegistrationSubmitting ? 'Отправка...' : 'Отправить заявку на регистрацию' }}
      </button>
    </article>
  </section>
</template>
