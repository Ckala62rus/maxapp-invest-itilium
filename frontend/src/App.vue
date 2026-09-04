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
import { purgeTouchBlockers } from '@/helpers/purgeTouchBlockers'
import { getMaxStartParam, parseTicketNumberFromStartParam } from '@/api/maxBridge'
import { getSessionItem, setSessionItem } from '@/helpers/persistenceStorage'

const store = useStore()
const ACTIVE_SCREEN_KEY = 'maxapp.activeScreen'

// Screen ids are used by the internal app navigator after onboarding is complete.
const baseScreenOptions = [
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
function readStoredActiveScreen() {
  const stored = getSessionItem(ACTIVE_SCREEN_KEY)
  if (stored && baseScreenOptions.some((screen) => screen.id === stored)) {
    return stored
  }
  return 'home'
}

const activeScreen = ref(readStoredActiveScreen())

watch(activeScreen, (screenId) => {
  setSessionItem(ACTIVE_SCREEN_KEY, screenId)
})

// The submission banner imitates the visual result of a completed action.
const submitBanner = ref('')
const isNavigationOpen = ref(false)
const isDebugUiEnabled = computed(() => {
  return import.meta.env.DEV || import.meta.env.VITE_DEBUG_UI === 'true'
})

const screenOptions = computed(() => {
  return baseScreenOptions.filter((screen) => isDebugUiEnabled.value || screen.id !== 'registration')
})

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
  responsibleSuccessTick,
  detailsOrigin,
  statusForm,
  selectedResponsibleId,
  createTicketForm,
  marketingFormData,
  selectedMarketingService,
  availableRequestTypes,
  isLoadingMyTickets,
  isLoadingResponsibleTickets,
  isCreatingTicket,
  isLoadingTicketDetails,
  isLoadingTicketComments,
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
  marketingErrorMessage,
  marketingServices,
  marketingSubdivisions,
  isLoadingMarketingServices,
  isLoadingMarketingSubdivisions,
  isCreatingMarketingRequest,
  currentMarketingSchema,
  normalizedResponsibleTickets,
  paginatedTickets,
  pageCount,
  selectedTicket,
  detailStatusTone,
  availableStatusOptions,
  availableResponsibleOptions,
  selectedTicketTimeline,
  paginatedTicketComments,
  commentsPageCount,
  currentCommentsPage,
  loadTicketLists,
  openTicketDetails,
  searchTicketByNumber,
  submitCreateTicket,
  setMarketingService,
  setMarketingFieldValue,
  setCreateExecutionDate,
  addCreateAttachments,
  createAttachmentsPreparing,
  removeCreateAttachment,
  submitComment,
  addCommentAttachments,
  removeCommentAttachment,
  submitStatusChange,
  assignResponsible,
  requestResponsibleOptions,
  submitTicketRating,
  setTicketsPage,
  setCommentsPage,
  setSearchQuery,
  setCommentDraft
} = useTicketFlow({
  store,
  currentUser,
  activeScreen,
  submitBanner
})

onMounted(() => {
  purgeTouchBlockers()
  // Номер из ?startapp= читаем до bootstrap: после auth bridge/query всё ещё доступны, но так проще отладить.
  const startupTicket = parseTicketNumberFromStartParam(getMaxStartParam())
  if (startupTicket && import.meta.env.DEV) {
    console.info('[nav] startup deep link ticket', { ticketNumber: startupTicket })
  }

  bootstrapAuth().then(async (response) => {
    const user = response?.data?.user || null
    const stage = response?.data?.stage
    const bootstrapOk = Boolean(response?.data?.success)
    const ticketNumber =
      startupTicket || parseTicketNumberFromStartParam(getMaxStartParam())

    // Deep link: после успешного auth открываем карточку (в DEV — даже если employeeFound ещё не подтянулся).
    if (
      ticketNumber &&
      bootstrapOk &&
      user &&
      (user.employeeFound || import.meta.env.DEV)
    ) {
      await openTicketDetails(ticketNumber, 'search')
      return
    }

    if (ticketNumber && import.meta.env.DEV) {
      console.warn('[nav] deep link skipped', {
        ticketNumber,
        bootstrapOk,
        stage,
        employeeFound: user?.employeeFound ?? null
      })
    }

    if (stage === 'ready' && user?.employeeFound) {
      loadTicketLists()
    }
  })
})

// This helper switches screens from the top navigation and action buttons.
function openScreen(screenId) {
  activeScreen.value = screenId
  // Сразу пишем в sessionStorage: при HMR/full reload @vite/client экран восстанавливается до срабатывания watch.
  setSessionItem(ACTIVE_SCREEN_KEY, screenId)
  submitBanner.value = ''
  isNavigationOpen.value = false
}

function toggleNavigation() {
  isNavigationOpen.value = !isNavigationOpen.value
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
          <div class="app-header-brand">
            <p class="eyebrow">MAX x ITILIUM</p>
            <strong>Сервисные заявки</strong>
            <p v-if="registrationIdentityUserId" class="eyebrow">MAX ID: {{ registrationIdentityUserId }}</p>
          </div>
          <div v-if="showPrototypeNavigation" class="app-header-nav">
            <button
              class="menu-toggle"
              type="button"
              :aria-expanded="isNavigationOpen ? 'true' : 'false'"
              aria-controls="app-navigation-menu"
              @click="toggleNavigation"
            >
              <span></span>
              <span></span>
              <span></span>
              <strong>Меню</strong>
            </button>
            <transition name="menu">
              <nav
                v-if="isNavigationOpen"
                id="app-navigation-menu"
                class="burger-menu"
                aria-label="Навигация по приложению"
              >
                <button
                  v-for="screen in screenOptions"
                  :key="screen.id"
                  class="burger-menu-item"
                  :class="{ active: activeScreen === screen.id }"
                  type="button"
                  @click="openScreen(screen.id)"
                >
                  {{ screen.label }}
                </button>
              </nav>
            </transition>
          </div>
        </header>

        <div v-if="isAuthBootstrapping" class="submit-banner">
          Проверяем MAX-сессию...
        </div>

        <HomeScreen
          v-if="activeScreen === 'home'"
          :max-bridge-state="maxBridgeState"
          :raw-init-data="rawInitData"
          :raw-init-data-unsafe-user-id="rawInitDataUnsafeUserId"
          :show-debug-info="isDebugUiEnabled"
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
          :show-debug-info="isDebugUiEnabled"
          @submit-registration="submitRegistration"
        />

        <CreateTicketScreen
          v-else-if="activeScreen === 'create'"
          :create-ticket-form="createTicketForm"
          :create-errors="createErrors"
          :create-validation-errors="createValidationErrors"
          :create-validation-started="createValidationStarted"
          :create-error-message="createErrorMessage"
          :marketing-error-message="marketingErrorMessage"
          :marketing-services="marketingServices"
          :marketing-subdivisions="marketingSubdivisions"
          :is-loading-marketing-services="isLoadingMarketingServices"
          :is-loading-marketing-subdivisions="isLoadingMarketingSubdivisions"
          :is-creating-marketing-request="isCreatingMarketingRequest"
          :marketing-form-data="marketingFormData"
          :selected-marketing-service="selectedMarketingService"
          :current-marketing-schema="currentMarketingSchema"
          :is-creating-ticket="isCreatingTicket"
          :create-attachments-preparing="createAttachmentsPreparing"
          :available-request-types="availableRequestTypes"
          @submit-create-ticket="submitCreateTicket"
          @set-marketing-service="setMarketingService"
          @set-marketing-field="setMarketingFieldValue"
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
          :paginated-ticket-comments="paginatedTicketComments"
          :comments-page-count="commentsPageCount"
          :current-comments-page="currentCommentsPage"
          :is-loading-ticket-comments="isLoadingTicketComments"
          :comment-draft="commentDraft"
          :comment-attachment-files="commentAttachmentFiles"
          :comment-success-tick="commentSuccessTick"
          :rating-success-tick="ratingSuccessTick"
          :responsible-success-tick="responsibleSuccessTick"
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
          @set-comments-page="setCommentsPage"
        />

        <transition name="banner">
          <div v-if="submitBanner" class="submit-banner">
            {{ submitBanner }}
          </div>
        </transition>
      </div>
    </main>
</template>
