<script setup>
/**
 * Корневой layout мини-приложения: после onMounted вызывается bootstrapAuth,
 * дальше activeScreen переключает экраны (прототип навигации без vue-router).
 */
import { computed, onMounted, ref, watch } from 'vue'
import { useStore } from 'vuex'

import HomeScreen from '@/screens/HomeScreen.vue'
import ProfileScreen from '@/screens/ProfileScreen.vue'
import RegistrationScreen from '@/screens/RegistrationScreen.vue'
import CreateTicketScreen from '@/screens/CreateTicketScreen.vue'
import MyTicketsScreen from '@/screens/MyTicketsScreen.vue'
import ResponsibleTicketsScreen from '@/screens/ResponsibleTicketsScreen.vue'
import SearchTicketScreen from '@/screens/SearchTicketScreen.vue'
import TicketDetailsScreen from '@/screens/TicketDetailsScreen.vue'
import { useAuthFlow } from '@/composables/useAuthFlow'
import { useTicketFlow } from '@/composables/useTicketFlow'

const store = useStore()

// Screen ids are used by the internal app navigator after onboarding is complete.
const screenOptions = [
  { id: 'home', label: 'Главная' },
  { id: 'profile', label: 'Профиль' },
  { id: 'registration', label: 'Регистрация' },
  { id: 'create', label: 'Создать заявку' },
  { id: 'myTickets', label: 'Мои заявки' },
  { id: 'responsible', label: 'В ответственности' },
  { id: 'search', label: 'Поиск' },
  { id: 'details', label: 'Карточка заявки' }
]

// The active screen simulates page transitions for the stakeholder demo.
const activeScreen = ref('home')

// The submission banner imitates the visual result of a completed action.
const submitBanner = ref('')

const showPrototypeNavigation = computed(() => {
  return Boolean(currentUser.value?.employeeFound) && !currentUser.value?.registrationRequired
})

const {
  registrationForm,
  currentUser,
  currentIdentity,
  isRegistrationSubmitting,
  isAuthBootstrapping,
  authErrors,
  profileInitials,
  profileStatusText,
  profileRegion,
  registrationIdentityUserId,
  maxBridgeState,
  rawInitData,
  rawInitDataUnsafeUserId,
  bootstrapAuth,
  submitRegistration,
  registrationValidationStarted,
  registrationValidationErrors
} = useAuthFlow({
  store,
  activeScreen,
  submitBanner
})

const {
  searchQuery,
  currentTicketsPage,
  commentDraft,
  commentAttachmentFiles,
  commentSuccessTick,
  ratingSuccessTick,
  detailsOrigin,
  statusForm,
  selectedResponsibleId,
  createTicketForm,
  availableRequestTypes,
  isLoadingMyTickets,
  isLoadingResponsibleTickets,
  isCreatingTicket,
  isLoadingTicketDetails,
  isLoadingResponsibleOptions,
  isSubmittingComment,
  isChangingStatus,
  isChangingResponsible,
  isSubmittingTicketRating,
  listErrors,
  ticketErrors,
  createErrors,
  createValidationErrors,
  createValidationStarted,
  createErrorMessage,
  normalizedResponsibleTickets,
  summaryCards,
  paginatedTickets,
  pageCount,
  selectedTicket,
  detailStatusTone,
  availableStatusOptions,
  availableResponsibleOptions,
  selectedTicketTimeline,
  loadTicketLists,
  openTicketDetails,
  searchTicketByNumber,
  submitCreateTicket,
  setCreateExecutionDate,
  addCreateAttachments,
  removeCreateAttachment,
  submitComment,
  addCommentAttachments,
  removeCommentAttachment,
  submitStatusChange,
  assignResponsible,
  requestResponsibleOptions,
  submitTicketRating,
  setTicketsPage,
  setSearchQuery,
  setCommentDraft
} = useTicketFlow({
  store,
  currentUser,
  activeScreen,
  submitBanner
})

onMounted(() => {
  bootstrapAuth().then((response) => {
    const user = response?.data?.user || null
    if (response?.data?.stage === 'ready' && user?.employeeFound) {
      loadTicketLists()
    }
  })
})

// This helper switches screens from the top navigation and action buttons.
function openScreen(screenId) {
  activeScreen.value = screenId
  submitBanner.value = ''
}

function openMyTicketDetails(ticketNumber) {
  openTicketDetails(ticketNumber, 'myTickets')
}

function openResponsibleTicketDetails(ticketNumber) {
  openTicketDetails(ticketNumber, 'responsible')
}

</script>

<template>
  <main class="phone-stage phone-stage--standalone">
      <div class="phone-frame">
        <header class="app-header">
          <div>
            <p class="eyebrow">MAX x ITILIUM</p>
            <strong>Сервисные заявки</strong>
            <p v-if="registrationIdentityUserId" class="eyebrow">MAX ID: {{ registrationIdentityUserId }}</p>
          </div>
          <button
            v-if="showPrototypeNavigation"
            class="ghost-button"
            @click="openScreen('home')"
          >
            Домой
          </button>
        </header>

        <div v-if="isAuthBootstrapping" class="submit-banner">
          Проверяем MAX-сессию...
        </div>

        <div v-if="showPrototypeNavigation" class="tab-strip">
          <button
            v-for="screen in screenOptions"
            :key="screen.id"
            class="tab-button"
            :class="{ active: activeScreen === screen.id }"
            @click="openScreen(screen.id)"
          >
            {{ screen.label }}
          </button>
        </div>

        <HomeScreen
          v-if="activeScreen === 'home'"
          :summary-cards="summaryCards"
          :max-bridge-state="maxBridgeState"
          :raw-init-data="rawInitData"
          :raw-init-data-unsafe-user-id="rawInitDataUnsafeUserId"
          @open-screen="openScreen"
        />

        <ProfileScreen
          v-else-if="activeScreen === 'profile'"
          :current-user="currentUser"
          :current-identity="currentIdentity"
          :profile-initials="profileInitials"
          :profile-status-text="profileStatusText"
          :profile-region="profileRegion"
          @open-screen="openScreen"
        />

        <RegistrationScreen
          v-else-if="activeScreen === 'registration'"
          :current-identity="currentIdentity"
          :current-user="currentUser"
          :registration-form="registrationForm"
          :max-bridge-state="maxBridgeState"
          :raw-init-data="rawInitData"
          :raw-init-data-unsafe-user-id="rawInitDataUnsafeUserId"
          :auth-errors="authErrors"
          :is-registration-submitting="isRegistrationSubmitting"
          :registration-validation-started="registrationValidationStarted"
          :registration-validation-errors="registrationValidationErrors"
          @submit-registration="submitRegistration"
        />

        <CreateTicketScreen
          v-else-if="activeScreen === 'create'"
          :create-ticket-form="createTicketForm"
          :create-errors="createErrors"
          :create-validation-errors="createValidationErrors"
          :create-validation-started="createValidationStarted"
          :create-error-message="createErrorMessage"
          :is-creating-ticket="isCreatingTicket"
          :available-request-types="availableRequestTypes"
          @submit-create-ticket="submitCreateTicket"
          @set-execution-date="setCreateExecutionDate"
          @add-attachments="addCreateAttachments"
          @remove-attachment="removeCreateAttachment"
          @open-screen="openScreen"
        />

        <MyTicketsScreen
          v-else-if="activeScreen === 'myTickets'"
          :is-loading-my-tickets="isLoadingMyTickets"
          :list-errors="listErrors"
          :paginated-tickets="paginatedTickets"
          :page-count="pageCount"
          :current-tickets-page="currentTicketsPage"
          @open-ticket-details="openMyTicketDetails"
          @set-tickets-page="setTicketsPage"
        />

        <ResponsibleTicketsScreen
          v-else-if="activeScreen === 'responsible'"
          :is-loading-responsible-tickets="isLoadingResponsibleTickets"
          :list-errors="listErrors"
          :normalized-responsible-tickets="normalizedResponsibleTickets"
          @open-ticket-details="openResponsibleTicketDetails"
        />

        <SearchTicketScreen
          v-else-if="activeScreen === 'search'"
          :search-query="searchQuery"
          :ticket-errors="ticketErrors"
          :is-loading-ticket-details="isLoadingTicketDetails"
          @update:search-query="setSearchQuery"
          @search-ticket="searchTicketByNumber"
        />

        <TicketDetailsScreen
          v-else-if="activeScreen === 'details'"
          :selected-ticket="selectedTicket"
          :search-query="searchQuery"
          :detail-status-tone="detailStatusTone"
          :is-loading-ticket-details="isLoadingTicketDetails"
          :ticket-errors="ticketErrors"
          :details-origin="detailsOrigin"
          :selected-ticket-timeline="selectedTicketTimeline"
          :comment-draft="commentDraft"
          :comment-attachment-files="commentAttachmentFiles"
          :comment-success-tick="commentSuccessTick"
          :rating-success-tick="ratingSuccessTick"
          :is-submitting-comment="isSubmittingComment"
          :status-form="statusForm"
          :available-status-options="availableStatusOptions"
          :is-changing-status="isChangingStatus"
          :is-loading-responsible-options="isLoadingResponsibleOptions"
          :available-responsible-options="availableResponsibleOptions"
          :is-changing-responsible="isChangingResponsible"
          :selected-responsible-id="selectedResponsibleId"
          :is-submitting-ticket-rating="isSubmittingTicketRating"
          @open-screen="openScreen"
          @update:comment-draft="setCommentDraft"
          @add-comment-files="addCommentAttachments"
          @remove-comment-file="removeCommentAttachment"
          @submit-comment="submitComment"
          @submit-status-change="submitStatusChange"
          @assign-responsible="assignResponsible"
          @request-responsible-options="requestResponsibleOptions"
          @submit-ticket-rating="submitTicketRating"
        />

        <transition name="banner">
          <div v-if="submitBanner" class="submit-banner">
            {{ submitBanner }}
          </div>
        </transition>
      </div>
    </main>
</template>
