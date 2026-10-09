<template>
  <main class="login-page">
    <section class="brand-panel">
      <div class="license-motion" aria-hidden="true">
        <div class="motion-orbit motion-orbit--outer"></div>
        <div class="motion-orbit motion-orbit--inner"></div>
        <div class="motion-core">
          <span></span><span></span><span></span>
        </div>
        <div class="motion-signal motion-signal--one"></div>
        <div class="motion-signal motion-signal--two"></div>
        <div class="motion-signal motion-signal--three"></div>
      </div>
      <div class="brand-mark">
        <img v-if="publicConfig.logoUrl" :src="publicConfig.logoUrl" :alt="siteName" class="brand-logo-img" />
        <svg v-else viewBox="0 0 41 40" fill="none" aria-hidden="true">
          <path d="M26.3125 11.4814L22.25 19.5947V22.7148L27.1191 13.0576L29.7393 18.1777L18.7988 40H0L13.5938 22.8037H14L13.8125 23.1201L7.46875 33.541H17.8438V16.9111L7.1875 25.6475L18.0312 10.1406L11.625 14.4463V14.2588L20.4375 0L26.3125 11.4814Z" fill="currentColor"/>
          <path d="M34.5498 39.9996H28.75L24.5938 32.8864L27.125 27.6246L34.5498 39.9996ZM41 39.9996H36.2705L27.9346 25.941L30.7188 20.1559L41 39.9996Z" fill="currentColor" opacity=".7"/>
        </svg>
        <span>{{ siteName }}</span>
      </div>
      <div class="brand-copy">
        <p class="eyebrow">{{ t('login.communityEdition') }}</p>
        <h1>{{ t('login.heroTitleLine1') }}{{ t('login.heroTitleLine2') }}</h1>
        <p>{{ t('login.heroSubtitle') }}</p>
        <div class="authorization-flow" aria-hidden="true">
          <span>{{ t('login.flow.issue') }}</span>
          <i></i>
          <span>{{ t('login.flow.activate') }}</span>
          <i></i>
          <span>{{ t('login.flow.verify') }}</span>
        </div>
      </div>
      <div class="brand-footer">
        <p class="brand-note">{{ t('login.brandNote') }}</p>
        <a href="https://cedar-v.com/products/cedar-license-cloud/index.html" target="_blank" rel="noopener noreferrer">
          {{ t('login.cloudPrompt') }} <span>{{ t('login.cloudLink') }} →</span>
        </a>
      </div>
    </section>

    <section class="login-panel">
      <div class="language-switcher">
        <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <circle cx="12" cy="12" r="8.5" />
          <path d="M3.8 12h16.4M12 3.5c2.2 2.3 3.3 5.1 3.3 8.5S14.2 18.2 12 20.5M12 3.5C9.8 5.8 8.7 8.6 8.7 12s1.1 6.2 3.3 8.5" />
        </svg>
        <el-select v-model="currentLanguage" size="small" @change="handleLanguageChange">
          <el-option label="中文" value="zh" />
          <el-option label="English" value="en" />
          <el-option label="日本語" value="ja" />
        </el-select>
      </div>
      <div class="login-card">
        <header>
          <h2>{{ t('login.welcomeTitle') }}</h2>
        </header>
        <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" label-position="top" class="login-form" @submit.prevent="handleLogin">
          <el-form-item prop="username" :label="t('login.usernameLabel')">
            <el-input
              v-model="loginForm.username"
              :prefix-icon="User"
              size="large"
              clearable
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item prop="password" :label="t('login.passwordLabel')">
            <el-input
              v-model="loginForm.password"
              type="password"
              :prefix-icon="Lock"
              size="large"
              show-password
              autocomplete="current-password"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <div class="form-options">
            <el-checkbox v-model="rememberMe">{{ t('login.remember') }}</el-checkbox>
          </div>
          <el-button type="primary" native-type="submit" size="large" class="login-button" :loading="loading">
            {{ t('login.submit') }}
          </el-button>
        </el-form>
        <footer>{{ t('login.copyright', { year: currentYear }) }}</footer>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { Login, type LoginRequest } from '@/api/user'
import { getPublicConfig } from '@/api/systemConfig'
import { useUserStore } from '@/store/modules/user'
import { changeLanguage, type SupportedLocale } from '@/utils/language'

const REMEMBER_KEY = 'loginInfo'
const DEFAULT_COMMUNITY_LOGIN = {
  username: 'admin',
  password: 'admin@123'
}
const { t, locale } = useI18n()

// 站点名与 Logo 来自公开配置：登录页在未认证时也要能读取。
// 取不到时退回内置品牌，不阻塞登录。
const publicConfig = reactive({ siteName: 'Cedar-V', logoUrl: '' })
const siteName = computed(() => publicConfig.siteName)

const loadPublicConfig = async () => {
  try {
    const res = await getPublicConfig()
    const configs = res.data || {}
    if (configs['system.site_name']) publicConfig.siteName = configs['system.site_name']
    if (configs['system.logo_url']) publicConfig.logoUrl = configs['system.logo_url']
    if (configs['system.favicon_url']) {
      const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
      if (link) link.href = configs['system.favicon_url']
    }
    document.title = publicConfig.siteName
  } catch {
    // 公开配置不可用时保持内置品牌
  }
}
const router = useRouter()
const userStore = useUserStore()
const loginFormRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(false)
const currentLanguage = ref(locale.value as SupportedLocale)
const currentYear = new Date().getFullYear()
const loginForm = reactive({ ...DEFAULT_COMMUNITY_LOGIN })
const loginRules: FormRules = {
  username: [
    { required: true, message: () => t('login.error.usernameRequired'), trigger: 'blur' },
    { min: 3, message: () => t('login.error.usernameMinLength'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: () => t('login.error.passwordRequired'), trigger: 'blur' },
    { min: 6, message: () => t('login.error.passwordMinLength'), trigger: 'blur' }
  ]
}

const handleLanguageChange = (lang: SupportedLocale) => {
  changeLanguage(lang)
  currentLanguage.value = lang
}

const loadRememberedUsername = () => {
  const saved = localStorage.getItem(REMEMBER_KEY)
  if (!saved) return
  try {
    const value = JSON.parse(saved) as { username?: string; password?: string }
    if (value.username) {
      loginForm.username = value.username
      loginForm.password = ''
      rememberMe.value = true
      localStorage.setItem(REMEMBER_KEY, JSON.stringify({ username: value.username }))
    } else {
      localStorage.removeItem(REMEMBER_KEY)
    }
  } catch {
    localStorage.removeItem(REMEMBER_KEY)
  }
}

const handleLogin = async () => {
  if (!loginFormRef.value || loading.value) return
  const valid = await loginFormRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const loginData: LoginRequest = { username: loginForm.username, password: loginForm.password }
    const response = await Login(loginData)
    if (response.code !== '000000' || !response.data?.token || !response.data?.user_info) {
      ElMessage.error(response.message || t('login.error.general'))
      return
    }
    userStore.setLoginData(response.data.token, {
      username: response.data.user_info.username,
      role: response.data.user_info.role
    })
    if (rememberMe.value) localStorage.setItem(REMEMBER_KEY, JSON.stringify({ username: loginForm.username }))
    else localStorage.removeItem(REMEMBER_KEY)
    ElMessage.success(response.message || t('login.success'))
    await router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error?.backendMessage || error?.response?.data?.message || error?.message || t('login.error.general'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
	loadPublicConfig()
	loadRememberedUsername()
})
</script>

<style lang="scss" scoped>
.login-page { min-height:100vh; display:grid; grid-template-columns:minmax(0,1.12fr) minmax(420px,.88fr); background:#f8fbfa; }
.brand-panel { position:relative; overflow:hidden; padding:48px clamp(40px,6vw,96px); display:flex; flex-direction:column; justify-content:space-between; color:#fff; background:linear-gradient(135deg,#07584b 0%,#087765 100%); isolation:isolate; }
.brand-panel::before { content:''; position:absolute; inset:0; z-index:-2; opacity:.42; background:linear-gradient(rgba(82,224,179,.08) 1px,transparent 1px),linear-gradient(90deg,rgba(82,224,179,.08) 1px,transparent 1px); background-size:64px 64px; mask-image:linear-gradient(90deg,transparent 5%,#000 45%,#000); }
.brand-panel::after { content:''; position:absolute; width:70%; aspect-ratio:1; right:-26%; top:8%; z-index:-2; background:radial-gradient(circle,rgba(1,156,124,.5),transparent 68%); }
.brand-panel>*:not(.license-motion) { position:relative; z-index:2; }
.brand-mark { display:flex; align-items:center; gap:12px; font-family:'Swis721 BlkCn BT',sans-serif; font-size:28px; }
.brand-mark svg { width:40px; height:40px; }
.brand-copy { max-width:620px; }
.brand-copy .eyebrow { margin:0 0 14px; color:#7ce9c5; font-size:13px; font-weight:700; letter-spacing:.14em; }
.brand-copy h1 { margin:0; width:max-content; max-width:100%; font-size:clamp(38px,3.6vw,58px); line-height:1.08; letter-spacing:-.035em; white-space:nowrap; }
.brand-copy>p { max-width:580px; margin:24px 0 0; font-size:17px; line-height:1.8; opacity:.78; }
.brand-note { margin:0; font-size:13px; opacity:.58; }
.brand-footer { display:flex; align-items:center; justify-content:space-between; gap:20px; }
.brand-footer>a { color:rgba(226,255,245,.72); font-size:13px; text-decoration:none; }
.brand-footer>a span { color:#8debcf; font-weight:600; }
.brand-footer>a:hover span { color:#fff; }
.authorization-flow { margin-top:40px; display:flex; align-items:center; gap:16px; color:#e2fff5; font-family:ui-monospace,SFMono-Regular,Consolas,monospace; font-size:17px; font-weight:600; letter-spacing:.1em; }
.authorization-flow span { opacity:.72; animation:flow-stage 6s ease-in-out infinite; }
.authorization-flow span:nth-of-type(2) { animation-delay:2s; }
.authorization-flow span:nth-of-type(3) { animation-delay:4s; }
.authorization-flow i { position:relative; overflow:hidden; width:52px; height:2px; background:rgba(181,255,230,.52); }
.authorization-flow i::after { content:''; position:absolute; inset:0; background:linear-gradient(90deg,transparent,#7eebc7,transparent); transform:translateX(-100%); animation:flow-line 6s ease-in-out infinite; }
.authorization-flow i:nth-of-type(2)::after { animation-delay:2s; }
.license-motion { position:absolute; width:min(46vw,620px); aspect-ratio:1; right:-16%; top:50%; z-index:0; opacity:.72; transform:translateY(-50%); pointer-events:none; }
.motion-orbit { position:absolute; inset:12%; border:1px solid rgba(82,224,179,.24); transform:rotate(45deg); animation:orbit-rotate 22s linear infinite; }
.motion-orbit--inner { inset:29%; border-color:rgba(124,233,197,.38); animation-duration:14s; animation-direction:reverse; }
.motion-core { position:absolute; inset:43%; display:grid; grid-template-columns:repeat(3,1fr); gap:3px; transform:rotate(45deg); }
.motion-core span { background:#52e0b3; animation:core-pulse 2.4s ease-in-out infinite; }
.motion-core span:nth-child(2) { animation-delay:.3s; }
.motion-core span:nth-child(3) { animation-delay:.6s; }
.motion-signal { position:absolute; width:7px; height:7px; background:#9af3d6; box-shadow:0 0 18px #52e0b3; animation:signal-travel 6s ease-in-out infinite; }
.motion-signal--one { left:17%; top:50%; }
.motion-signal--two { left:50%; top:17%; animation-delay:-2s; }
.motion-signal--three { right:17%; bottom:17%; animation-delay:-4s; }
.login-panel { position:relative; padding:64px 32px 32px; display:flex; align-items:center; justify-content:center; background:#f8fbfa; }
.language-switcher { position:absolute; top:24px; right:28px; width:132px; }
.language-switcher>svg { position:absolute; top:50%; left:12px; z-index:2; width:17px; height:17px; color:#647870; transform:translateY(-50%); pointer-events:none; }
.language-switcher>svg :is(circle,path) { stroke:currentColor; stroke-width:1.5; stroke-linecap:round; stroke-linejoin:round; }
.language-switcher :deep(.el-select__wrapper) { min-height:38px; padding-left:38px; border:1px solid #cddcd7; border-radius: $border-radius-base; background:#fff; box-shadow:none; transition:border-color .2s ease,background-color .2s ease; }
.language-switcher :deep(.el-select__wrapper:hover) { border-color:#82a69a; background:#f4faf7; }
.language-switcher :deep(.el-select__wrapper.is-focused) { border-color:#019c7c; box-shadow:0 0 0 1px #019c7c; }
.language-switcher:focus-within>svg { color:#019c7c; }
.login-card { width:min(100%,400px); }
.login-card header { margin-bottom:36px; }
.login-card h2 { margin:0; color:#17231f; font-size:32px; line-height:1.25; letter-spacing:-.02em; }
.login-form :deep(.el-form-item) { position:relative; margin-bottom:30px; }
.login-form :deep(.el-form-item__label) { position:absolute; top:-8px; left:12px; z-index:2; width:auto; height:18px; margin:0; padding:0 6px; color:#52665f; background:#f8fbfa; font-size:12px; line-height:18px; }
.login-form :deep(.el-form-item__label::before) { display:none; }
.login-form :deep(.el-input__wrapper) { min-height:54px; border:1px solid #cddcd7; border-radius: $border-radius-base; background:#fff; box-shadow:none; }
.login-form :deep(.el-input__wrapper:hover) { border-color:#8eaaa1; }
.login-form :deep(.el-input__wrapper.is-focus) { border-color:#019c7c; box-shadow:0 0 0 1px #019c7c; }
.login-form :deep(input:-webkit-autofill),
.login-form :deep(input:-webkit-autofill:hover),
.login-form :deep(input:-webkit-autofill:focus) { -webkit-text-fill-color:#17231f; -webkit-box-shadow:0 0 0 1000px #fff inset; caret-color:#17231f; }
.form-options { margin:-8px 0 24px; display:flex; justify-content:space-between; }
/* 控件尺寸对齐企业版登录页：48px 高、8px 圆角。
   企业版的 4px 基准令牌不适用于此处 —— 其登录页用自定义类覆盖为 8px，
   这里显式覆盖而非改令牌，避免影响业务页的控件圆角。 */
.login-form :deep(.el-input__wrapper) {
  height: 48px;
  min-height: 48px;
  border-radius: 8px;
}
.login-button {
  width: 100%;
  height: 48px;
  min-height: 48px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
}
.login-card footer { margin-top:40px; color:#708079; font-size:12px; text-align:left; }
@keyframes orbit-rotate { to { transform:rotate(405deg); } }
@keyframes core-pulse { 0%,100% { opacity:.28; transform:scaleY(.55); } 50% { opacity:1; transform:scaleY(1); } }
@keyframes signal-travel { 0%,100% { transform:translate(0,0); opacity:.3; } 50% { transform:translate(92px,-92px); opacity:1; } }
@keyframes flow-stage { 0%,32% { color:#fff; opacity:1; text-shadow:0 0 18px rgba(126,235,199,.9); transform:translateY(-1px); } 40%,100% { color:#e2fff5; opacity:.72; text-shadow:none; transform:translateY(0); } }
@keyframes flow-line { 0%,18% { transform:translateX(-100%); opacity:0; } 25% { opacity:1; } 34%,100% { transform:translateX(100%); opacity:0; } }
@media (max-width:1023px) {
  .login-page { grid-template-columns:1fr; }
  .brand-panel { display:none; }
  .login-panel { min-height:100vh; padding:72px 24px 32px; }
}
@media (max-width:480px) {
  .login-panel { padding-inline:18px; }
  .language-switcher { right:18px; }
  .login-card h2 { font-size:26px; }
}
@media (prefers-reduced-motion:reduce) {
  .motion-orbit,.motion-core span,.motion-signal,.authorization-flow span,.authorization-flow i::after { animation:none; }
}
</style>
