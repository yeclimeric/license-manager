import { computed, getCurrentScope, onScopeDispose, ref, watch, type Ref } from 'vue'

export const useInitialSkeleton = (loading: Ref<boolean>, delay = 120) => {
  const initialized = ref(false)
  const skeletonVisible = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined

  const clearTimer = () => {
    if (timer) clearTimeout(timer)
    timer = undefined
  }

  watch(
    loading,
    isLoading => {
      clearTimer()
      if (isLoading && !initialized.value) {
        timer = setTimeout(() => {
          skeletonVisible.value = true
        }, delay)
      } else {
        skeletonVisible.value = false
      }
    },
    { immediate: true }
  )

  const markInitialized = () => {
    initialized.value = true
    skeletonVisible.value = false
    clearTimer()
  }

  // 组件内必有 effect scope；纯单测直接调用时可没有，
  // 此时 onScopeDispose 会告警且不会注册清理，故显式判断。
  if (getCurrentScope()) {
    onScopeDispose(clearTimer)
  }

  return {
    initialized,
    skeletonVisible,
    isUpdating: computed(() => loading.value && initialized.value),
    markInitialized
  }
}
