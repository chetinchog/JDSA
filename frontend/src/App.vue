<script setup>
import { reactive, ref, onMounted, watch } from 'vue'
import { ScrapeJob, ExportCSV, BulkScrape, ExportBulkCSV, OpenURL, SaveConfig, GetConfig, CheckConfig, GetClipboardText, CancelSearch, CancelExport, ExportDebugHTML, GetDebugHTML } from '../wailsjs/go/backend/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import SplashScreen from './components/SplashScreen.vue'
import ScrapingLoader from './components/ScrapingLoader.vue'
import LoginScreen from './components/LoginScreen.vue'
import WaitingScreen from './components/WaitingScreen.vue'
import ToastNotification from './components/ToastNotification.vue'
import HtmlViewerModal from './components/HtmlViewerModal.vue'
import { auth, onUserSnapshot, updateUserCookie, updateUserConfig } from './firebase'
import { onAuthStateChanged, signOut } from 'firebase/auth'

const isDark = ref(false)
const showSplash = ref(true)

const state = reactive({
  mode: 'bulk', // 'single' | 'bulk'
  jobUrl: '',
  bulkQuery: '',
  bulkPlatform: 'indeed',
  loading: false,
  error: '',
  result: null, // Scraped job details
  bulkResults: [], // List of jobs from search
  isPlatformDropdownOpen: false,
  isExporting: false,
  exportFinished: false,
  isCancelingExport: false,
  exportSummary: {
    success: 0,
    errors: 0,
    details: []
  },
  exportProgress: 0,
  exportCurrent: 0,
  exportTotal: 0,
  // Scraping progress
  scrapingCurrent: 0,
  scrapingTotal: 0,
  scrapingFound: 0,
  scrapingPage: 0,
  // Pagination (no longer used for "load more" but kept for offset tracking)
  nextOffset: 0,
  hasMore: false,
  isSearching: false,
  isBlockedByLogin: false,
  manualConfig: {
    session_cookie: '',
    wait_pages_min: 10,
    wait_pages_max: 15,
    wait_job_min: 10,
    wait_job_max: 15
  },
  showCookieInput: false,
  hasSavedCookie: false,
  // Live export counters
  exportSuccess: 0,
  exportErrors: 0,
  // Auth state
  authStatus: 'loading', // 'loading', 'unauthenticated', 'waiting', 'authenticated'
  user: null,
  userData: null,
  lastCheckedCookie: '',
  toast: {
    show: false,
    message: '',
    type: 'error',
    actions: []
  },
  htmlModal: {
    show: false,
    htmlContent: '',
    targetUrl: ''
  }
})

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

const onSplashFinish = () => {
  showSplash.value = false
}

const checkCookies = async () => {
    const hasConfig = await CheckConfig(state.bulkPlatform)
    state.hasSavedCookie = hasConfig
    // If we have a config, we don't show the input unless explicitly opened
    if (hasConfig) {
        state.showCookieInput = false
    } else if (state.bulkPlatform === 'indeed') {
        state.showCookieInput = true
    }
}

const checkAuth = () => {
  // Timeout de seguridad: Si en 5 segundos Firebase no responde, forzamos unauthenticated para no quedar trabados en "loading"
  const authTimeout = setTimeout(() => {
    if (state.authStatus === 'loading') {
      console.warn("Auth timeout reached. Defaulting to unauthenticated.")
      state.authStatus = 'unauthenticated'
    }
  }, 5000)

  onAuthStateChanged(auth, (user) => {
    clearTimeout(authTimeout)
    if (user) {
      state.user = user
      onUserSnapshot(user.uid, async (data) => {
        state.userData = data
        if (data && data.is_enabled) {
          state.authStatus = 'authenticated'
          
          // Apply stored configs for indeed if they changed
          const firebaseConfig = data.config_indeed || null
          if (firebaseConfig) {
             console.log("Syncing Indeed config from Firebase to Backend...")
             await SaveConfig('indeed', firebaseConfig)
             state.lastCheckedCookie = firebaseConfig.session_cookie
             state.manualConfig = { ...firebaseConfig }
             await checkCookies()
          } else if (data.cookies_indeed) {
             // Migration
             const migratedConfig = { session_cookie: data.cookies_indeed, wait_pages_min: 1, wait_pages_max: 3, wait_job_min: 1, wait_job_max: 2 }
             console.log("Migrating older Indeed cookie from Firebase to new Config object...")
             await SaveConfig('indeed', migratedConfig)
             await updateUserConfig(user.uid, 'indeed', migratedConfig)
             state.manualConfig = { ...migratedConfig }
             await checkCookies()
          } else {
             console.log("No Indeed config found in Firebase.")
             state.hasSavedCookie = false
             state.showCookieInput = true
          }
        } else {
          state.authStatus = 'waiting'
        }
      }, (err) => {
        console.error("User snapshot error:", err)
        state.authStatus = 'waiting'
      })
    } else {
      state.authStatus = 'unauthenticated'
      state.user = null
      state.userData = null
      state.lastCheckedCookie = ''
    }
  }, (err) => {
    clearTimeout(authTimeout)
    console.error("Auth state error:", err)
    state.authStatus = 'unauthenticated'
  })
}

const handleLoggedIn = () => {
    // Rely on snapshot listener to change state
}

const handleWaitingApproval = () => {
    // Rely on snapshot listener to change state
}

const doLogout = async () => {
    try {
        await signOut(auth);
    } catch (err) {
        console.error("Logout error", err);
    }
}

const handleExportComplete = (res) => {
  console.log('Finalizing UI with results:', res)
  state.exportFinished = true
  if (res) {
    state.exportSummary = {
      success: res.successCount || 0,
      errors: res.errorCount || 0,
      details: res.errors || [],
      filePath: res.filePath || ''
    }

  }
}

const setMode = (mode) => {
  state.mode = mode
  state.error = ''
}

const showToast = (message, type = 'error', actions = []) => {
  state.toast.message = message
  state.toast.type = type
  state.toast.actions = actions
  state.toast.show = true
}

const handleExportDebugHTML = async () => {
  try {
    const savedPath = await ExportDebugHTML(state.bulkPlatform)
    if (savedPath) {
      showToast(`HTML guardado exitosamente en: ${savedPath}`, 'success')
    }
  } catch (err) {
    showToast(`Error al guardar HTML de diagnóstico: ${err}`, 'error')
  }
}

const handleViewResponseHTML = async () => {
  try {
    const res = await GetDebugHTML(state.bulkPlatform)
    if (res && res.html) {
      state.htmlModal.htmlContent = res.html
      state.htmlModal.targetUrl = res.url || ''
      state.htmlModal.show = true
    }
  } catch (err) {
    showToast(`Error al cargar HTML de respuesta: ${err}`, 'error')
  }
}

watch(() => state.error, (newVal) => {
  if (newVal && !state.toast.show) {
    showToast(newVal, 'error')
  }
})

const doBulkScrap = async (isNew = true) => {
  if (!state.bulkQuery) {
    state.error = 'Por favor, ingresá una búsqueda.'
    return
  }
  
  state.loading = true
  state.isSearching = true
  state.isBlockedByLogin = false
  state.error = ''
  
  if (isNew) {
    state.bulkResults = []
    state.result = null
    state.nextOffset = 0
    state.hasMore = false
  }
  
  state.scrapingCurrent = 0
  state.scrapingFound = 0

  try {
    const res = await BulkScrape(state.bulkQuery, state.bulkPlatform, state.nextOffset)
    if (res.results && res.results.length > 0) {
        state.bulkResults = [...state.bulkResults, ...res.results]
    }
    state.hasMore = res.hasMore
    state.nextOffset = res.nextOffset
    state.isBlockedByLogin = res.isBlockedByLogin
    
    if (state.bulkResults.length === 0) {
        const failedURL = res.failedURL || `https://ar.indeed.com/jobs?q=${encodeURIComponent(state.bulkQuery)}&l=&from=searchOnDesktopSerp`
        let errorMsg = res.blockedReason || 'No se encontraron empleos para esa búsqueda.'
        if (state.isBlockedByLogin && !res.blockedReason) {
            errorMsg = 'La búsqueda fue bloqueada por Indeed (requiere verificar acceso / CAPTCHA).'
        }
        state.error = errorMsg

        // Acciones interactivas para la notificación Toast
        const toastActions = []
        if (failedURL) {
            toastActions.push({
                label: 'Abrir en Navegador',
                primary: true,
                onClick: () => OpenURL(failedURL)
            })
        }
        if (res.hasDebugHTML) {
            toastActions.push({
                label: 'Ver Respuesta',
                primary: false,
                onClick: handleViewResponseHTML
            })
            toastActions.push({
                label: 'Guardar HTML',
                primary: false,
                onClick: handleExportDebugHTML
            })
        }
        if (state.isBlockedByLogin || state.bulkPlatform === 'indeed') {
            toastActions.push({
                label: 'Configurar Cookie',
                primary: false,
                onClick: () => { state.showCookieInput = true }
            })
        }

        showToast(errorMsg, 'error', toastActions)
    }
  } catch (err) {
    const errorMsg = 'Error al buscar: ' + err
    state.error = errorMsg
    showToast(errorMsg, 'error', [
      { label: 'Ver Respuesta', primary: false, onClick: handleViewResponseHTML },
      { label: 'Guardar HTML', primary: false, onClick: handleExportDebugHTML }
    ])
  } finally {
    state.loading = false
    state.isSearching = false
  }
}

const handleCancelSearch = async () => {
    try {
        await CancelSearch()
    } catch (err) {
        console.error('Cancel error:', err)
    }
}

const handleCancelExport = async () => {
    state.isCancelingExport = true
    try {
        await CancelExport()
    } catch (err) {
        console.error('Cancel export error:', err)
    }
}

const handleOpenLogin = () => {
    const url = `https://ar.indeed.com/jobs?q=${encodeURIComponent(state.bulkQuery)}&l=&from=searchOnDesktopSerp`
    OpenURL(url)
}

const handleSaveCookie = async () => {
    state.loading = true
    try {
        const platformKey = state.bulkPlatform === 'indeed' ? 'indeed' : state.bulkPlatform
        const configToSave = {
            session_cookie: state.manualConfig.session_cookie,
            wait_pages_min: Number(state.manualConfig.wait_pages_min) || 10,
            wait_pages_max: Number(state.manualConfig.wait_pages_max) || 15,
            wait_job_min: Number(state.manualConfig.wait_job_min) || 10,
            wait_job_max: Number(state.manualConfig.wait_job_max) || 15
        }
        await SaveConfig(platformKey, configToSave)
        if (state.user) {
            console.log("Saving config to Firebase for user:", state.user.uid)
            await updateUserConfig(state.user.uid, platformKey, configToSave)
        }
        state.lastCheckedCookie = configToSave.session_cookie
        state.hasSavedCookie = true
        state.showCookieInput = false
        state.error = ''
    } catch (err) {
        state.error = 'Error al guardar configuración: ' + err
    } finally {
        state.loading = false
    }
}

const handlePasteCookie = async () => {
    try {
        const text = await GetClipboardText()
        if (text) {
            state.manualConfig.session_cookie = text
        }
    } catch (err) {
        console.error('Clipboard error:', err)
    }
}

const selectJob = async (jobId) => {
  const url = `https://ar.indeed.com/viewjob?jk=${jobId}`
  state.jobUrl = url
  state.loading = true
  state.error = ''
  state.result = null

  try {
    const res = await ScrapeJob(url)
    state.result = res
    if (!res.job_title) {
        state.error = 'No se pudo extraer información del empleo.'
    }
  } catch (err) {
    state.error = 'Error al scrapear: ' + err
  } finally {
    state.loading = false
  }
}

const doBulkExport = async () => {
  if (state.bulkResults.length === 0) return
  state.isExporting = true
  state.exportFinished = false
  state.isCancelingExport = false
  state.exportProgress = 0
  try {
    const res = await ExportBulkCSV(state.bulkQuery, state.bulkPlatform, state.bulkResults, '')
    if (!state.exportFinished) {

        handleExportComplete(res)
    }
  } catch (err) {
    state.error = 'Error al exportar: ' + err
    state.isExporting = false
  }
}

const closeExport = () => {
  state.isExporting = false
  state.exportFinished = false
  state.isCancelingExport = false
}

const retryFailedExport = async () => {
  // Extract failed job IDs from error details (format: "- Title (ID: xxxx): error")
  const failedIds = state.exportSummary.details
    .map(d => {
      const match = d.match(/\(ID:\s*([a-f0-9]+)\)/)
      return match ? match[1] : null
    })
    .filter(Boolean)
  
  if (failedIds.length === 0) return

  // Filter bulkResults to only the failed ones
  const failedResults = state.bulkResults.filter(r => failedIds.includes(r.job_id))
  if (failedResults.length === 0) return

  state.exportFinished = false
  state.isCancelingExport = false
  state.exportProgress = 0
  state.exportSuccess = 0
  state.exportErrors = 0
  
  try {
    // Pass the original file path so backend appends to it instead of asking for a new file
    const res = await ExportBulkCSV(state.bulkQuery, state.bulkPlatform, failedResults, state.exportSummary.filePath)
    if (!state.exportFinished) {

        handleExportComplete(res)
    }
  } catch (err) {
    state.error = 'Error al reintentar: ' + err
    state.isExporting = false
  }
}

const doScrap = async () => {
  if (!state.jobUrl) {
    state.error = 'Por favor, ingresá una URL válida.'
    return
  }
  
  state.loading = true
  state.error = ''
  state.result = null

  try {
    const res = await ScrapeJob(state.jobUrl)
    state.result = res
    if (!res.job_title) {
        state.error = 'No se pudo extraer información del empleo. Verificá la URL.'
    }
  } catch (err) {
    state.error = 'Error al scrapear: ' + err
  } finally {
    state.loading = false
  }
}

const doExport = async () => {
  if (!state.result) return
  try {
    await ExportCSV(state.result)
  } catch (err) {
    state.error = 'Error al exportar: ' + err
  }
}

// Watchers
watch(() => state.bulkPlatform, () => {
    checkCookies()
})

// Lifecycle
onMounted(() => {
  const saved = localStorage.getItem('theme')
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }

  EventsOn('export-progress', (data) => {
    state.isExporting = true
    state.exportCurrent = data.current
    state.exportTotal = data.total
    state.exportProgress = data.total > 0 ? (data.current / data.total) * 100 : 0
    state.exportSuccess = data.success || 0
    state.exportErrors = data.errors || 0
  })

  EventsOn('scraping-progress', (data) => {
    state.scrapingCurrent = data.current
    state.scrapingFound = data.found
    state.scrapingPage = data.page || 0
  })

  EventsOn('export-finished', (res) => {
    if (!state.exportFinished) {
      handleExportComplete(res)
    }
  })

  checkCookies()
  checkAuth()
})
</script>

<template>
  <SplashScreen v-if="showSplash" @finish="onSplashFinish" />
  
  <ScrapingLoader 
    v-if="state.isSearching" 
    :current="state.scrapingCurrent" 
    :found="state.scrapingFound"
    :page="state.scrapingPage"
    @cancel="handleCancelSearch"
  />
  
  <!-- Loading Auth State -->
  <div v-else-if="state.authStatus === 'loading'" class="h-screen w-full flex items-center justify-center bg-slate-50 dark:bg-slate-900">
    <span class="animate-spin h-8 w-8 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full"></span>
  </div>

  <LoginScreen v-else-if="state.authStatus === 'unauthenticated'" @logged-in="handleLoggedIn" @waiting-approval="handleWaitingApproval" />
  
  <WaitingScreen v-else-if="state.authStatus === 'waiting'" />

  <main v-else class="h-screen w-full p-6 pb-10 flex flex-col gap-6 overflow-hidden bg-slate-50/30 dark:bg-slate-900 transition-colors duration-300">
    <!-- Header -->
    <header class="flex items-center gap-4 shrink-0">
      <div class="h-12 w-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-200 dark:shadow-indigo-900/40">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <div class="flex-1 flex items-center justify-between">
        <div class="flex flex-col">
          <h1 class="text-2xl font-bold text-slate-800 dark:text-slate-100 leading-tight">JDS Assistant</h1>
          <p class="text-sm text-slate-500 dark:text-slate-400">Extractor de Descripciones de Empleo</p>
        </div>
        <!-- Mode Switcher moved here -->
        <div class="flex gap-2 p-1 bg-slate-100 dark:bg-slate-800 rounded-xl w-fit shrink-0">
          <button 
            @click="setMode('bulk')"
            :class="state.mode === 'bulk' ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
            class="px-4 py-1.5 rounded-lg text-xs font-bold transition-all"
          >
            Masivo
          </button>
          <button 
            @click="setMode('single')"
            :class="state.mode === 'single' ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
            class="px-4 py-1.5 rounded-lg text-xs font-bold transition-all"
          >
            Individual
          </button>
        </div>
      </div>
      <!-- Theme Toggle -->
      <button 
        @click="toggleTheme"
        class="h-9 w-9 flex items-center justify-center rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors shadow-sm"
        :title="isDark ? 'Modo claro' : 'Modo oscuro'"
      >
        <!-- Sun (shown in dark mode) -->
        <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5 text-amber-400" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z"/>
        </svg>
        <!-- Moon (shown in light mode) -->
        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5 text-slate-500" viewBox="0 0 24 24" fill="currentColor">
          <path fill-rule="evenodd" d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z" clip-rule="evenodd"/>
        </svg>
      </button>

      <!-- Logout -->
      <button 
        @click="doLogout"
        class="h-9 px-3 flex items-center justify-center rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/30 dark:hover:text-red-400 text-slate-500 transition-colors shadow-sm text-xs font-bold"
        title="Cerrar sesión"
      >
        Salir
      </button>
    </header>
    
    <!-- Scraper Input Card -->
    <!-- Scraper Input Card (Single) -->
    <section v-if="state.mode === 'single'" class="bg-white dark:bg-slate-800 rounded-2xl p-5 shadow-sm border border-slate-100 dark:border-slate-700 flex flex-col gap-4 shrink-0 transition-colors">
      <div class="flex flex-col gap-2">
        <label for="job-url" class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">URL del Empleo</label>
        <div class="flex gap-3">
          <input 
            id="job-url"
            v-model="state.jobUrl"
            type="text" 
            placeholder="Pegá el link del empleo acá..."
            class="flex-1 px-4 py-2.5 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 text-sm"
            @keyup.enter="doScrap"
          />
          <button 
            @click="doScrap"
            :disabled="state.loading"
            class="px-6 py-2.5 bg-indigo-600 text-white font-semibold rounded-xl hover:bg-indigo-700 active:scale-95 disabled:opacity-50 disabled:active:scale-100 transition-all shadow-md shadow-indigo-100 dark:shadow-indigo-900/30 flex items-center gap-2 text-sm"
          >
            <span v-if="state.loading" class="animate-spin h-3.5 w-3.5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ state.loading ? 'Scrapeando...' : 'Scrapear' }}
          </button>
        </div>
        <p v-if="state.error" class="text-xs text-red-500 dark:text-red-400 mt-1">{{ state.error }}</p>
      </div>
    </section>

    <!-- Scraper Input Card (Bulk) -->
    <section v-else class="bg-white dark:bg-slate-800 rounded-2xl p-5 shadow-sm border border-slate-100 dark:border-slate-700 flex flex-col gap-4 shrink-0 transition-colors">
      <div class="flex flex-col gap-2">
        <label for="bulk-query" class="text-xs font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">Búsqueda de Empleos</label>
        <div class="flex gap-3 relative">
          <!-- Backdrop for dropdown -->
          <div 
            v-if="state.isPlatformDropdownOpen" 
            class="fixed inset-0 z-40" 
            @click="state.isPlatformDropdownOpen = false"
          ></div>
          
          <div class="relative z-50">
            <button 
              @click="state.isPlatformDropdownOpen = !state.isPlatformDropdownOpen"
              class="w-[120px] h-full px-4 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-600/50 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all text-sm flex items-center justify-between"
              :class="{ 'ring-2 ring-indigo-500 border-transparent': state.isPlatformDropdownOpen }"
            >
              <span class="truncate pr-2 font-medium">{{ state.bulkPlatform === 'indeed' ? 'Indeed' : state.bulkPlatform }}</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-500 dark:text-slate-400 shrink-0 transition-transform duration-200" :class="{ 'rotate-180': state.isPlatformDropdownOpen }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            
            <div 
              v-if="state.isPlatformDropdownOpen"
              class="absolute top-[calc(100%+0.5rem)] left-0 w-full bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-xl shadow-xl shadow-slate-200/50 dark:shadow-black/20 overflow-hidden flex flex-col animate-in fade-in slide-in-from-top-2 duration-200"
            >
              <button 
                @click="state.bulkPlatform = 'indeed'; state.isPlatformDropdownOpen = false; checkCookies()"
                class="w-full text-left px-4 py-3 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors"
                :class="{ 'font-semibold bg-indigo-50/50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300': state.bulkPlatform === 'indeed' }"
              >
                Indeed
              </button>
            </div>
          </div>
          <input 
            id="bulk-query"
            v-model="state.bulkQuery"
            type="text" 
            placeholder="Ej: Data Analyst, Software Engineer..."
            class="flex-1 px-4 py-2.5 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 text-sm"
            @keyup.enter="doBulkScrap"
          />

          <!-- Compact Session Badge (Inline) -->
          <button 
            v-if="state.hasSavedCookie && state.bulkPlatform === 'indeed'" 
            @click="state.showCookieInput = !state.showCookieInput"
            class="flex items-center gap-2 px-3 h-[42px] bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-600 transition-all shrink-0 active:scale-95 group self-center"
            :class="{ 'ring-2 ring-indigo-500 border-transparent bg-indigo-50 dark:bg-indigo-900/20': state.showCookieInput }"
            title="Ver/Cambiar sesión activa"
          >
            <div class="h-1.5 w-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" :class="{ 'animate-pulse': !state.showCookieInput }"></div>
            <span class="text-[10px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest hidden sm:inline">Sesión</span>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-200 transition-transform duration-200" :class="{ 'rotate-180': state.showCookieInput }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          <button 
            @click="doBulkScrap(true)"
            :disabled="state.loading"
            class="px-6 py-2.5 bg-indigo-600 text-white font-semibold rounded-xl hover:bg-indigo-700 active:scale-95 disabled:opacity-50 disabled:active:scale-100 transition-all shadow-md shadow-indigo-100 dark:shadow-indigo-900/30 flex items-center gap-2 text-sm"
          >
            <span v-if="state.loading" class="animate-spin h-3.5 w-3.5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ state.loading ? 'Buscando...' : 'Buscar' }}
          </button>
        </div>
        
        <p v-if="state.error" class="text-xs text-red-500 dark:text-red-400 mt-1">{{ state.error }}</p>

        <!-- Login Alert -->
        <div v-if="state.isBlockedByLogin" class="mt-2 flex flex-col gap-2">
          <div class="p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800 rounded-xl flex items-center justify-between gap-4 animate-in fade-in slide-in-from-top-2 duration-300">
            <div class="flex items-center gap-3">
              <div class="h-8 w-8 rounded-full bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-amber-600 dark:text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4M10 17l5-5-5-5M13.8 12H3"/>
                </svg>
              </div>
              <div>
                <p class="text-[11px] text-amber-800 dark:text-amber-300 leading-tight">
                  Indeed bloqueó el acceso. <span class="font-bold">Iniciá sesión en tu navegador</span> para continuar.
                </p>
                <button 
                  @click="state.showCookieInput = !state.showCookieInput"
                  class="text-[10px] text-amber-600 dark:text-amber-400 underline mt-1 block"
                >
                  {{ state.showCookieInput ? 'Ocultar configuración' : 'Configurar Cookie (Avanzado)' }}
                </button>
              </div>
            </div>
            <button 
              @click="handleOpenLogin"
              class="px-4 py-1.5 bg-amber-600 hover:bg-amber-700 text-white text-[11px] font-bold rounded-lg transition-colors shadow-sm shrink-0"
            >
              Abrir Indeed
            </button>
          </div>
        </div>

        <!-- Advanced Cookie Input (Refined Section) -->
        <div v-if="state.showCookieInput && state.bulkPlatform === 'indeed'" class="mt-2 p-4 bg-slate-50/50 dark:bg-slate-900/30 border border-slate-200 dark:border-slate-700 rounded-2xl animate-in fade-in slide-in-from-top-1 duration-200">
          <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between">
              <p class="text-[10px] text-slate-500 dark:text-slate-400 font-bold uppercase tracking-wider">Configuración de Sesión (Indeed)</p>
              <button @click="state.showCookieInput = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <div class="flex gap-2">
              <button 
                @click="handlePasteCookie"
                class="px-3 py-2 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 text-[11px] font-bold rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-600 flex items-center gap-2 shrink-0 transition-all shadow-sm active:scale-95"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-indigo-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 0 2 2h2a2 2 0 0 0 2-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2m-6 9h4m-4 4h4"/>
                </svg>
                Pegar
              </button>
              <input 
                v-model="state.manualConfig.session_cookie"
                type="password" 
                placeholder="Session Cookie de Indeed..."
                class="flex-1 px-4 py-2 text-[11px] rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
              />
            </div>
            
            <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mt-1">
              <div class="flex flex-col gap-1">
                <label class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-bold">Espera entre páginas (min seg)</label>
                <input v-model.number="state.manualConfig.wait_pages_min" type="number" min="0" class="px-3 py-1.5 text-[11px] rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 w-full" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-bold">Espera entre páginas (max seg)</label>
                <input v-model.number="state.manualConfig.wait_pages_max" type="number" min="0" class="px-3 py-1.5 text-[11px] rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 w-full" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-bold">Espera entre empleos (min seg)</label>
                <input v-model.number="state.manualConfig.wait_job_min" type="number" min="0" class="px-3 py-1.5 text-[11px] rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 w-full" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-bold">Espera entre empleos (max seg)</label>
                <input v-model.number="state.manualConfig.wait_job_max" type="number" min="0" class="px-3 py-1.5 text-[11px] rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 w-full" />
              </div>
            </div>

            <div class="flex justify-end mt-1">
              <button 
                @click="handleSaveCookie"
                class="px-6 py-2 bg-indigo-600 text-white text-[11px] font-bold rounded-xl hover:bg-indigo-700 shadow-md shadow-indigo-100 dark:shadow-indigo-900/20 active:scale-95 transition-all"
              >
                Guardar Configuración
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Main Content Area -->
    <div class="flex-1 flex gap-6 min-h-0 relative">
      
      <!-- Left Column / Single Area -->
      <div 
        class="flex-1 flex flex-col min-w-0 transition-all duration-300"
        :class="{ 'w-1/2 flex-none': state.mode === 'bulk' && state.bulkResults.length > 0 }"
      >
        <!-- Bulk Results Table -->
        <section v-if="state.mode === 'bulk' && state.bulkResults.length > 0" class="flex-1 min-h-0 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700 overflow-hidden flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-500 transition-colors">
          <div class="p-4 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/50 shrink-0">
            <h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Resultados de Búsqueda ({{ state.bulkResults.length }})</h2>
            <button 
              @click="doBulkExport"
              class="shrink-0 flex items-center gap-2 px-3 py-1.5 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/50 font-semibold rounded-lg transition-colors border border-emerald-100 dark:border-emerald-800 text-xs"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Exportar Lista
            </button>
          </div>
          <div class="flex-1 overflow-auto custom-scrollbar">
            <table class="w-full text-left text-sm border-collapse">
              <thead class="sticky top-0 bg-slate-50 dark:bg-slate-900 z-10">
                <tr class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 border-b border-slate-100 dark:border-slate-800">
                  <th class="px-5 py-3">Título</th>
                  <th class="px-5 py-3">Empresa</th>
                  <th class="px-5 py-3">Ubicación</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50 dark:divide-slate-800">
                <tr 
                  v-for="job in state.bulkResults" 
                  :key="job.job_id"
                  @click="selectJob(job.job_id)"
                  class="hover:bg-indigo-50/30 dark:hover:bg-indigo-900/10 cursor-pointer transition-colors group"
                  :class="state.result?.job_id === job.job_id ? 'bg-indigo-50/50 dark:bg-indigo-900/20' : ''"
                >
                  <td class="px-5 py-3 font-medium text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-400">{{ job.title }}</td>
                  <td class="px-5 py-3 text-slate-500 dark:text-slate-400">{{ job.company }}</td>
                  <td class="px-5 py-3 text-slate-500 dark:text-slate-400 italic text-xs">{{ job.location }}</td>
                </tr>
              </tbody>
            </table>
            
            <!-- Load More button at the end of the table -->
            <div v-if="state.hasMore" class="p-6 flex justify-center border-t border-slate-50 dark:border-slate-800">
              <button 
                @click="doBulkScrap(false)"
                :disabled="state.loading"
                class="group relative flex items-center gap-3 px-8 py-3 bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 font-bold rounded-2xl border-2 border-indigo-100 dark:border-indigo-900/30 hover:border-indigo-500 dark:hover:border-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 active:scale-95 transition-all shadow-sm"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 transition-transform group-hover:rotate-180 duration-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                <span>Cargar más empleos</span>
                <span class="absolute -top-2 -right-2 flex h-5 w-5">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-5 w-5 bg-indigo-500 border-2 border-white dark:border-slate-800 text-[10px] items-center justify-center text-white">+</span>
                </span>
              </button>
            </div>
          </div>
        </section>

        <!-- Empty State (No Results or Initial State) -->
        <section v-else-if="!state.loading && state.mode === 'bulk' && state.bulkResults.length === 0" class="flex-1 flex flex-col items-center justify-center text-center opacity-40 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-slate-300 dark:text-slate-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <p class="text-slate-500 dark:text-slate-400">Ingresá una búsqueda para encontrar empleos</p>
        </section>

        <!-- Result Display (Individual Mode Only) -->
        <section v-if="state.mode === 'single' && state.result" class="flex-1 min-h-0 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700 overflow-hidden flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-500 transition-colors">
          <div class="p-5 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/50 shrink-0">
            <div class="flex items-center gap-2 min-w-0 pr-4 flex-1">
              <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 truncate min-w-0">{{ state.result.job_title || 'Sin título' }}</h2>
              <span v-if="state.result.is_expired" class="shrink-0 px-2.5 py-0.5 text-[10px] font-bold rounded-full bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400 uppercase tracking-wider border border-red-200 dark:border-red-800/50">Expirado</span>
            </div>
            <button 
              @click="doExport"
              class="shrink-0 flex items-center gap-2 px-4 py-2 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/50 font-semibold rounded-lg transition-colors border border-emerald-100 dark:border-emerald-800 text-sm"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Exportar
            </button>
          </div>
          
          <div class="flex-1 flex flex-col p-5 gap-5 min-h-0 overflow-hidden">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 shrink-0">
              <div class="p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">Empresa</p>
                <p class="text-sm text-slate-700 dark:text-slate-200 font-semibold truncate leading-tight">{{ state.result.company_name || 'No disponible' }}</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">Ubicación</p>
                <p class="text-sm text-slate-700 dark:text-slate-200 font-semibold truncate leading-tight">{{ state.result.location || 'No disponible' }}</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">ID del Empleo</p>
                <p class="text-xs text-slate-500 dark:text-slate-400 font-mono truncate">{{ state.result.job_id || 'No disponible' }}</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">URL de Postulación</p>
                <a :href="state.result.apply_url" target="_blank" class="text-xs text-indigo-600 dark:text-indigo-400 truncate block hover:underline leading-tight">{{ state.result.apply_url }}</a>
              </div>
            </div>

            <div class="flex-1 flex flex-col min-h-0">
              <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-3 shrink-0">Descripción del Empleo</h3>
              <div class="flex-1 overflow-y-auto pr-3 custom-scrollbar text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-wrap text-sm border-t border-slate-50 dark:border-slate-700 pt-4">
                {{ state.result.job_description }}
              </div>
            </div>
          </div>
        </section>

        <!-- Empty State (Single Mode) -->
        <section v-else-if="!state.loading && state.mode === 'single'" class="flex-1 flex flex-col items-center justify-center text-center opacity-40 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-slate-300 dark:text-slate-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <p class="text-slate-500 dark:text-slate-400">Ingresá una URL para empezar a scrapear los datos</p>
        </section>
      </div>

      <!-- Right Column / Bulk Result Area -->
      <div 
        v-if="state.mode === 'bulk' && state.result" 
        class="flex-1 flex flex-col min-w-0 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700 overflow-hidden animate-in fade-in slide-in-from-right-4 duration-500 transition-all"
      >
        <div class="p-4 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/50 shrink-0">
          <div class="flex items-center gap-2 min-w-0 pr-4 flex-1">
            <button 
              @click="state.result = null"
              class="mr-1 p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 transition-colors shrink-0"
              title="Cerrar descripción"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
            <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 truncate min-w-0">{{ state.result.job_title || 'Sin título' }}</h2>
            <span v-if="state.result.is_expired" class="shrink-0 px-2 py-0.5 text-[9px] font-bold rounded-full bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400 uppercase tracking-wider border border-red-200 dark:border-red-800/50">Expirado</span>
          </div>
          <button 
            @click="doExport"
            class="shrink-0 flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/50 font-semibold rounded-lg transition-colors border border-emerald-100 dark:border-emerald-800 text-xs"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Exportar
          </button>
        </div>
        
        <div class="flex-1 flex flex-col p-4 gap-4 min-h-0 overflow-hidden">
          <div class="grid grid-cols-1 gap-2 shrink-0">
            <div class="flex gap-2">
              <div class="flex-1 p-2 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[9px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">Empresa</p>
                <p class="text-xs text-slate-700 dark:text-slate-200 font-semibold truncate leading-tight">{{ state.result.company_name || 'No disponible' }}</p>
              </div>
              <div class="flex-1 p-2 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[9px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">Ubicación</p>
                <p class="text-xs text-slate-700 dark:text-slate-200 font-semibold truncate leading-tight">{{ state.result.location || 'No disponible' }}</p>
              </div>
            </div>
            <div class="flex gap-2">
              <div class="w-1/3 p-2 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50">
                <p class="text-[9px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">ID</p>
                <p class="text-[10px] text-slate-500 dark:text-slate-400 font-mono truncate">{{ state.result.job_id || 'No disponible' }}</p>
              </div>
              <div class="flex-1 p-2 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-100/50 dark:border-slate-600/50 min-w-0">
                <p class="text-[9px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-0.5">URL</p>
                <a :href="state.result.apply_url" target="_blank" class="text-[10px] text-indigo-600 dark:text-indigo-400 truncate block hover:underline leading-tight">{{ state.result.apply_url }}</a>
              </div>
            </div>
          </div>

          <div class="flex-1 flex flex-col min-h-0">
            <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-2 shrink-0">Descripción del Empleo</h3>
            <div class="flex-1 overflow-y-auto pr-3 custom-scrollbar text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-wrap text-[13px] border-t border-slate-50 dark:border-slate-700 pt-3">
              {{ state.result.job_description }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Credits -->
    <footer class="fixed bottom-3 right-6 opacity-50 hover:opacity-100 transition-opacity z-10">
      <p class="text-[9px] text-slate-400 dark:text-slate-500 font-medium tracking-wide">
        By <span class="font-bold text-indigo-500 dark:text-indigo-400">iCTG</span>
        <span class="mx-1">-</span>
        Powered by <span class="font-bold text-slate-500 dark:text-slate-400">Antigravity</span>
      </p>
    </footer>

    <!-- Export Progress Overlay -->
    <div v-if="state.isExporting" class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/80 backdrop-blur-sm animate-in fade-in duration-300">
      <div class="bg-white dark:bg-slate-800 p-8 rounded-3xl shadow-2xl max-w-md w-full mx-4 border border-slate-100 dark:border-slate-700 flex flex-col items-center justify-center gap-6 animate-in zoom-in-95 duration-500 delay-100">
        <!-- Animated Icon Container (Only show if NOT finished) -->
        <div v-if="!state.exportFinished" class="relative w-24 h-24 mb-2 flex items-center justify-center">
          <!-- Outer rotating ring -->
          <div class="absolute inset-0 border-4 border-indigo-100 dark:border-indigo-900/50 rounded-full"></div>
          <div class="absolute inset-0 border-4 border-indigo-500 rounded-full border-t-transparent animate-spin" style="animation-duration: 2s;"></div>
          
          <!-- Inner bouncing icon -->
          <div class="animate-bounce mt-1">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
            </svg>
          </div>
        </div>

        <!-- Progress View (Only show if NOT finished) -->
        <template v-if="!state.exportFinished">
          <div class="text-center flex flex-col gap-2">
            <h2 class="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-indigo-400 bg-clip-text text-transparent">
              Exportando empleos...
            </h2>
            <p class="text-slate-500 dark:text-slate-400 text-sm max-w-[280px] mx-auto">
              Exportando los detalles de las Descripciones de Empleo
            </p>
          </div>

          <!-- Progress Bar -->
          <div class="w-full space-y-2 mt-2">
            <div class="flex justify-between text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              <span>Progreso</span>
              <span>{{ state.exportCurrent }} / {{ state.exportTotal }}</span>
            </div>
            <div class="h-3 w-full bg-slate-100 dark:bg-slate-700/50 rounded-full overflow-hidden shadow-inner flex">
              <div 
                class="h-full bg-indigo-500 relative transition-all duration-300 ease-out"
                :style="{ width: `${state.exportProgress}%` }"
              >
                <!-- Animated shimmering effect -->
                <div class="absolute top-0 bottom-0 w-24 bg-gradient-to-r from-transparent via-white/30 to-transparent -translate-x-full animate-[shimmer_1.5s_infinite] skew-x-[-20deg]"></div>
              </div>
            </div>
            <div class="text-center pt-1 font-bold text-lg text-indigo-600 dark:text-indigo-400">
              {{ Math.round(state.exportProgress) }}%
            </div>
          </div>

          <!-- Live counters -->
          <div class="w-full flex gap-3">
            <div class="flex-1 bg-emerald-50 dark:bg-emerald-900/20 p-2 rounded-xl border border-emerald-100 dark:border-emerald-800 text-center">
              <p class="text-[9px] font-bold uppercase text-emerald-500 dark:text-emerald-400">Exitosos</p>
              <p class="text-lg font-bold text-emerald-600 dark:text-emerald-400">{{ state.exportSuccess }}</p>
            </div>
            <div class="flex-1 bg-red-50 dark:bg-red-900/20 p-2 rounded-xl border border-red-100 dark:border-red-800 text-center">
              <p class="text-[9px] font-bold uppercase text-red-400 dark:text-red-400">Errores</p>
              <p class="text-lg font-bold text-red-500 dark:text-red-400">{{ state.exportErrors }}</p>
            </div>
          </div>
          
          <button 
            @click="handleCancelExport"
            :disabled="state.isCancelingExport"
            class="w-full mt-2 py-2.5 font-bold rounded-xl transition-all outline-none border flex justify-center items-center gap-2 text-sm"
            :class="state.isCancelingExport ? 'bg-slate-100 text-slate-400 border-slate-200 dark:bg-slate-800 dark:text-slate-500 dark:border-slate-700 cursor-not-allowed' : 'bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 active:scale-95 border-red-200 dark:border-red-800'"
          >
            <span v-if="state.isCancelingExport" class="animate-spin h-4 w-4 border-2 border-slate-400/30 border-t-slate-400 rounded-full"></span>
            {{ state.isCancelingExport ? 'Deteniendo y Guardando...' : 'Detener y Guardar' }}
          </button>
        </template>

        <!-- Summary View (Shown when finished) -->
        <template v-else>
          <div class="flex flex-col items-center gap-4 w-full">
            <div class="h-16 w-16 bg-emerald-100 dark:bg-emerald-900/30 rounded-full flex items-center justify-center text-emerald-600 dark:text-emerald-400 mb-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            
            <h2 class="text-2xl font-bold text-slate-800 dark:text-slate-100">Exportación Lista</h2>
            
            <div class="w-full grid grid-cols-2 gap-3 mt-2">
              <div class="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-2xl border border-slate-100 dark:border-slate-600 text-center">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-1">Exitosos</p>
                <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ state.exportSummary.success }}</p>
              </div>
              <div class="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-2xl border border-slate-100 dark:border-slate-600 text-center">
                <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-1">Errores</p>
                <p class="text-2xl font-bold text-red-500 dark:text-red-400">{{ state.exportSummary.errors }}</p>
              </div>
            </div>

            <!-- Error list if exists -->
            <div v-if="state.exportSummary.details.length > 0" class="w-full mt-2">
              <p class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 mb-2 px-1">Detalles de errores</p>
              <div class="max-h-32 overflow-y-auto custom-scrollbar bg-slate-50 dark:bg-slate-900/50 rounded-xl p-3 text-[11px] text-slate-500 dark:text-slate-400 border border-slate-100 dark:border-slate-800 italic">
                <div v-for="(err, idx) in state.exportSummary.details" :key="idx" class="mb-1 last:mb-0">
                  {{ err }}
                </div>
              </div>
            </div>

            <button 
              @click="closeExport"
              class="w-full mt-4 py-3 bg-indigo-600 text-white font-bold rounded-xl hover:bg-indigo-700 active:scale-95 transition-all shadow-lg shadow-indigo-100 dark:shadow-indigo-900/30"
            >
              Entendido
            </button>
            <button 
              v-if="state.exportSummary.errors > 0"
              @click="retryFailedExport"
              class="w-full py-2.5 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 font-bold rounded-xl hover:bg-amber-100 dark:hover:bg-amber-900/30 active:scale-95 transition-all border border-amber-200 dark:border-amber-800 text-sm"
            >
              ⟳ Reintentar fallidos ({{ state.exportSummary.errors }})
            </button>
          </div>
        </template>
      </div>
    </div>
  </main>

  <ToastNotification
    :show="state.toast.show"
    @update:show="state.toast.show = $event"
    :message="state.toast.message"
    :type="state.toast.type"
    :actions="state.toast.actions"
  />

  <HtmlViewerModal
    :show="state.htmlModal.show"
    @update:show="state.htmlModal.show = $event"
    :htmlContent="state.htmlModal.htmlContent"
    :targetUrl="state.htmlModal.targetUrl"
    @export="handleExportDebugHTML"
  />
</template>

<style>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background: #475569;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #64748b;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideInUp {
  from { transform: translateY(1rem); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

@keyframes shimmer {
  0% { left: -100px; }
  50% { left: 100%; }
  100% { left: -100px; }
}

.animate-in {
  animation: fadeIn 0.5s ease-out, slideInUp 0.5s ease-out;
}
</style>
