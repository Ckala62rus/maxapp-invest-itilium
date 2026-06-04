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
  marketingErrorMessage: {
    type: String,
    required: true
  },
  marketingServices: {
    type: Array,
    required: true
  },
  marketingSubdivisions: {
    type: Array,
    required: true
  },
  isLoadingMarketingServices: {
    type: Boolean,
    required: true
  },
  isLoadingMarketingSubdivisions: {
    type: Boolean,
    required: true
  },
  isCreatingMarketingRequest: {
    type: Boolean,
    required: true
  },
  marketingFormData: {
    type: Object,
    required: true
  },
  selectedMarketingService: {
    type: Object,
    required: false,
    default: null
  },
  currentMarketingSchema: {
    type: Object,
    required: false,
    default: null
  },
  isCreatingTicket: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits([
  'submit-create-ticket',
  'open-screen',
  'set-execution-date',
  'add-attachments',
  'remove-attachment',
  'set-marketing-service',
  'set-marketing-field'
])

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

function setMarketingService(value) {
  emit('set-marketing-service', value)
}

function setMarketingField(key, value) {
  emit('set-marketing-field', key, value)
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
      <template v-if="createTicketForm.requestType !== 'Маркетинговая заявка'">
        <label>
          Тема
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
      </template>

      <template v-else>
        <label>
          Тип маркетинговой заявки
          <select
            :value="selectedMarketingService?.code || ''"
            :disabled="isLoadingMarketingServices || isCreatingMarketingRequest"
            :class="{ 'field-invalid': createValidationStarted && createValidationErrors.serviceCode }"
            @change="setMarketingService($event.target.value)"
          >
            <option value="" disabled>Выберите тип</option>
            <option v-for="service in marketingServices" :key="service.code" :value="service.code">
              {{ service.name }} (форма {{ service.formNumber || '—' }})
            </option>
          </select>
          <small v-if="createValidationStarted && createValidationErrors.serviceCode" class="field-error">{{ createValidationErrors.serviceCode }}</small>
        </label>

        <label>
          Выбор подразделения <span>*</span>
          <select
            v-if="marketingSubdivisions.length > 0"
            v-model="createTicketForm.department"
            :disabled="isLoadingMarketingSubdivisions || isCreatingMarketingRequest"
            :class="{ 'field-invalid': createValidationStarted && createValidationErrors.department }"
          >
            <option value="" disabled>Выберите подразделение</option>
            <option v-for="subdivision in marketingSubdivisions" :key="subdivision.code || subdivision.name" :value="subdivision.name">
              {{ subdivision.name }}
            </option>
          </select>
          <input
            v-else
            v-model="createTicketForm.department"
            type="text"
            :disabled="isLoadingMarketingSubdivisions || isCreatingMarketingRequest"
            :class="{ 'field-invalid': createValidationStarted && createValidationErrors.department }"
            placeholder="Укажите подразделение"
          />
          <small v-if="createValidationStarted && createValidationErrors.department" class="field-error">{{ createValidationErrors.department }}</small>
        </label>

        <label>
          Дата исполнения <span>*</span>
          <el-date-picker
            class="date-field"
            :class="{ 'field-invalid': createValidationStarted && createValidationErrors.executionDate }"
            :model-value="createTicketForm.executionDate"
            type="date"
            value-format="YYYY-MM-DD"
            format="DD.MM.YYYY"
            placeholder="Выберите дату"
            :disabled="isCreatingMarketingRequest"
            :teleported="false"
            @update:model-value="setExecutionDate"
          />
          <small v-if="createValidationStarted && createValidationErrors.executionDate" class="field-error">{{ createValidationErrors.executionDate }}</small>
        </label>

        <div class="content-card" style="padding: 12px">
          <p class="eyebrow">Параметры заявки</p>
          <p v-if="currentMarketingSchema?.formNumber">Форма № {{ currentMarketingSchema.formNumber }}</p>
          <div v-for="field in currentMarketingSchema?.fields || []" :key="field.key">
            <label>
              {{ field.label }} <span v-if="field.required">*</span>
              <select
                v-if="field.type === 'select'"
                :value="marketingFormData[field.key] || ''"
                @change="setMarketingField(field.key, $event.target.value)"
              >
                <option value="" disabled>Выберите вариант</option>
                <option v-for="option in field.options || []" :key="field.key + option" :value="option">{{ option }}</option>
              </select>
              <textarea
                v-else-if="field.type === 'textarea'"
                rows="4"
                :value="marketingFormData[field.key] || ''"
                @input="setMarketingField(field.key, $event.target.value)"
              ></textarea>
              <input
                v-else
                type="text"
                :value="marketingFormData[field.key] || ''"
                @input="setMarketingField(field.key, $event.target.value)"
              />
              <small v-if="createValidationStarted && createValidationErrors['form_' + field.key]" class="field-error">
                {{ createValidationErrors['form_' + field.key] }}
              </small>
            </label>
          </div>
        </div>
      </template>

      <p v-if="isCreatingTicket || isCreatingMarketingRequest" class="status-pill">
        Создание в ITILIUM может занять до минуты — не закрывайте приложение.
      </p>

      <p v-if="createErrorMessage || marketingErrorMessage" class="status-pill rose">{{ createErrorMessage || marketingErrorMessage }}</p>

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
          :disabled="isCreatingTicket || isCreatingMarketingRequest"
        >
          {{ (isCreatingTicket || isCreatingMarketingRequest) ? 'Отправка...' : 'Отправить заявку' }}
        </button>
        <button class="secondary-button" type="button" :disabled="isCreatingTicket || isCreatingMarketingRequest" @click="openScreen('home')">Отмена</button>
      </div>
    </form>
  </section>
</template>
