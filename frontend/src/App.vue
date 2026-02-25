<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ScrapeJob, ExportJSON } from '../wailsjs/go/backend/App'

const isDark = ref(false)

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
})

const state = reactive({
  jobUrl: '',
  loading: false,
  error: '',
  result: null
})

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
    await ExportJSON(state.result)
  } catch (err) {
    state.error = 'Error al exportar: ' + err
  }
}
</script>

<template>
  <main class="h-screen p-6 pb-10 max-w-4xl mx-auto flex flex-col gap-6 overflow-hidden bg-slate-50/30 dark:bg-slate-900 transition-colors duration-300">
    <!-- Header -->
    <header class="flex items-center gap-4 shrink-0">
      <div class="h-12 w-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-200 dark:shadow-indigo-900/40">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <div class="flex-1">
        <h1 class="text-2xl font-bold text-slate-800 dark:text-slate-100 leading-tight">JDSA Assistant</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400">Extractor de Descripciones de Empleo</p>
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
    </header>

    <!-- Scraper Input Card -->
    <section class="bg-white dark:bg-slate-800 rounded-2xl p-5 shadow-sm border border-slate-100 dark:border-slate-700 flex flex-col gap-4 shrink-0 transition-colors">
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

    <!-- Result Display -->
    <section v-if="state.result" class="flex-1 min-h-0 bg-white dark:bg-slate-800 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700 overflow-hidden flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-500 transition-colors">
      <!-- Title and Export -->
      <div class="p-5 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/50 shrink-0">
        <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 truncate pr-4">{{ state.result.job_title }}</h2>
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
        <!-- Info Grid (Company, Location, ID, URL) -->
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

        <!-- Job Description Scrollable Area -->
        <div class="flex-1 flex flex-col min-h-0">
          <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-3 shrink-0">Descripción del Empleo</h3>
          <div class="flex-1 overflow-y-auto pr-3 custom-scrollbar text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-wrap text-sm border-t border-slate-50 dark:border-slate-700 pt-4">
            {{ state.result.job_description }}
          </div>
        </div>
      </div>
    </section>

    <!-- Empty State -->
    <section v-else-if="!state.loading" class="flex-1 flex flex-col items-center justify-center text-center opacity-40">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-slate-300 dark:text-slate-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p class="text-slate-500 dark:text-slate-400">Ingresá una URL para empezar a scrapear los datos</p>
    </section>

    <!-- Footer Credits -->
    <footer class="fixed bottom-3 right-6 opacity-50 hover:opacity-100 transition-opacity">
      <p class="text-[9px] text-slate-400 dark:text-slate-500 font-medium tracking-wide">
        By <span class="font-bold text-indigo-500 dark:text-indigo-400">iCTG</span>
        <span class="mx-1">-</span>
        Powered by <span class="font-bold text-slate-500 dark:text-slate-400">Antigravity</span>
      </p>
    </footer>
  </main>
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

.animate-in {
  animation: fadeIn 0.5s ease-out, slideInUp 0.5s ease-out;
}
</style>
