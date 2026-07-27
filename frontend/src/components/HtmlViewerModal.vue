<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-md animate-in fade-in">
    <div class="relative flex flex-col w-full max-w-5xl h-[88vh] bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-800 overflow-hidden transform transition-all">
      
      <!-- Modal Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center gap-3 overflow-hidden pr-4">
          <div class="p-2 rounded-xl bg-indigo-50 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </div>
          <div class="overflow-hidden">
            <h3 class="text-base font-bold text-slate-800 dark:text-slate-100">Vista Previa de la Respuesta HTML</h3>
            <p v-if="targetUrl" class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-xl" :title="targetUrl">
              URL: <span class="font-mono">{{ targetUrl }}</span>
            </p>
          </div>
        </div>

        <!-- Tab Selector & Close -->
        <div class="flex items-center gap-3 shrink-0">
          <div class="flex p-1 bg-slate-200/80 dark:bg-slate-800 rounded-xl text-xs font-semibold">
            <button 
              @click="activeTab = 'preview'"
              class="px-3 py-1.5 rounded-lg transition-all"
              :class="activeTab === 'preview' 
                ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-300 shadow-sm font-bold' 
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              👁️ Vista Previa
            </button>
            <button 
              @click="activeTab = 'code'"
              class="px-3 py-1.5 rounded-lg transition-all"
              :class="activeTab === 'code' 
                ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-300 shadow-sm font-bold' 
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              📄 Código Fuente
            </button>
          </div>

          <button 
            @click="close" 
            class="p-2 rounded-xl text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Modal Body -->
      <div class="flex-1 p-4 bg-slate-100 dark:bg-slate-950 overflow-hidden">
        <!-- Tab Preview (rendered HTML inside iframe) -->
        <div v-show="activeTab === 'preview'" class="w-full h-full">
          <iframe 
            :srcdoc="htmlContent" 
            class="w-full h-full bg-white rounded-xl border border-slate-200 dark:border-slate-800 shadow-inner"
            title="HTML Preview"
            sandbox="allow-same-origin allow-scripts"
          ></iframe>
        </div>

        <!-- Tab Source Code (raw text) -->
        <div v-show="activeTab === 'code'" class="w-full h-full">
          <textarea 
            readonly 
            :value="htmlContent" 
            class="w-full h-full p-4 font-mono text-xs bg-slate-900 text-slate-100 rounded-xl overflow-auto border border-slate-800 resize-none shadow-inner focus:outline-none custom-scrollbar"
            placeholder="No hay código HTML disponible"
          ></textarea>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="flex items-center justify-between px-6 py-3.5 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <div class="text-xs text-slate-400">
          Tamaño: {{ (htmlContent ? htmlContent.length : 0).toLocaleString() }} caracteres
        </div>

        <div class="flex items-center gap-3">
          <button 
            @click="copyCode" 
            class="px-4 py-2 text-xs font-semibold text-slate-700 dark:text-slate-200 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all shadow-sm flex items-center gap-2"
          >
            <span>{{ copied ? '✓ Copiado!' : '📋 Copiar HTML' }}</span>
          </button>

          <button 
            @click="emitExport" 
            class="px-4 py-2 text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl transition-all shadow-sm flex items-center gap-2"
          >
            <span>💾 Exportar HTML</span>
          </button>

          <button 
            @click="close" 
            class="px-4 py-2 text-xs font-semibold text-slate-600 dark:text-slate-400 hover:bg-slate-200/60 dark:hover:bg-slate-800 rounded-xl transition-all"
          >
            Cerrar
          </button>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  show: Boolean,
  htmlContent: String,
  targetUrl: String
})

const emit = defineEmits(['update:show', 'export'])

const activeTab = ref('preview')
const copied = ref(false)

const close = () => {
  emit('update:show', false)
}

const emitExport = () => {
  emit('export')
}

const copyCode = async () => {
  if (!props.htmlContent) return
  try {
    await navigator.clipboard.writeText(props.htmlContent)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Clipboard copy error:', err)
  }
}
</script>
