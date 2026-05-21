import { ref, computed } from 'vue'

export type RiskLevel = 'safe' | 'caution' | 'dangerous'

interface ConfirmationCallbacks {
  onConfirm: () => Promise<void>
  onCancel?: () => void
}

/** Manages confirmation state for destructive actions. Supports typed confirmation for high-risk operations. */
export function useActionConfirmation() {
  /* ── State ────────────────────────────────────────────────── */

  const showConfirmModal = ref(false)
  const executingConfirm = ref(false)
  const pendingConfirmMessage = ref('')
  const pendingActionName = ref('')
  const pendingActionParams = ref<Record<string, string>>({})
  const pendingRiskLevel = ref<RiskLevel>('safe')
  const pendingConfirmationToken = ref('')
  const typedConfirmationInput = ref('')

  let pendingCallbacks: ConfirmationCallbacks | null = null

  /* ── Computed ─────────────────────────────────────────────── */

  const typedConfirmationRequired = computed(() => pendingRiskLevel.value === 'dangerous')

  const typedConfirmationExpected = computed(() => 'CONFIRM')

  const isConfirmDisabled = computed(() => {
    if (executingConfirm.value) return true
    if (typedConfirmationRequired.value && typedConfirmationInput.value !== 'CONFIRM') return true
    return false
  })

  /* ── Actions ──────────────────────────────────────────────── */

  function triggerConfirmation(
    actionName: string,
    params: Record<string, string>,
    message: string,
    riskLevel: RiskLevel,
    token: string,
    onConfirm: () => Promise<void>,
    onCancel?: () => void,
  ): void {
    pendingActionName.value = actionName
    pendingActionParams.value = params
    pendingConfirmMessage.value = message
    pendingRiskLevel.value = riskLevel
    pendingConfirmationToken.value = token
    typedConfirmationInput.value = ''
    executingConfirm.value = false
    pendingCallbacks = { onConfirm, onCancel }
    showConfirmModal.value = true
  }

  async function confirmAction(): Promise<void> {
    if (isConfirmDisabled.value || !pendingCallbacks) return
    executingConfirm.value = true

    try {
      await pendingCallbacks.onConfirm()
    } finally {
      executingConfirm.value = false
      closeModal()
    }
  }

  function cancelAction(): void {
    pendingCallbacks?.onCancel?.()
    closeModal()
  }

  function closeModal(): void {
    showConfirmModal.value = false
    executingConfirm.value = false
    typedConfirmationInput.value = ''
    pendingCallbacks = null
  }

  /* ── Expose ───────────────────────────────────────────────── */

  return {
    // State
    showConfirmModal,
    executingConfirm,
    pendingConfirmMessage,
    pendingActionName,
    pendingActionParams,
    pendingRiskLevel,
    pendingConfirmationToken,
    typedConfirmationInput,

    // Computed
    typedConfirmationRequired,
    typedConfirmationExpected,
    isConfirmDisabled,

    // Actions
    triggerConfirmation,
    confirmAction,
    cancelAction,
  }
}
