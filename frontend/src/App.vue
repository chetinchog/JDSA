<script setup>
import { reactive, ref } from 'vue'
import { ScrapeJob, ExportJSON } from '../wailsjs/go/backend/App'

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
  <main class="h-screen p-6 max-w-4xl mx-auto flex flex-col gap-6 overflow-hidden bg-slate-50/30">
    <!-- Header -->
    <header class="flex items-center gap-4 shrink-0">
      <div class="h-12 w-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-200">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <div>
        <h1 class="text-2xl font-bold text-slate-800 leading-tight">JDSA Assistant</h1>
        <p class="text-sm text-slate-500">Extractor de Descripciones de Empleo</p>
      </div>
    </header>

    <!-- Scraper Input Card -->
    <section class="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 flex flex-col gap-4 shrink-0">
      <div class="flex flex-col gap-2">
        <label for="job-url" class="text-xs font-bold uppercase tracking-wider text-slate-400">URL del Empleo</label>
        <div class="flex gap-3">
          <input 
            id="job-url"
            v-model="state.jobUrl"
            type="text" 
            placeholder="Pegá el link del empleo acá..."
            class="flex-1 px-4 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 text-sm"
            @keyup.enter="doScrap"
          />
          <button 
            @click="doScrap"
            :disabled="state.loading"
            class="px-6 py-2.5 bg-indigo-600 text-white font-semibold rounded-xl hover:bg-indigo-700 active:scale-95 disabled:opacity-50 disabled:active:scale-100 transition-all shadow-md shadow-indigo-100 flex items-center gap-2 text-sm"
          >
            <span v-if="state.loading" class="animate-spin h-3.5 w-3.5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ state.loading ? 'Scrapeando...' : 'Scrapear' }}
          </button>
        </div>
        <p v-if="state.error" class="text-xs text-red-500 mt-1">{{ state.error }}</p>
      </div>
    </section>

    <!-- Result Display -->
    <section v-if="state.result" class="flex-1 min-h-0 bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden flex flex-col animate-in fade-in slide-in-from-bottom-4 duration-500">
      <!-- Title and Export -->
      <div class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50 shrink-0">
        <h2 class="text-lg font-bold text-slate-800 truncate pr-4">{{ state.result.job_title }}</h2>
        <button 
          @click="doExport"
          class="shrink-0 flex items-center gap-2 px-4 py-2 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 font-semibold rounded-lg transition-colors border border-emerald-100 text-sm"
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
          <div class="p-3 bg-slate-50 rounded-lg border border-slate-100/50">
            <p class="text-[10px] font-bold uppercase text-slate-400 mb-0.5">Empresa</p>
            <p class="text-sm text-slate-700 font-semibold truncate leading-tight">{{ state.result.company_name || 'No disponible' }}</p>
          </div>
          <div class="p-3 bg-slate-50 rounded-lg border border-slate-100/50">
            <p class="text-[10px] font-bold uppercase text-slate-400 mb-0.5">Ubicación</p>
            <p class="text-sm text-slate-700 font-semibold truncate leading-tight">{{ state.result.location || 'No disponible' }}</p>
          </div>
          <div class="p-3 bg-slate-50 rounded-lg border border-slate-100/50">
            <p class="text-[10px] font-bold uppercase text-slate-400 mb-0.5">ID del Empleo</p>
            <p class="text-xs text-slate-500 font-mono truncate">{{ state.result.job_id || 'No disponible' }}</p>
          </div>
          <div class="p-3 bg-slate-50 rounded-lg border border-slate-100/50">
            <p class="text-[10px] font-bold uppercase text-slate-400 mb-0.5">URL de Postulación</p>
            <a :href="state.result.apply_url" target="_blank" class="text-xs text-indigo-600 truncate block hover:underline leading-tight">{{ state.result.apply_url }}</a>
          </div>
        </div>

        <!-- Job Description Scrollable Area -->
        <div class="flex-1 flex flex-col min-h-0">
          <h3 class="text-[10px] font-bold uppercase tracking-wider text-slate-400 mb-3 shrink-0">Descripción del Empleo</h3>
          <div class="flex-1 overflow-y-auto pr-3 custom-scrollbar text-slate-700 leading-relaxed whitespace-pre-wrap text-sm border-t border-slate-50 pt-4">
            {{ state.result.job_description }}
          </div>
        </div>
      </div>
    </section>

    <!-- Empty State -->
    <section v-else-if="!state.loading" class="flex-1 flex flex-col items-center justify-center text-center opacity-40">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-slate-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p class="text-slate-500">Ingresá una URL para empezar a scrapear los datos</p>
    </section>

    <!-- Footer Credits -->
    <footer class="fixed bottom-4 right-6 flex flex-col items-end opacity-60 hover:opacity-100 transition-opacity">
      <p class="text-[9px] text-slate-400 font-bold uppercase tracking-widest leading-tight">
        By <span class="text-indigo-500">iCTG</span>
      </p>
      <p class="text-[9px] text-slate-400 font-medium tracking-tight leading-tight">
        Powered by <span class="text-slate-500">Antigravity</span>
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
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
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
