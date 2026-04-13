<script setup>
defineProps({
  registrationForm: {
    type: Object,
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
        <p class="eyebrow">Регистрация</p>
        <h2>Вас не нашли в ITILIUM</h2>
      </div>
      <span class="status-pill warning">Требуется заполнение формы</span>
    </div>

    <article class="content-card form-card">
      <label>
        Табельный номер
        <input v-model="registrationForm.employeeNumber" type="text" />
      </label>
      <label>
        ФИО
        <input v-model="registrationForm.fullName" type="text" />
      </label>
      <label>
        Магазин / подразделение
        <input v-model="registrationForm.department" type="text" />
      </label>
      <label>
        Телефон
        <input v-model="registrationForm.phone" type="text" />
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
