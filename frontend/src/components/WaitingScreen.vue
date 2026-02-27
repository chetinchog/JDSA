<template>
  <div class="h-screen w-full flex flex-col items-center justify-center bg-slate-50 dark:bg-slate-900 transition-colors p-6 overflow-hidden">
    <div class="max-w-xl w-full flex flex-col items-center text-center px-4">
      
      <!-- Calming animated icon -->
      <div class="relative w-32 h-32 mb-10 flex items-center justify-center">
        <!-- Soothing ripples -->
        <div class="absolute inset-0 bg-indigo-200/50 dark:bg-indigo-900/40 rounded-full animate-ping" style="animation-duration: 3s;"></div>
        <div class="absolute inset-4 bg-indigo-300/40 dark:bg-indigo-800/40 rounded-full animate-ping" style="animation-duration: 3s; animation-delay: 1s;"></div>
        
        <div class="relative bg-white dark:bg-slate-800 w-full h-full rounded-full shadow-2xl flex items-center justify-center border-4 border-indigo-50 dark:border-slate-700/50 z-10 transition-transform hover:scale-105 duration-500">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-indigo-500 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
      </div>

      <h1 class="text-4xl md:text-5xl font-extrabold text-slate-800 dark:text-slate-100 mb-6 leading-tight tracking-tight bg-gradient-to-r from-indigo-600 to-indigo-400 bg-clip-text text-transparent">
        ¡Tu cuenta se está creando y se activará lo antes posible!
      </h1>
      
      <div class="h-1 w-16 bg-indigo-500 rounded-full mb-8"></div>
      
      <p class="text-lg md:text-xl text-slate-500 dark:text-slate-400 max-w-lg leading-relaxed font-medium">
        Contáctate con <span class="text-indigo-600 dark:text-indigo-400 font-bold">iCTG</span> si se demora el proceso.
      </p>
      
      <!-- Fun interactive elements at the bottom -->
      <div class="mt-16 flex items-center gap-2 text-sm text-slate-400 dark:text-slate-500">
        <span class="animate-bounce">⏳</span>
        <span>Aguardá pacientemente...</span>
      </div>
    </div>
  </div>
</template>

<script setup>
// This is a static screen that shows up instead of the app
</script>
