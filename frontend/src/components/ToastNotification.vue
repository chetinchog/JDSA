<template>
  <div 
    v-if="show"
    @mouseenter="pauseTimer"
    @mouseleave="resumeTimer"
    class="fixed bottom-6 right-6 z-50 flex flex-col overflow-hidden rounded-xl shadow-2xl border w-96 max-w-lg transform transition-all duration-300 animate-in fade-in slide-in-from-bottom-4"
    :class="{
      'bg-red-50 dark:bg-red-900/95 border-red-200 dark:border-red-800': type === 'error',
      'bg-emerald-50 dark:bg-emerald-900/95 border-emerald-200 dark:border-emerald-800': type === 'success',
      'bg-indigo-50 dark:bg-indigo-900/95 border-indigo-200 dark:border-indigo-800': type === 'info'
    }"
  >
    <!-- Contenido del Mensaje -->
    <div class="flex items-start gap-3 px-4 py-4"
      :class="{
        'text-red-800 dark:text-red-100': type === 'error',
        'text-emerald-800 dark:text-emerald-100': type === 'success',
        'text-indigo-800 dark:text-indigo-100': type === 'info'
      }"
    >
      <div class="flex-1 text-sm font-semibold pr-2 leading-relaxed break-words">{{ message }}</div>
      <button @click="close" class="shrink-0 p-1.5 rounded-lg opacity-70 hover:opacity-100 transition-colors" :class="{
        'hover:bg-red-100 dark:hover:bg-red-800': type === 'error',
        'hover:bg-emerald-100 dark:hover:bg-emerald-800': type === 'success',
        'hover:bg-indigo-100 dark:hover:bg-indigo-800': type === 'info'
      }">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Botones de Acción -->
    <div v-if="actions && actions.length > 0" class="flex flex-wrap gap-2 px-4 pb-3">
      <button 
        v-for="(action, index) in actions" 
        :key="index"
        @click="handleAction(action)"
        class="text-xs font-semibold px-3 py-1.5 rounded-lg transition-colors shadow-sm flex items-center gap-1.5"
        :class="action.primary 
          ? 'bg-red-600 hover:bg-red-700 text-white dark:bg-red-700 dark:hover:bg-red-600' 
          : 'bg-white hover:bg-slate-100 text-slate-800 border border-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-100 dark:border-slate-700'"
      >
        <span v-if="action.icon" v-html="action.icon"></span>
        <span>{{ action.label }}</span>
      </button>
    </div>

    <!-- Barra de progreso de tiempo -->
    <div class="h-1 w-full shrink-0"
      :class="{
        'bg-red-200 dark:bg-red-950': type === 'error',
        'bg-emerald-200 dark:bg-emerald-950': type === 'success',
        'bg-indigo-200 dark:bg-indigo-950': type === 'info'
      }"
    >
      <div class="h-full transition-all duration-75 ease-linear" 
        :style="{ width: `${progress}%` }"
        :class="{
          'bg-red-500 dark:bg-red-400': type === 'error',
          'bg-emerald-500 dark:bg-emerald-400': type === 'success',
          'bg-indigo-500 dark:bg-indigo-400': type === 'info'
        }"
      ></div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue'

const props = defineProps({
  show: Boolean,
  message: String,
  type: {
    type: String,
    default: 'error' // 'error' | 'success' | 'info'
  },
  duration: {
    type: Number,
    default: 8000
  },
  actions: {
    type: Array,
    default: () => []
  }
})

const handleAction = (action) => {
  if (typeof action.onClick === 'function') {
    action.onClick()
  }
}

const emit = defineEmits(['update:show'])

const progress = ref(100)
const remainingTime = ref(props.duration)
const lastFrameTime = ref(0)
const isPaused = ref(false)
const animationFrame = ref(null)

const close = () => {
  emit('update:show', false)
  stopLoop()
}

const stopLoop = () => {
  if (animationFrame.value) {
    cancelAnimationFrame(animationFrame.value)
    animationFrame.value = null
  }
}

const tick = (timestamp) => {
  if (!lastFrameTime.value) lastFrameTime.value = timestamp
  
  const delta = timestamp - lastFrameTime.value
  lastFrameTime.value = timestamp

  if (!isPaused.value) {
    remainingTime.value = Math.max(0, remainingTime.value - delta)
    progress.value = (remainingTime.value / props.duration) * 100

    if (remainingTime.value <= 0) {
      close()
      return
    }
  }

  animationFrame.value = requestAnimationFrame(tick)
}

const startLoop = () => {
  stopLoop()
  lastFrameTime.value = 0
  animationFrame.value = requestAnimationFrame(tick)
}

const pauseTimer = () => {
  isPaused.value = true
}

const resumeTimer = () => {
  isPaused.value = false
  lastFrameTime.value = 0 // Evita que salte de golpe restando el tiempo transcurrido en pausa
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    remainingTime.value = props.duration
    progress.value = 100
    isPaused.value = false
    startLoop()
  } else {
    stopLoop()
  }
})

onUnmounted(() => {
  stopLoop()
})
</script>
