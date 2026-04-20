<script setup>
// Создание заявки: тип (ИТ / маркетинг / DAX по флагам профиля), текст, срок.
defineProps({
  createTicketForm: {
    type: Object,
    required: true
  },
  availableRequestTypes: {
    type: Array,
    required: true
  },
  createErrors: {
    type: Array,
    required: true
  },
  createValidationErrors: {
    type: Object,
    required: true
  },
  createValidationStarted: {
    type: Boolean,
    required: true
  },
  createErrorMessage: {
    type: String,
    required: true
  },
  isCreatingTicket: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['submit-create-ticket', 'open-screen', 'set-execution-date', 'add-attachments', 'remove-attachment'])

// Экран только редактирует общий объект формы из App.vue: отправка и переход на карточку — в composable/store.
function submitCreateTicket() {
  emit('submit-create-ticket')
}

function openScreen(screenId) {
  emit('open-screen', screenId)
}

function setExecutionDate(value) {
  emit('set-execution-date', value)
}

function addAttachments(event) {
  emit('add-attachments', event?.target?.files || [])
  event.target.value = ''
}

function removeAttachment(index) {
  emit('remove-attachment', index)
}
</script>

<template>
  <section class="screen">
    <div class="section-header">
      <div>
        <p class="eyebrow">Новая заявка</p>
        <h2>Создание обращения</h2>
      </div>
      <span class="status-pill info">Типы заявки зависят от прав пользователя</span>
    </div>

    <form class="content-card form-card" @submit.prevent="submitCreateTicket">
      <label>
        Тип заявки
        <select
          v-model="createTicketForm.requestType"
          :class="{ 'field-invalid': createValidationStarted && createValidationErrors.requestType }"
        >
          <option v-for="requestType in availableRequestTypes" :key="requestType" :value="requestType">
            {{ requestType }}
          </option>
        </select>
        <small v-if="createValidationStarted && createValidationErrors.requestType" class="field-error">{{ createValidationErrors.requestType }}</small>
      </label>
      <label>
        Краткая тема
        <input
          v-model="createTicketForm.title"
          type="text"
          :class="{ 'field-invalid': createValidationStarted && createValidationErrors.title }"
        />
        <small v-if="createValidationStarted && createValidationErrors.title" class="field-error">{{ createValidationErrors.title }}</small>
      </label>
      <label>
        Подробное описание
        <textarea
          v-model="createTicketForm.description"
          rows="5"
          :class="{ 'field-invalid': createValidationStarted && createValidationErrors.description }"
        ></textarea>
        <small v-if="createValidationStarted && createValidationErrors.description" class="field-error">{{ createValidationErrors.description }}</small>
      </label>
      <label>
        Подразделение
        <select
          v-model="createTicketForm.department"
          :class="{ 'field-invalid': createValidationStarted && createValidationErrors.department }"
        >
          <option>Отдел ИТ</option>
          <option>Маркетинг</option>
        </select>
        <small v-if="createValidationStarted && createValidationErrors.department" class="field-error">{{ createValidationErrors.department }}</small>
      </label>
      <label>
        Исполнить до
        <input
          type="date"
          class="date-field"
          :class="{ 'field-invalid': createValidationStarted && createValidationErrors.executionDate }"
          :value="createTicketForm.executionDate || ''"
          @input="setExecutionDate($event.target.value)"
        />
        <small v-if="createValidationStarted && createValidationErrors.executionDate" class="field-error">{{ createValidationErrors.executionDate }}</small>
      </label>

      <p v-if="createErrorMessage" class="status-pill rose">{{ createErrorMessage }}</p>

      <div class="upload-box">
        <div>
          <strong>Вложения</strong>
          <p>Скриншоты, фото, документы, голосовые сообщения.</p>
        </div>
        <label class="secondary-button upload-trigger">
          Добавить файл
          <!-- accept: типичные вложения к заявке; при необходимости расширить под политику ИБ -->
          <input
            class="upload-input"
            type="file"
            multiple
            accept="image/*,audio/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.zip"
            @change="addAttachments"
          />
        </label>
      </div>

      <div class="chip-list">
        <div
          v-for="(file, index) in createTicketForm.attachmentFiles || []"
          :key="file.name + '-' + index"
          class="file-chip"
        >
          <span>{{ file.name }}</span>
          <button class="file-chip-remove" type="button" @click="removeAttachment(index)">Удалить</button>
        </div>
      </div>

      <div class="hero-actions">
        <button
          class="primary-button"
          type="submit"
          :disabled="isCreatingTicket"
        >
          {{ isCreatingTicket ? 'Отправка...' : 'Отправить заявку' }}
        </button>
        <button class="secondary-button" type="button" :disabled="isCreatingTicket" @click="openScreen('home')">Отмена</button>
      </div>
    </form>
  </section>
</template>
