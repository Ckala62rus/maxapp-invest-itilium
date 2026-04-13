<script setup>
defineProps({
  createTicketForm: {
    type: Object,
    required: true
  },
  createErrors: {
    type: Array,
    required: true
  },
  isCreatingTicket: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['submit-create-ticket', 'open-screen'])

// The create screen edits the shared create form object from App.vue so the
// ticket create request and navigation outcome stay centralized in one place.
function submitCreateTicket() {
  emit('submit-create-ticket')
}

function openScreen(screenId) {
  emit('open-screen', screenId)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Новая заявка</p>
        <h2>Создание обращения</h2>
      </div>
      <span class="status-pill info">Поддерживает файлы и маркетинг</span>
    </div>

    <article class="content-card form-card">
      <label>
        Тип заявки
        <select v-model="createTicketForm.requestType">
          <option>Заявка в отдел ИТ</option>
          <option>Маркетинговая заявка</option>
        </select>
      </label>
      <label>
        Краткая тема
        <input v-model="createTicketForm.title" type="text" />
      </label>
      <label>
        Подробное описание
        <textarea v-model="createTicketForm.description" rows="5"></textarea>
      </label>
      <label>
        Подразделение
        <select v-model="createTicketForm.department">
          <option>Отдел ИТ</option>
          <option>Маркетинг</option>
        </select>
      </label>
      <label>
        Исполнить до
        <input v-model="createTicketForm.executionDate" type="date" />
      </label>

      <p v-if="createErrors.length" class="status-pill rose">{{ createErrors[0] }}</p>

      <div class="upload-box">
        <div>
          <strong>Вложения</strong>
          <p>Скриншоты, фото, документы, голосовые сообщения.</p>
        </div>
        <button class="secondary-button" disabled>Добавить файл</button>
      </div>

      <div class="chip-list">
        <span v-for="fileName in createTicketForm.attachments" :key="fileName" class="file-chip">{{ fileName }}</span>
      </div>

      <div class="hero-actions">
        <button
          class="primary-button"
          :disabled="isCreatingTicket || !createTicketForm.title || !createTicketForm.description"
          @click="submitCreateTicket"
        >
          {{ isCreatingTicket ? 'Отправка...' : 'Отправить заявку' }}
        </button>
        <button class="secondary-button" :disabled="isCreatingTicket" @click="openScreen('home')">Отмена</button>
      </div>
    </article>
  </section>
</template>
