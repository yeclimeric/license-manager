<template>
  <div class="form-action-dock">
    <footer class="form-action-bar">
      <el-button class="form-action-bar__cancel" :disabled="loading" @click="emit('cancel')">
        {{ cancelLabel }}
      </el-button>
      <span v-if="hint" class="form-action-bar__hint">{{ hint }}</span>
      <div class="form-action-bar__spacer" />
      <div class="form-action-bar__buttons">
        <!-- extra：提交前的附加动作（如向导的上一步） -->
        <slot name="extra" />
        <el-button type="primary" :loading="loading" :disabled="disabled" @click="emit('submit')">
          {{ submitLabel }}
        </el-button>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  cancelLabel: string
  submitLabel: string
  hint?: string
  loading?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{ cancel: []; submit: [] }>()
</script>

<style scoped>
/* 底部固定操作栏：取消靠左，附加动作与主提交按钮靠右。
   所有页面级表单（含向导）共用，保证操作位置一致。 */
.form-action-dock { height:76px; flex:0 0 76px; }
.form-action-bar { position:fixed; left:var(--layout-sidebar-width); right:0; bottom:0; z-index:1900; min-height:64px; padding:12px var(--layout-content-padding); display:flex; align-items:center; gap:20px; background:color-mix(in srgb,var(--app-content-bg) 86%,#c8ecde); border-top:1px solid #bfd7cd; box-shadow:0 -8px 22px rgba(25,82,66,.07); backdrop-filter:blur(10px); transition:left .2s ease; }
.form-action-bar__cancel { flex-shrink:0; min-width:92px; }
.form-action-bar__hint { color:#667b72; font-size:12px; }
.form-action-bar__spacer { flex:1; min-width:0; }
.form-action-bar__buttons { display:flex; align-items:center; gap:8px; }
.form-action-bar__buttons :deep(.el-button) { min-width:92px; }
[data-theme='dark'] .form-action-bar { border-color:var(--app-border-color); background:color-mix(in srgb,var(--app-content-bg) 88%,#145848); }
:global(.app-layout--collapsed) .form-action-bar { left:var(--layout-sidebar-collapsed-width); }
@media (min-width:768px) and (max-width:1023px) {
  .form-action-bar,
  :global(.app-layout--collapsed) .form-action-bar { left:var(--layout-sidebar-collapsed-width); }
}
@media (max-width:767px) {
  .form-action-dock { height:72px; flex-basis:72px; }
  .form-action-bar,
  :global(.app-layout--collapsed) .form-action-bar { left:0; right:0; bottom:0; padding:10px 12px; }
  .form-action-bar__hint { display:none; }
  .form-action-bar__spacer { display:none; }
  .form-action-bar__cancel { flex:1; }
  .form-action-bar__buttons { flex:1; }
  .form-action-bar__buttons :deep(.el-button) { flex:1; }
}
</style>
