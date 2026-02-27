<script setup>
import { ref } from 'vue';
import { loginWithGoogle, getUserData, createUserData } from '../firebase';

const emit = defineEmits(['logged-in', 'waiting-approval']);

const isLoading = ref(false);
const errorMsg = ref('');

const handleLogin = async () => {
  isLoading.value = true;
  errorMsg.value = '';
  
  try {
    const user = await loginWithGoogle();
    
    // Check if user exists in DB
    let userData = await getUserData(user.uid);
    if (!userData) {
      // First time login, create user document
      await createUserData(user);
      userData = await getUserData(user.uid); // Fetch to ensure reactivity works later if needed
    }

    if (userData.is_enabled) {
      emit('logged-in', { user, userData });
    } else {
      emit('waiting-approval', { user, userData });
    }
  } catch (error) {
    if (error.code === 'auth/popup-closed-by-user') {
      errorMsg.value = 'El inicio de sesión fue cancelado.';
    } else {
      errorMsg.value = 'Error al iniciar sesión: ' + error.message;
    }
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="h-screen w-full flex flex-col items-center justify-center bg-slate-50 dark:bg-slate-900 transition-colors p-6">
    <div class="max-w-md w-full bg-white dark:bg-slate-800 rounded-3xl p-8 shadow-xl shadow-slate-200/50 dark:shadow-black/20 border border-slate-100 dark:border-slate-700 flex flex-col items-center animate-in zoom-in-95 duration-500">
      
      <!-- Logo Container -->
      <div class="h-20 w-20 bg-indigo-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-200 dark:shadow-indigo-900/40 mb-6 relative">
		<div class="absolute inset-0 bg-indigo-500/30 rounded-2xl animate-ping opacity-75"></div>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white relative z-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>

      <h1 class="text-3xl font-bold text-slate-800 dark:text-slate-100 mb-2 tracking-tight">Bienvenido a JDSA</h1>
      <p class="text-slate-500 dark:text-slate-400 text-center text-sm mb-8 leading-relaxed">
        Iniciá sesión para comenzar a extraer descripciones de empleo automáticamente y sin límites.
      </p>

      <button 
        @click="handleLogin" 
        :disabled="isLoading"
        class="w-full flex items-center justify-center gap-3 px-6 py-4 bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-600 transition-all shadow-sm active:scale-95 disabled:opacity-50 disabled:active:scale-100 font-semibold"
      >
        <span v-if="isLoading" class="animate-spin h-5 w-5 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full"></span>
        <img v-else src="https://www.gstatic.com/firebasejs/ui/2.0.0/images/auth/google.svg" alt="Google" class="w-6 h-6" />
        {{ isLoading ? 'Conectando...' : 'Continuar con Google' }}
      </button>

      <p v-if="errorMsg" class="mt-4 text-sm text-red-500 dark:text-red-400 font-medium text-center bg-red-50 dark:bg-red-900/20 px-4 py-2 rounded-lg border border-red-100 dark:border-red-800 w-full animate-in fade-in slide-in-from-top-2">
        {{ errorMsg }}
      </p>
	  
	  <p class="mt-8 text-xs text-slate-400 dark:text-slate-500 uppercase tracking-widest font-bold">iCTG Assistant</p>
    </div>
  </div>
</template>
