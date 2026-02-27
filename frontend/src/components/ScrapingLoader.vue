<script setup>
const props = defineProps({
  current: {
    type: Number,
    default: 0
  },
  found: {
    type: Number,
    default: 0
  },
  page: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['cancel'])
</script>

<template>
  <div class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 backdrop-blur-md animate-in fade-in duration-300">
    <div class="bg-white dark:bg-slate-800 p-10 rounded-[2.5rem] shadow-2xl max-w-sm w-full mx-4 border border-white/20 dark:border-slate-700/50 flex flex-col items-center gap-8 animate-in zoom-in-95 duration-500">
      
      <!-- Premium Animation Container -->
      <div class="relative w-32 h-32 flex items-center justify-center">
        <!-- Pulse effect -->
        <div class="absolute inset-0 bg-indigo-500/20 dark:bg-indigo-500/10 rounded-full animate-ping"></div>
        <div class="absolute inset-4 bg-indigo-500/10 dark:bg-indigo-500/5 rounded-full animate-ping" style="animation-delay: 0.5s"></div>
        
        <!-- Rotating ring (infinite spin) -->
        <svg class="absolute inset-0 w-full h-full animate-spin" style="animation-duration: 3s;">
          <circle 
            cx="64" cy="64" r="60" 
            class="stroke-slate-100 dark:stroke-slate-700 fill-none" 
            stroke-width="8" 
          />
          <circle 
            cx="64" cy="64" r="60" 
            class="stroke-indigo-500 fill-none" 
            stroke-width="8" 
            stroke-linecap="round"
            :stroke-dasharray="2 * Math.PI * 60"
            :stroke-dashoffset="2 * Math.PI * 60 * 0.75"
          />
        </svg>

        <!-- Center Icon -->
        <div class="relative z-10 text-indigo-500 animate-pulse">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10 7v3m0 0v3m0-3h3m-3 0H7" />
          </svg>
        </div>
      </div>

      <div class="text-center space-y-3">
        <h3 class="text-xl font-bold text-slate-800 dark:text-slate-100 italic">JDSA Engine</h3>
        <p class="text-sm font-medium text-slate-500 dark:text-slate-400 min-h-[1.25rem]">
          Escaneando empleos en Indeed...
        </p>
      </div>

      <!-- Stats -->
      <div class="w-full bg-slate-50 dark:bg-slate-900/50 p-5 rounded-3xl border border-slate-100 dark:border-slate-700 flex flex-col items-center">
        <span class="text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500 tracking-widest mb-1">Empleos encontrados</span>
        <span class="text-4xl font-black text-indigo-600 dark:text-indigo-400">{{ found }}</span>
        <span class="text-[11px] text-slate-400 dark:text-slate-500 mt-2">Página {{ page }} revisada</span>
      </div>

      <!-- Cancel Button -->
      <button 
        @click="emit('cancel')"
        class="w-full py-3 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 font-bold rounded-xl hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400 active:scale-95 transition-all border border-slate-200 dark:border-slate-600"
      >
        Cancelar búsqueda
      </button>

      <!-- Decorative dots -->
      <div class="flex gap-2">
        <div v-for="i in 3" :key="i" 
             class="w-1.5 h-1.5 rounded-full bg-indigo-500/30 animate-bounce" 
             :style="{ animationDelay: `${(i-1) * 0.2}s` }">
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes zoom-in-95 {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.animate-in {
  animation: fadeIn 0.4s ease-out;
}

.zoom-in-95 {
  animation: zoom-in-95 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}
</style>
