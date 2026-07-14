<template>
  <BaseDialog :show="show" :title="t('admin.accounts.inviteResetTitle')" width="wide" :z-index="10000" @close="handleClose">
    <div v-if="account" class="space-y-5">
      <div class="flex flex-col gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-700 sm:flex-row sm:items-center sm:justify-between">
        <div class="min-w-0">
          <div class="truncate font-semibold text-gray-900 dark:text-white">{{ account.name }}</div>
          <div class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.accounts.inviteResetSubtitle') }}</div>
        </div>
        <button type="button" class="btn btn-secondary shrink-0" :disabled="loading" @click="loadStatus()">
          <Icon name="refresh" size="sm" :class="loading && 'animate-spin'" />
          {{ t('admin.accounts.inviteResetRefresh') }}
        </button>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)]">
          <section class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-dark-600">
            <div>
              <div class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.inviteResetAvailable') }}</div>
              <div class="mt-2 flex items-end gap-2">
                <span class="text-4xl font-bold text-gray-900 dark:text-white">{{ availableCountLabel }}</span>
                <span class="pb-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.accounts.inviteResetAvailableUnit') }}</span>
              </div>
            </div>

            <div v-if="hasNoCredits" class="rounded-lg border border-dashed border-gray-300 p-4 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
              {{ t('admin.accounts.inviteResetNoCredits') }}
            </div>

            <button type="button" class="btn btn-primary w-full" :disabled="!canConsume" @click="handleConsume">
              <Icon name="refresh" size="sm" :class="consuming && 'animate-spin'" />
              {{ consuming ? t('admin.accounts.inviteResetUsing') : t('admin.accounts.inviteResetUseReset') }}
            </button>

            <button
              type="button"
              class="flex w-full items-center justify-between rounded-lg border border-gray-200 px-3 py-2 text-left text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-dark-600 dark:text-gray-200 dark:hover:bg-dark-700"
              @click="showRules = !showRules"
            >
              <span>{{ t('admin.accounts.inviteResetRules') }}</span>
              <Icon name="chevronDown" size="sm" :class="['transition-transform', showRules && 'rotate-180']" />
            </button>
            <div v-if="showRules" class="rounded-lg bg-gray-50 p-3 text-sm text-gray-600 dark:bg-dark-800 dark:text-gray-300">
              <ul v-if="rules.length > 0" class="list-disc space-y-1 pl-5">
                <li v-for="rule in rules" :key="rule">{{ rule }}</li>
              </ul>
              <p v-else>{{ t('admin.accounts.inviteResetRulesEmpty') }}</p>
            </div>
          </section>

          <section class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-dark-600">
            <div>
              <label class="input-label" for="codex-invite-reset-emails">{{ t('admin.accounts.inviteResetInviteEmails') }}</label>
              <textarea
                id="codex-invite-reset-emails"
                v-model="emailInput"
                rows="8"
                class="input mt-2 min-h-[180px] resize-y font-mono text-sm leading-6"
                :placeholder="t('admin.accounts.inviteResetPlaceholder')"
              />
              <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.inviteResetEmailHint', { max: maxEmails }) }}</p>
            </div>

            <label class="flex items-start gap-2 rounded-lg border border-gray-200 p-3 text-sm text-gray-700 dark:border-dark-600 dark:text-gray-300">
              <input
                v-model="consentConfirmed"
                type="checkbox"
                class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <span>{{ t('admin.accounts.inviteResetConsent') }}</span>
            </label>

            <div v-if="message" :class="['rounded-lg p-3 text-sm', messageClass]">{{ message }}</div>

            <button type="button" class="btn btn-secondary w-full justify-center" :disabled="sendingInvite" @click="handleSendInvite">
              <Icon name="mail" size="sm" :class="sendingInvite && 'animate-pulse'" />
              {{ sendingInvite ? t('admin.accounts.inviteResetSending') : t('admin.accounts.inviteResetSendInvite') }}
            </button>
          </section>
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button type="button" class="btn btn-secondary" @click="handleClose">{{ t('common.close') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Account } from '@/types'
import type { CodexInviteResetStatus, OpenAIQuotaUsage } from '@/api/admin/accounts'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const maxEmails = 5
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  close: []
  updated: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const sendingInvite = ref(false)
const consuming = ref(false)
const status = ref<CodexInviteResetStatus | null>(null)
const quota = ref<OpenAIQuotaUsage | null>(null)
const emailInput = ref('')
const consentConfirmed = ref(false)
const message = ref('')
const messageType = ref<'success' | 'error' | ''>('')
const showRules = ref(false)

const availableCount = computed(() => quota.value?.rate_limit_reset_credits?.available_count ?? status.value?.available_count)
const availableCountLabel = computed(() => availableCount.value ?? '--')
const hasNoCredits = computed(() => availableCount.value !== undefined && availableCount.value <= 0)
const rules = computed(() => status.value?.eligibility_rules ?? [])

const canConsume = computed(() => {
  return (availableCount.value ?? 0) > 0 && !loading.value && !consuming.value
})

const messageClass = computed(() => {
  if (messageType.value === 'success') return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (messageType.value === 'error') return 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-300'
  return 'bg-gray-50 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
})

const parseEmails = () => {
  const emails = emailInput.value
    .split(/[,\s;]+/)
    .map((item) => item.trim())
    .filter(Boolean)
  const unique = [...new Map(emails.map((email) => [email.toLowerCase(), email])).values()]
  if (unique.length === 0) throw new Error(t('admin.accounts.inviteResetEmailsRequired'))
  if (unique.length > maxEmails) throw new Error(t('admin.accounts.inviteResetEmailLimit', { max: maxEmails }))
  const invalid = unique.find((email) => !emailPattern.test(email))
  if (invalid) throw new Error(t('admin.accounts.inviteResetInvalidEmail', { email: invalid }))
  return unique
}

const setMessage = (type: 'success' | 'error', text: string) => {
  messageType.value = type
  message.value = text
}

const loadStatus = async (clearMessage = true) => {
  if (!props.account) return
  loading.value = true
  if (clearMessage) {
    message.value = ''
    messageType.value = ''
  }
  try {
    const [inviteStatusResult, quotaUsageResult] = await Promise.allSettled([
      adminAPI.accounts.getCodexInviteResetStatus(props.account.id),
      adminAPI.accounts.getOpenAIQuota(props.account.id)
    ])
    if (inviteStatusResult.status === 'fulfilled') status.value = inviteStatusResult.value
    if (quotaUsageResult.status === 'fulfilled') quota.value = quotaUsageResult.value
    if (inviteStatusResult.status === 'rejected' && quotaUsageResult.status === 'rejected') {
      throw quotaUsageResult.reason || inviteStatusResult.reason
    }
    if (inviteStatusResult.status === 'rejected') {
      setMessage('error', inviteStatusResult.reason?.message || t('admin.accounts.inviteResetLoadFailed'))
    }
  } catch (error: any) {
    status.value = null
    quota.value = null
    setMessage('error', error?.message || t('admin.accounts.inviteResetLoadFailed'))
    appStore.showError(error?.message || t('admin.accounts.inviteResetLoadFailed'))
  } finally {
    loading.value = false
  }
}

const handleSendInvite = async () => {
  if (!props.account || sendingInvite.value) return
  try {
    const emails = parseEmails()
    if ((status.value?.requires_consent ?? true) && !consentConfirmed.value) {
      throw new Error(t('admin.accounts.inviteResetConsentRequired'))
    }
    sendingInvite.value = true
    const result = await adminAPI.accounts.sendCodexInviteResetInvite(props.account.id, emails)
    const failed = result.failed_emails?.filter(Boolean) ?? []
    if (failed.length > 0) {
      setMessage('error', t('admin.accounts.inviteResetInvitePartialFailed', { emails: failed.join(', ') }))
      return
    }
    emailInput.value = ''
    setMessage('success', result.message || t('admin.accounts.inviteResetInviteSuccess'))
    appStore.showSuccess(t('admin.accounts.inviteResetInviteSuccess'))
  } catch (error: any) {
    setMessage('error', error?.message || t('admin.accounts.inviteResetInviteFailed'))
  } finally {
    sendingInvite.value = false
  }
}

const consumeSuccessMessage = (code?: string) => {
  if (code === 'nothing_to_reset') return t('admin.accounts.inviteResetNothingToReset')
  if (code === 'already_redeemed') return t('admin.accounts.inviteResetAlreadyRedeemed')
  if (code === 'no_credit') return t('admin.accounts.inviteResetNoCredit')
  return t('admin.accounts.inviteResetConsumeSuccess')
}

const handleConsume = async () => {
  if (!props.account || consuming.value) return
  if (!window.confirm(t('admin.accounts.inviteResetConfirmUse', { name: props.account.name }))) return
  if (!window.confirm(t('admin.accounts.inviteResetConfirmUseAgain'))) return
  consuming.value = true
  try {
    const result = await adminAPI.accounts.resetOpenAIQuota(props.account.id)
    const text = consumeSuccessMessage(result.code)
    setMessage(result.code === 'reset' || !result.code ? 'success' : 'error', text)
    if (result.code === 'reset' || !result.code) {
      appStore.showSuccess(text)
      emit('updated')
    }
    await loadStatus(false)
  } catch (error: any) {
    setMessage('error', error?.message || t('admin.accounts.inviteResetConsumeFailed'))
  } finally {
    consuming.value = false
  }
}

const resetLocalState = () => {
  status.value = null
  quota.value = null
  emailInput.value = ''
  consentConfirmed.value = false
  message.value = ''
  messageType.value = ''
  showRules.value = false
}

const handleClose = () => {
  emit('close')
}

watch(
  () => [props.show, props.account?.id],
  ([visible]) => {
    if (visible && props.account) {
      loadStatus()
      return
    }
    resetLocalState()
  }
)
</script>
