<script setup>
// Создание заявки: тип (ИТ / маркетинг / DAX по флагам профиля), текст, срок.
import { ElDatePicker } from 'element-plus'

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

    <article class="content-card form-card">
      <label>
        Тип заявки
        <select v-model="createTicketForm.requestType">
          <option v-for="requestType in availableRequestTypes" :key="requestType" :value="requestType">
            {{ requestType }}
          </option>
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
        <ElDatePicker
          :model-value="createTicketForm.executionDate || null"
          type="date"
          format="DD.MM.YYYY"
          value-format="YYYY-MM-DD"
          placeholder="Выберите дату"
          @update:model-value="setExecutionDate"
        />
      </label>

      <p v-if="createErrors.length" class="status-pill rose">{{ createErrors[0] }}</p>

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
          :disabled="isCreatingTicket || !createTicketForm.title || !createTicketForm.description || !createTicketForm.executionDate"
          @click="submitCreateTicket"
        >
          {{ isCreatingTicket ? 'Отправка...' : 'Отправить заявку' }}
        </button>
        <button class="secondary-button" :disabled="isCreatingTicket" @click="openScreen('home')">Отмена</button>
      </div>
    </article>
  </section>
</template>
